package dto

import (
	"movies-api/internal/models"
	"movies-api/internal/validator"
	"time"
)

func ValidateMovie(v *validator.Validator, movie *models.Movie) {
	v.Check(movie.Title != "", "title", "Title must be provided")
	v.Check(len(movie.Title) <= 500, "title", "Title length must be less than 500 characters")

	v.Check(movie.Year != 0, "year", "Year must be provided")
	v.Check(movie.Year >= 1800, "year", "Year must be greater than 1800")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "Year cant be greater than current year")

	v.Check(movie.Runtime != 0, "runtime", "Runtime must be provided")
	v.Check(movie.Runtime > 0, "runtime", "Runtime cant be less than 0")

	v.Check(movie.Genres != nil, "genres", "Genres must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "Genres must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "Genres must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "Genres must not contain duplicate values")
}
