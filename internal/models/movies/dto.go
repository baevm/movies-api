package movies

import (
	"math"
	"movies-api/internal/validator"
	"strings"
	"time"
)

type MovieFilters struct {
	Title        string
	Genres       []string
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
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

func ValidateFilters(v *validator.Validator, f MovieFilters) {
	v.Check(f.Page > 0, "page", "Page must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "Page must be less than 10 million")
	v.Check(f.PageSize > 1, "page_size", "Page size must be greater than 1")
	v.Check(f.PageSize <= 100, "page_size", "Page size must be less than 100")
	v.Check(validator.AllowedValues(f.Sort, f.SortSafelist...), "sort", "Invalid sort value")
}

func calcMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}

}

func (f MovieFilters) sortColumn() string {
	for _, v := range f.SortSafelist {
		if f.Sort == v {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort param: " + f.Sort)
}

func (f MovieFilters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	} else {
		return "ASC"
	}
}

func (f MovieFilters) limit() int {
	return f.PageSize
}

func (f MovieFilters) offset() int {
	return (f.Page - 1) * f.PageSize
}
