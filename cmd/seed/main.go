package main

import (
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"notflex_client_api/models"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		godotenv.Load(".env")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect DB: %v", err)
	}

	// â”€â”€ Genres â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
	genres := []models.Genre{
		{Name: "Action"},
		{Name: "Drama"},
		{Name: "Thriller"},
		{Name: "Comedy"},
		{Name: "Sci-Fi"},
		{Name: "Documentary"},
		{Name: "Animation"},
		{Name: "Horror"},
		{Name: "Romance"},
		{Name: "Crime"},
	}
	for i := range genres {
		db.Where(models.Genre{Name: genres[i].Name}).FirstOrCreate(&genres[i])
	}
	fmt.Println("âœ“ Genres seeded")

	// â”€â”€ Movies / Series â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
	ptr := func(s string) *string { return &s }
	ptrI := func(n int16) *int16 { return &n }
	ptrInt := func(n int) *int { return &n }

	titles := []struct {
		movie    models.Movie
		genreIDs []int
	}{
		{
			movie: models.Movie{
				Title:        "House of Ninjas",
				Type:         "series",
				Description:  ptr("A dysfunctional ninja family must return to shadowy missions to counteract a string of looming threats."),
				PosterURL:    ptr("https://image.tmdb.org/t/p/w500/56v2KjBlU4XaOv9rVYEQypROD7P.jpg"),
				BannerURL:    ptr("https://image.tmdb.org/t/p/original/xDMIl84Qo5Tsu62c9DGWhmPI67A.jpg"),
				ReleaseYear:  ptrI(2024),
				Rating:       ptr("TV-MA"),
				DurationMins: nil,
				IsPremium:    true,
			},
			genreIDs: []int{1, 2, 3}, // Action, Drama, Thriller
		},
		{
			movie: models.Movie{
				Title:        "Inception",
				Type:         "movie",
				Description:  ptr("A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O."),
				PosterURL:    ptr("https://image.tmdb.org/t/p/w500/edv5CZvWj09upOsy2Y6IwDhK8bt.jpg"),
				BannerURL:    ptr("https://image.tmdb.org/t/p/original/s3TBrRGB1iav7gFOCNx3H31MoES.jpg"),
				ReleaseYear:  ptrI(2010),
				Rating:       ptr("PG-13"),
				DurationMins: ptrInt(148),
				IsPremium:    true,
			},
			genreIDs: []int{1, 5, 3}, // Action, Sci-Fi, Thriller
		},
		{
			movie: models.Movie{
				Title:        "The Dark Knight",
				Type:         "movie",
				Description:  ptr("When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice."),
				PosterURL:    ptr("https://image.tmdb.org/t/p/w500/qJ2tW6WMUDux911r6m7haRef0WH.jpg"),
				BannerURL:    ptr("https://image.tmdb.org/t/p/original/nMKdUUepR0i5zn0y1T4CejMPAmd.jpg"),
				ReleaseYear:  ptrI(2008),
				Rating:       ptr("PG-13"),
				DurationMins: ptrInt(152),
				IsPremium:    true,
			},
			genreIDs: []int{1, 10, 3}, // Action, Crime, Thriller
		},
		{
			movie: models.Movie{
				Title:        "Interstellar",
				Type:         "movie",
				Description:  ptr("A team of explorers travel through a wormhole in space in an attempt to ensure humanity's survival."),
				PosterURL:    ptr("https://image.tmdb.org/t/p/w500/gEU2QniE6E77NI6lCU6MxlNBvIx.jpg"),
				BannerURL:    ptr("https://image.tmdb.org/t/p/original/xJHokMbljvjADYdit5fK5VQsXEG.jpg"),
				ReleaseYear:  ptrI(2014),
				Rating:       ptr("PG-13"),
				DurationMins: ptrInt(169),
				IsPremium:    true,
			},
			genreIDs: []int{2, 5}, // Drama, Sci-Fi
		},
		{
			movie: models.Movie{
				Title:        "Squid Game",
				Type:         "series",
				Description:  ptr("Hundreds of cash-strapped players accept a strange invitation to compete in children's games. Inside, a deadly game awaits to determine one winner."),
				PosterURL:    ptr("https://image.tmdb.org/t/p/w500/dDlEmu3EZ0Pgg93K2SVNLCjCSvE.jpg"),
				BannerURL:    ptr("https://image.tmdb.org/t/p/original/qw3J9cNeLioOLoR68WX7z79aCdK.jpg"),
				ReleaseYear:  ptrI(2021),
				Rating:       ptr("TV-MA"),
				DurationMins: nil,
				IsPremium:    true,
			},
			genreIDs: []int{2, 10, 3}, // Drama, Crime, Thriller
		},
		{
			movie: models.Movie{
				Title:        "Wednesday",
				Type:         "series",
				Description:  ptr("Follows Wednesday Addams' years as a student at Nevermore Academy, where she attempts to master her emerging psychic ability."),
				PosterURL:    ptr("https://image.tmdb.org/t/p/w500/jeGtaMwGxPmQN5xM4ClnwPQcDef.jpg"),
				BannerURL:    ptr("https://image.tmdb.org/t/p/original/iHSwvRVsRyxpX7FE7GbviaDvgGZ.jpg"),
				ReleaseYear:  ptrI(2022),
				Rating:       ptr("TV-14"),
				DurationMins: nil,
				IsPremium:    false,
			},
			genreIDs: []int{4, 8}, // Comedy, Horror
		},
	}

	for i := range titles {
		m := &titles[i].movie
		// Chá»‰ insert náº¿u chÆ°a tá»“n táº¡i
		var existing models.Movie
		if db.Where("title = ?", m.Title).First(&existing).Error == nil {
			fmt.Printf("  skip (exists): %s\n", m.Title)
			continue
		}
		m.ID = uuid.NewString()
		if err := db.Create(m).Error; err != nil {
			log.Printf("  error inserting %s: %v", m.Title, err)
			continue
		}
		// GÃ¡n genres
		var gs []models.Genre
		db.Where("id IN ?", titles[i].genreIDs).Find(&gs)
		db.Model(m).Association("Genres").Replace(gs)
		fmt.Printf("  âœ“ %s\n", m.Title)
	}

	fmt.Println("âœ“ Movies seeded")
	fmt.Println("\nSeed completed!")
}
