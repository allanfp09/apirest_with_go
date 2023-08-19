package main

import (
	"errors"
	"movie-api/internal/data"
	"movie-api/internal/validators"
	"net/http"
	"time"
)

func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		return
	}

	v := validators.New()

	if data.ValidateEmail(v, input.Email); !v.IsValid() {
		app.failedValidationResponse(w, v.Errors)
		return
	}

	user, err := app.models.Users.GetUserByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddErr("email", "no matching email address found")
			app.failedValidationResponse(w, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	if user.Activated {
		v.AddErr("email", "user has already been activated")
		app.failedValidationResponse(w, v.Errors)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Email the user with their additional activation token.
	app.background(func() {
		templateData := map[string]any{
			"activationToken": token.Plaintext}
		// Since email addresses MAY be case sensitive, notice that we are sending this
		//email using the address stored in our database for the user --- not to the // input.Email address provided by the client in this request.
		err = app.mailer.Send(user.Email, "token_activation.tmpl", templateData)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	// Send a 202 Accepted response and confirmation message to the client.
	env := envelope{"message": "an email will be sent to you containing activation instructions"}
	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
