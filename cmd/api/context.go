package main

import (
	"context"
	"net/http"

	"github.com/fiston7-code/invoxa-api/internal/data"
)

// Define a custom contextKey type, with the underlying type string.
type contextKey string

// Convert the string "authenticatedUser" to a contextKey type and assign it to the
// authenticatedUserContextKey constant. We'll use this constant as the key for getting
// and setting authenticated user information in the request context.
const authenticatedUserContextKey = contextKey("authenticatedUser")

// The contextSetAuthenticatedUser() method returns a new copy of the request with the
// provided User struct added to the context. Note that we use our
// authenticatedUserContextKey constant as the key.
func (app *application) contextSetAuthenticatedUser(r *http.Request, user data.User) *http.Request {
	ctx := context.WithValue(r.Context(), authenticatedUserContextKey, user)
	return r.WithContext(ctx)
}

// The contextGetAuthenticatedUser() method retrieves the User struct from the request
// context. If the context does not contain a valid User struct under the
// authenticatedUserContextKey key, then the type assertion will fail and the value of
// the `ok` variable will be false.
func (app *application) contextGetAuthenticatedUser(r *http.Request) (data.User, bool) {
	user, ok := r.Context().Value(authenticatedUserContextKey).(data.User)
	return user, ok
}
