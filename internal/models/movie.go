package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

type Movie struct {
	Id        int64     `json:"id"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   int32     `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	CreatedAt time.Time `json:"-"`
	Version   int32     `json:"version"`
}

type MovieService struct {
	db *sql.DB
}

func NewMovieService(db *sql.DB) *MovieService {
	return &MovieService{db: db}
}

func (m MovieService) Create(movie *Movie) error {
	query := `
	INSERT INTO movies (title, year, runtime, genres) 
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.db.
		QueryRowContext(ctx, query, args...).
		Scan(&movie.Id, &movie.CreatedAt, &movie.Version)
}

func (m MovieService) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	var movie Movie

	query := `
	SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	WHERE id = $1`

	// create context that will release after 5 second
	// if db query is not completed
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.
		QueryRowContext(ctx, query, id).
		Scan(
			&[]byte{},
			&movie.Id,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		} else {
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieService) Update(movie *Movie) error {
	query := `
	UPDATE movies
	SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(&movie.Genres),
		movie.Id,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.db.
		QueryRowContext(ctx, query, args...).
		Scan(&movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieService) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := ` 
	DELETE FROM movies
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.db.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
