# Notflex API

Backend for **Notflex**, a subscription-based movie streaming platform with AI-powered recommendations.
Built as my bachelor graduation project at the University of Transport and Communications (2026).

Related repositories:

- [notflex-client](https://github.com/notflex-client/notflex-client) – customer web app (Nuxt 4) · [live demo](https://notflex-client.vercel.app)
- [notflex-admin](https://github.com/notflex-client/notflex-admin) – admin dashboard (Nuxt UI) · [live demo](https://notflex-admin.vercel.app)

## Features

**Accounts & auth**
- Registration with e-mail verification code, password login, and passwordless login via one-time e-mail code
- Forgot / reset password flow, change password, bearer-token sessions stored per device (`user_tokens`)
- Multiple viewing profiles per account (`X-Profile-Id` header), profile transfer between accounts
- Role-based authorization (`user` / `admin`) via middleware

**Catalogue & streaming**
- Movies and series with episodes, genres, tags, banners, subtitles and alternate audio tracks
- HLS proxy (`/proxy/hls`) that rewrites `.m3u8` manifests so segments stream through the API
- Image and video upload endpoints for the admin dashboard

**Subscriptions & payments**
- Subscription plans, Stripe Checkout session creation, session verification and signed webhook handling
- Payment history and current subscription per user

**Personalisation & recommendations**
- Watch history, ratings and favourites per profile
- Content-based similarity: `scripts/compute_similarities.py` builds TF-IDF vectors from movie metadata and stores top-N cosine-similar titles in `movie_similarities`
- AI recommendations (`/recommendations/me`): a candidate pool is built from watch history, similar titles and favourite genres, then re-ranked by an LLM (Gemini) with a structured JSON prompt; results are cached per user with a rule-based fallback and a trending fallback when history is empty

**Admin**
- CRUD for movies, episodes, genres, tags, plans, banners, subtitles and audio tracks
- User management, payment list and dashboard statistics

**Platform**
- Localised transactional e-mails (vi / en, from `Accept-Language`) rendered with [templ](https://templ.guide) and sent over SMTP
- Request ID, real IP, structured logging (`slog`), panic recovery, timeouts, CORS, graceful shutdown

## Tech stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP | [chi](https://github.com/go-chi/chi) v5, go-chi/render, go-chi/cors |
| Database | PostgreSQL via GORM |
| Payments | Stripe (Checkout + webhooks) |
| AI | Google Gemini API (LLM re-ranking), scikit-learn TF-IDF (offline similarity) |
| E-mail | templ templates + gomail (SMTP) |
| Validation | go-playground/validator |
| Deployment | Docker (multi-stage, Alpine) |

## Project structure

```
api/              HTTP handlers grouped by domain (auth, movie, subscription, recommendation_ai, admin_*)
middlewares/      authentication, role guard, locale, content-type
models/           GORM models
common/database/  DB connection
helpers/          context, params, validation, translator, Gemini client
mailer/           SMTP sender
templates/        templ e-mail templates (register, login code, reset password)
docs/             init_schema.sql
scripts/          compute_similarities.py (+ requirements.txt)
cmd/seed/         sample data seeder
```

## Getting started

Requirements: Go 1.25+, PostgreSQL 14+, (optional) Python 3.10+ for the similarity script.

```bash
# 1. Configure
cp .env.example .env            # set DB_DSN, SMTP_*, STRIPE_*, GEMINI_API_KEY

# 2. Create the schema
psql "$DB_DSN" -f docs/init_schema.sql

# 3. Seed sample genres, plans and titles
go run ./cmd/seed

# 4. Run the API
go run .                        # listens on PORT from .env

# 5. (optional) Pre-compute similar titles for recommendations
pip install -r scripts/requirements.txt
python scripts/compute_similarities.py --env .env --top-n 10
```

Stripe webhooks in development:

```bash
stripe listen --forward-to localhost:3000/webhooks/stripe
```

Docker:

```bash
docker build -t notflex-api .
docker run --env-file .env -p 8080:8080 notflex-api
```

## API overview

| Group | Auth | Examples |
|---|---|---|
| Public | none | `POST /auth/login`, `POST /registration/request`, `GET /movies`, `GET /movies/{id}/similar`, `GET /subscription/plans`, `GET /proxy/hls`, `POST /webhooks/stripe` |
| User | bearer token | `GET /auth/me`, `GET/POST /profiles`, `POST /subscription/stripe-checkout`, `GET /recommendations/me`, `POST /watch-history`, `POST /ratings`, `POST /favorites/{movieId}` |
| Admin | bearer token + `admin` role | `POST /admin/movies`, `POST /admin/uploads/video`, `GET /admin/users`, `GET /admin/payments`, `GET /admin/stats` |

All routes are registered in `main.go`.

## Author

Vu Nhat Minh – [github.com/MinhNhat232](https://github.com/MinhNhat232)
