package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"movie-api/internal/validators"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type envelope map[string]any

func (app application) getId(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	param := params.ByName("id")

	id, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return 0, err
	}

	if id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.WriteHeader(status)
	w.Write(json)

	return nil
}

func (app application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_000_000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var syntaxError *json.SyntaxError
	var invalidUnmarshalError *json.InvalidUnmarshalError
	var unmarshalTypeError *json.UnmarshalTypeError
	var maxBytesError *http.MaxBytesError

	err := d.Decode(dst)
	if err != nil {
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON syntax at character %d", syntaxError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body cannot be empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON value")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("invalid json type for field %s (%q)", unmarshalTypeError.Field, unmarshalTypeError.Type)
			}
			return fmt.Errorf("body contains invalid JSON type at character %d", unmarshalTypeError.Offset)
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			field := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains an unknown field %s", field)
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("the JSON object is over the appropiated bytes limit (%d)", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}

	}
	err = d.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body contains more than one JSON value")
	}

	return nil
}

func (app *application) readString(qs url.Values, key, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validators.Validators) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddErr(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// background helps a goroutine to recover from panic
func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()

}
