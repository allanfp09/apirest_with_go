package data

import "database/sql"

type Models struct {
	Movies interface {
		InsertMovie(m *Movies) error
		GetMovie(id int64) (*Movies, error)
		UpdateMovie(movie *Movies) (*Movies, error)
		GetAllMovies(title string, genres []string, filters Filters) ([]*Movies, Metadata, error)
	}
	Tokens      TokenModel
	Users       UserModel
	Permissions PermissionModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}

func NewMovieMockModel() Models {
	return Models{
		Movies: MovieMockModel{},
		Users:  UserModel{},
	}
}
