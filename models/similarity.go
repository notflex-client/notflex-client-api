package models

type MovieSimilarity struct {
	MovieID        string  `gorm:"type:uuid;primaryKey" json:"movie_id"`
	SimilarMovieID string  `gorm:"type:uuid;primaryKey" json:"similar_movie_id"`
	SimilarMovie   *Movie  `gorm:"foreignKey:SimilarMovieID" json:"similar_movie,omitempty"`
	Score          float64 `gorm:"type:numeric(6,5)" json:"score"`
	Rank           int     `json:"rank"`
}
