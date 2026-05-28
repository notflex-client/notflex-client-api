package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"notflex_client_api/common/database"
	"notflex_client_api/helpers"
	"notflex_client_api/models"
)

type aiRecCacheEntry struct {
	items   []models.Movie
	source  string
	expires time.Time
}

var (
	aiRecCache   = sync.Map{}
	aiRecCacheTTL = 30 * time.Minute
)

type llmRankedItem struct {
	MovieID string  `json:"movie_id"`
	Score   float64 `json:"score"`
	Reason  string  `json:"reason"`
}

func GetMyRecommendations(w http.ResponseWriter, r *http.Request) {
	user, _ := helpers.GetUserFromContext(r.Context())
	logParams := []any{"handler", "GetMyRecommendations", "userID", user.ID}

	if cached, ok := loadCache(user.ID); ok {
		json.NewEncoder(w).Encode(map[string]any{
			"items":  cached.items,
			"source": cached.source,
			"cached": true,
		})
		return
	}

	watched, ratings := loadUserHistory(r.Context(), user.ID)

	if len(watched) == 0 && len(ratings) == 0 {
		items, source := fallbackTrending(r.Context())
		storeCache(user.ID, items, source)
		json.NewEncoder(w).Encode(map[string]any{"items": items, "source": source})
		return
	}

	candidates := buildCandidatePool(r.Context(), user.ID, watched)
	if len(candidates) == 0 {
		items, source := fallbackTrending(r.Context())
		storeCache(user.ID, items, source)
		json.NewEncoder(w).Encode(map[string]any{"items": items, "source": source})
		return
	}

	ranked, err := callLLMRanker(r.Context(), watched, ratings, candidates)
	if err != nil {
		items, source := ruleBasedRecommendation(candidates)
		storeCache(user.ID, items, source)
		json.NewEncoder(w).Encode(map[string]any{
			"items":     items,
			"source":    source,
			"llm_error": err.Error(),
		})
		_ = logParams
		return
	}

	candidateByID := make(map[string]models.Movie, len(candidates))
	for _, m := range candidates {
		candidateByID[m.ID] = m
	}

	final := make([]models.Movie, 0, len(ranked))
	for _, item := range ranked {
		movie, ok := candidateByID[item.MovieID]
		if !ok {
			continue
		}
		if item.Score < 0 || item.Score > 1 {
			continue
		}
		final = append(final, movie)
	}

	if len(final) == 0 {
		items, source := ruleBasedRecommendation(candidates)
		storeCache(user.ID, items, source)
		json.NewEncoder(w).Encode(map[string]any{"items": items, "source": source})
		return
	}

	storeCache(user.ID, final, "ai-gemini")
	json.NewEncoder(w).Encode(map[string]any{"items": final, "source": "ai-gemini"})
}

func loadUserHistory(ctx context.Context, userID string) ([]models.Movie, map[string]int) {
	var history []models.WatchHistory
	database.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Movie.Genres").
		Order("watched_at DESC").
		Limit(20).
		Find(&history)

	watched := make([]models.Movie, 0, len(history))
	for _, h := range history {
		if h.Movie != nil {
			watched = append(watched, *h.Movie)
		}
	}

	var ratingRows []models.MovieRating
	database.DB.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&ratingRows)
	ratings := make(map[string]int, len(ratingRows))
	for _, r := range ratingRows {
		ratings[r.MovieID] = r.Rating
	}

	return watched, ratings
}

func buildCandidatePool(ctx context.Context, userID string, watched []models.Movie) []models.Movie {
	watchedIDs := make([]string, 0, len(watched))
	genreIDSet := make(map[int]struct{})
	for _, m := range watched {
		watchedIDs = append(watchedIDs, m.ID)
		for _, g := range m.Genres {
			genreIDSet[g.ID] = struct{}{}
		}
	}

	candidates := make([]models.Movie, 0, 40)
	query := database.DB.WithContext(ctx).
		Model(&models.Movie{}).
		Preload("Genres").
		Preload("Tags")

	if len(watchedIDs) > 0 {
		query = query.Where("movies.id NOT IN ?", watchedIDs)
	}

	if len(genreIDSet) > 0 {
		genreIDs := make([]int, 0, len(genreIDSet))
		for id := range genreIDSet {
			genreIDs = append(genreIDs, id)
		}
		query = query.
			Joins("JOIN movie_genres mg ON mg.movie_id = movies.id").
			Where("mg.genre_id IN ?", genreIDs).
			Distinct("movies.*")
	}

	query.Order("movies.avg_rating DESC, movies.created_at DESC").Limit(40).Find(&candidates)

	if len(candidates) < 30 {
		fill := make([]models.Movie, 0, 30-len(candidates))
		fillQuery := database.DB.WithContext(ctx).
			Model(&models.Movie{}).
			Preload("Genres").
			Preload("Tags").
			Order("avg_rating DESC").
			Limit(30 - len(candidates))
		seen := make(map[string]struct{}, len(candidates))
		for _, m := range candidates {
			seen[m.ID] = struct{}{}
		}
		for _, id := range watchedIDs {
			seen[id] = struct{}{}
		}
		excludeIDs := make([]string, 0, len(seen))
		for id := range seen {
			excludeIDs = append(excludeIDs, id)
		}
		if len(excludeIDs) > 0 {
			fillQuery = fillQuery.Where("id NOT IN ?", excludeIDs)
		}
		fillQuery.Find(&fill)
		candidates = append(candidates, fill...)
	}

	_ = userID
	return candidates
}

func callLLMRanker(ctx context.Context, watched []models.Movie, ratings map[string]int, candidates []models.Movie) ([]llmRankedItem, error) {
	prompt := buildRankerPrompt(watched, ratings, candidates)

	rawJSON, err := helpers.CallGeminiJSON(ctx, prompt)
	if err != nil {
		return nil, err
	}

	rawJSON = strings.TrimSpace(rawJSON)
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimPrefix(rawJSON, "```")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	var ranked []llmRankedItem
	err = json.Unmarshal([]byte(rawJSON), &ranked)
	if err != nil {
		return nil, fmt.Errorf("parse llm json: %w", err)
	}

	return ranked, nil
}

func buildRankerPrompt(watched []models.Movie, ratings map[string]int, candidates []models.Movie) string {
	var sb strings.Builder
	sb.WriteString("You are a movie recommendation engine. Your job is to rank candidate movies for a user based on their watch history.\n\n")
	sb.WriteString("USER'S RECENT WATCH HISTORY (most recent first):\n")
	for _, m := range watched {
		genreNames := make([]string, 0, len(m.Genres))
		for _, g := range m.Genres {
			genreNames = append(genreNames, g.Name)
		}
		line := fmt.Sprintf("- %q", m.Title)
		if len(genreNames) > 0 {
			line += " [" + strings.Join(genreNames, ", ") + "]"
		}
		if m.ReleaseYear != nil {
			line += fmt.Sprintf(" (%d)", *m.ReleaseYear)
		}
		if rating, ok := ratings[m.ID]; ok {
			line += fmt.Sprintf(" — user rated %d/5", rating)
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\nCANDIDATE MOVIES (you MUST only pick from this list, using the exact movie_id):\n")
	for _, m := range candidates {
		genreNames := make([]string, 0, len(m.Genres))
		for _, g := range m.Genres {
			genreNames = append(genreNames, g.Name)
		}
		desc := ""
		if m.Description != nil {
			d := *m.Description
			if len(d) > 120 {
				d = d[:120] + "..."
			}
			desc = " — " + d
		}
		year := ""
		if m.ReleaseYear != nil {
			year = fmt.Sprintf(" (%d)", *m.ReleaseYear)
		}
		sb.WriteString(fmt.Sprintf("id=%s | %q%s | genres=[%s]%s\n",
			m.ID, m.Title, year, strings.Join(genreNames, ", "), desc))
	}

	sb.WriteString("\nTASK:\n")
	sb.WriteString("Return a JSON ARRAY of up to 10 best-matching movies from CANDIDATE MOVIES, sorted by relevance.\n")
	sb.WriteString("Each item must have exactly these fields:\n")
	sb.WriteString(`  - "movie_id": string (must match one of the candidate IDs exactly)` + "\n")
	sb.WriteString(`  - "score": number between 0 and 1` + "\n")
	sb.WriteString(`  - "reason": short string explaining why (in Vietnamese)` + "\n")
	sb.WriteString("\nIMPORTANT:\n")
	sb.WriteString("- DO NOT invent movies. Only use IDs from the CANDIDATE MOVIES list.\n")
	sb.WriteString("- Return ONLY the JSON array, no markdown, no commentary.\n")
	sb.WriteString("\nExample output:\n")
	sb.WriteString(`[{"movie_id":"abc-123","score":0.92,"reason":"Cùng thể loại sci-fi và đạo diễn Nolan"}]` + "\n")

	return sb.String()
}

func ruleBasedRecommendation(candidates []models.Movie) ([]models.Movie, string) {
	if len(candidates) > 12 {
		return candidates[:12], "rule-based-fallback"
	}
	return candidates, "rule-based-fallback"
}

func fallbackTrending(ctx context.Context) ([]models.Movie, string) {
	movies := make([]models.Movie, 0, 12)
	database.DB.WithContext(ctx).
		Preload("Genres").
		Preload("Tags").
		Order("avg_rating DESC, created_at DESC").
		Limit(12).
		Find(&movies)
	return movies, "trending-fallback"
}

func loadCache(userID string) (aiRecCacheEntry, bool) {
	v, ok := aiRecCache.Load(userID)
	if !ok {
		return aiRecCacheEntry{}, false
	}
	entry := v.(aiRecCacheEntry)
	if time.Now().After(entry.expires) {
		aiRecCache.Delete(userID)
		return aiRecCacheEntry{}, false
	}
	return entry, true
}

func storeCache(userID string, items []models.Movie, source string) {
	aiRecCache.Store(userID, aiRecCacheEntry{
		items:   items,
		source:  source,
		expires: time.Now().Add(aiRecCacheTTL),
	})
}
