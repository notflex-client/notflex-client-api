package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
	"gorm.io/gorm"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

func InitStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}

func CreateStripeCheckout(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "CreateStripeCheckout", "userID", user.ID}

	var body struct {
		PlanID int `json:"plan_id" validate:"required"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	var plan models.SubscriptionPlan
	err = database.DB.WithContext(r.Context()).
		First(&plan, "id = ? AND is_active = TRUE", body.PlanID).Error
	if err != nil {
		HandleResponseError(w, r, NewNotFoundError("SubscriptionPlanNotFound", logParams...))
		return
	}

	clientURL := os.Getenv("CLIENT_APP_URL")
	if clientURL == "" {
		clientURL = "http://localhost:8386"
	}

	priceUSD := int64(plan.Price / 25000)
	if priceUSD < 1 {
		priceUSD = 1
	}

	params := &stripe.CheckoutSessionParams{
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		CustomerEmail: stripe.String(user.Email),
		SuccessURL:    stripe.String(clientURL + "/billing?status=success&session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:     stripe.String(clientURL + "/plans?status=canceled"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(plan.Name),
						Description: stripe.String(fmt.Sprintf("Notflex %s — %d days", plan.Name, plan.DurationDays)),
					},
					UnitAmount: stripe.Int64(priceUSD * 100),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id": user.ID,
			"plan_id": fmt.Sprintf("%d", plan.ID),
		},
	}

	sess, err := session.New(params)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating stripe session", err, logParams...))
		return
	}

	payment := models.Payment{
		UserID:          user.ID,
		PlanID:          plan.ID,
		Amount:          plan.Price,
		PaymentMethod:  "stripe",
		Status:          "pending",
		StripeSessionID: sess.ID,
	}
	err = database.DB.WithContext(r.Context()).Create(&payment).Error
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("creating payment record", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"url": sess.URL, "session_id": sess.ID})
}

func StripeWebhook(w http.ResponseWriter, r *http.Request) {
	logParams := []any{"handler", "StripeWebhook"}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}

	signatureHeader := r.Header.Get("Stripe-Signature")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	event, err := webhook.ConstructEvent(payload, signatureHeader, webhookSecret)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidWebhookSignature", logParams...))
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		err = handleCheckoutCompleted(r, event)
	case "checkout.session.expired", "checkout.session.async_payment_failed":
		err = handleCheckoutFailed(r, event)
	}

	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("processing stripe webhook", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"received": true})
}

func handleCheckoutCompleted(r *http.Request, event stripe.Event) error {
	var sess stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &sess)
	if err != nil {
		return err
	}
	return activateSession(r, sess.ID, sess.PaymentIntent)
}

func handleCheckoutFailed(r *http.Request, event stripe.Event) error {
	var sess stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &sess)
	if err != nil {
		return err
	}
	return database.DB.WithContext(r.Context()).
		Model(&models.Payment{}).
		Where("stripe_session_id = ?", sess.ID).
		Update("status", "failed").Error
}

func activateSession(r *http.Request, sessionID string, paymentIntent *stripe.PaymentIntent) error {
	var payment models.Payment
	err := database.DB.WithContext(r.Context()).
		Where("stripe_session_id = ?", sessionID).
		First(&payment).Error
	if err != nil {
		return err
	}

	if payment.Status == "success" {
		return nil
	}

	var plan models.SubscriptionPlan
	err = database.DB.WithContext(r.Context()).First(&plan, "id = ?", payment.PlanID).Error
	if err != nil {
		return err
	}

	return database.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		err = tx.Model(&models.UserSubscription{}).
			Where("user_id = ? AND status = ?", payment.UserID, "active").
			Update("status", "cancelled").Error
		if err != nil {
			return err
		}

		now := time.Now()
		subscription := models.UserSubscription{
			UserID:    payment.UserID,
			PlanID:    plan.ID,
			StartDate: now,
			EndDate:   now.AddDate(0, 0, plan.DurationDays),
			Status:    "active",
		}
		err = tx.Create(&subscription).Error
		if err != nil {
			return err
		}

		paymentIntentID := ""
		if paymentIntent != nil {
			paymentIntentID = paymentIntent.ID
		}
		err = tx.Model(&payment).Updates(map[string]any{
			"status":                   "success",
			"subscription_id":          subscription.ID,
			"stripe_payment_intent_id": paymentIntentID,
			"transaction_id":           sessionID,
		}).Error
		if err != nil {
			return err
		}

		return tx.Model(&models.User{}).Where("id = ?", payment.UserID).Update("role", "subscriber").Error
	})
}

func VerifyStripeSession(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "VerifyStripeSession", "userID", user.ID}

	var body struct {
		SessionID string `json:"session_id" validate:"required"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		HandleResponseError(w, r, NewBadRequestError("InvalidBodyStructure", logParams...))
		return
	}
	formErrors := helpers.ValidateStruct(body)
	if formErrors != nil {
		HandleResponseError(w, r, NewValidationError(formErrors))
		return
	}

	sess, err := session.Get(body.SessionID, nil)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("retrieving stripe session", err, logParams...))
		return
	}

	if sess.PaymentStatus != "paid" {
		json.NewEncoder(w).Encode(map[string]any{"status": string(sess.PaymentStatus), "activated": false})
		return
	}

	err = activateSession(r, sess.ID, sess.PaymentIntent)
	if err != nil {
		HandleResponseError(w, r, NewInternalServerError("activating subscription", err, logParams...))
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"status": "paid", "activated": true})
}
