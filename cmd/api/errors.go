package main

import "net/http"

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (app *application) errorResponse(w http.ResponseWriter, status int, message any) {
	env := envelope{"error": message}
	app.writeJSON(w, status, env, nil)
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "sorry the server could not process your response"
	app.errorResponse(w, http.StatusInternalServerError, envelope{"error": message})
}

func (app *application) failedValidationResponse(w http.ResponseWriter, errors map[string]string) {
	app.errorResponse(w, http.StatusExpectationFailed, errors)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, http.StatusTooManyRequests, message)
}