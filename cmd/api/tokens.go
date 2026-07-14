package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/fiston7-code/invoxa-api/internal/validator"
	"google.golang.org/api/idtoken"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the email and password from the request body.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the email and password provided by the client.
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup the user record based on the email address. If no matching user was
	// found, then we call the app.invalidCredentialsResponse() helper to send a 401
	// Unauthorized response to the client (we will create this helper in a moment).
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Check if the provided password matches the actual password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// If the passwords don't match, then we call the app.invalidCredentialsResponse()
	// helper again and return.
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	// Otherwise, if the password is correct, we generate a new token with a 24-hour
	// expiry time and the scope 'authentication'.
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	err = app.writeJSON(w, http.StatusCreated, envelope{
		"authentication_token": map[string]string{
			"token": token.Plaintext, // ✅ JUSTE LE PLAINTEXT
			"type":  "Bearer",
		},
		"user": user, // ✅ Bonus: Envoie aussi l'user
	}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// POST /v1/tokens/google - Connexion (ou création de compte) via Google
func (app *application) createGoogleAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDToken string `json:"id_token"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(input.IDToken != "", "id_token", "must be provided")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Vérifie la signature du JWT auprès des clés publiques de Google, ET
	// que l'audience correspond bien à TON client_id.
	payload, err := idtoken.Validate(r.Context(), input.IDToken, app.config.google.clientID)
	if err != nil {
		app.invalidCredentialsResponse(w, r)
		return
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)

	if email == "" || !emailVerified {
		app.invalidCredentialsResponse(w, r)
		return
	}

	user, err := app.models.Users.GetByEmail(email)
	if err != nil {
		if !errors.Is(err, data.ErrRecordNotFound) {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Aucun compte avec cet email : on en crée un à la volée, comme
		// registerUserHandler. L'email est déjà vérifié par Google, donc
		// Activated=true directement.
		newUser := data.User{
			Name:      name,
			Email:     email,
			Activated: true,
		}

		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		if err := newUser.Password.Set(base64.URLEncoding.EncodeToString(randomBytes)); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if data.ValidateUser(v, newUser); !v.Valid() {
			app.failedValidationResponse(w, r, v.Errors)
			return
		}

		newUser, err = app.models.Users.Insert(newUser)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrDuplicateEmail):
				user, err = app.models.Users.GetByEmail(email)
				if err != nil {
					app.serverErrorResponse(w, r, err)
					return
				}
			default:
				app.serverErrorResponse(w, r, err)
				return
			}
		} else {
			user = newUser

			err = app.models.Permissions.AddForUser(user.ID, "invoices:read", "invoices:write")
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{
		"authentication_token": map[string]string{
			"token": token.Plaintext,
			"type":  "Bearer",
		},
		"user": user,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
