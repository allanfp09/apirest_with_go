package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"movie-api/internal/validators"
	"time"
)

type MovieModel struct {
	DB *sql.DB
}

type Movies struct {
	ID        int64     `json:"id,omitempty"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title,omitempty"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
}

func CheckValidators(v *validators.Validators, m *Movies) {
	v.Check(m.Title != "", "title", "title must be provided")
	v.Check(m.Year != 0, "year", "year field must be provided")
	v.Check(m.Year >= 1990, "year", "year value must be from 1990 to actual year")
	v.Check(m.Runtime >= 20, "runtime", "runtime value must be equal or greater 20 minutes")
	v.Check(m.Runtime != 0, "runtime", "runtime field must be provided")
	v.Check(len(m.Genres) != 0, "genres", "at least one genre must be provided")
	v.Check(validators.Unique(m.Genres), "genres", "genres values cannot be duplicated")
}

func (mm MovieModel) InsertMovie(m *Movies) error {
	query := `INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return mm.DB.QueryRowContext(ctx, query, m.Title, m.Year, m.Runtime, pq.Array(m.Genres)).Scan(&m.ID, &m.CreatedAt, &m.Version)
}

func (mm MovieModel) GetAllMovies(title string, genres []string, filters Filters) ([]*Movies, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT 
		    count(*) OVER(), id, created_at, title, year, runtime, genres, version 
		FROM 
		    movies 
		WHERE 
		    (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1)
		   OR $1='')
		AND
		    (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := mm.DB.QueryContext(ctx, query, title, pq.Array(genres), filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var movies []*Movies

	for rows.Next() {
		var movie Movies

		args := []any{&totalRecords, &movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version}
		err := rows.Scan(args...)
		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}

func (mm MovieModel) GetMovie(id int64) (*Movies, error) {
	query := `SELECT id, created_at, title, year, runtime, genres, version FROM movies WHERE id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var movie Movies

	args := []any{&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version}

	row := mm.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (mm MovieModel) UpdateMovie(movie *Movies) (*Movies, error) {
	query := `UPDATE movies SET title=$1, year=$2, runtime=$3, genres=$4, version=version+1 WHERE id=$5 AND version=$6 RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	args := []any{&movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.ID, &movie.Version}
	row := mm.DB.QueryRowContext(ctx, query, args...)
	err := row.Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			return nil, err
		}
	}

	return movie, nil
}

type MovieMockModel struct{}

func (mmm MovieMockModel) InsertMovie(m *Movies) error                { return nil }
func (mmm MovieMockModel) GetMovie(id int64) (*Movies, error)         { return nil, nil }
func (mmm MovieMockModel) UpdateMovie(movie *Movies) (*Movies, error) { return nil, nil }
func (mmm MovieMockModel) GetAllMovies(title string, genres []string, filters Filters) ([]*Movies, Metadata, error) {
	return nil, Metadata{}, nil
}
