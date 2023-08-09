package main

import (
	"movie-api/internal/data"
	"movie-api/internal/validators"
	"net/http"
)

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "active",
		"system_info": map[string]string{
			"environment": "development",
			"version":     Version,
		},
	}

	err := app.writeJSON(w, 200, env, nil)
	if err != nil {
		return
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	movie := &data.Movies{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validators.New()
	if data.CheckValidators(v, movie); !v.IsValid() {
		app.failedValidationResponse(w, v.Errors)
		return
	}

	err = app.models.Movies.InsertMovie(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, 201, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getAllMovies(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Year  int32  `json:"year"`
		//Runtime data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
		data.Filters
	}

	v := validators.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, input.Filters); !v.IsValid() {
		app.failedValidationResponse(w, v.Errors)
		return
	}

	movies, metadata, err := app.models.Movies.GetAllMovies(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, 200, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.getId(r)
	if err != nil {
		return
	}

	reqMovie, err := app.models.Movies.GetMovie(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	movie := data.Movies{
		ID:        reqMovie.ID,
		CreatedAt: reqMovie.CreatedAt,
		Title:     reqMovie.Title,
		Year:      reqMovie.Year,
		Runtime:   reqMovie.Runtime,
		Genres:    reqMovie.Genres,
		Version:   reqMovie.Version,
	}

	err = app.writeJSON(w, 200, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {

	movieId, err := app.getId(r)
	if err != nil {
		return
	}

	movie, err := app.models.Movies.GetMovie(movieId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validators.New()
	if data.CheckValidators(v, movie); !v.IsValid() {
		app.failedValidationResponse(w, v.Errors)
		return
	}

	newMovie, err := app.models.Movies.UpdateMovie(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, 200, envelope{"movie": newMovie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
