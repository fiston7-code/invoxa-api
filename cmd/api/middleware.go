package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/fiston7-code/invoxa-api/internal/validator"
	"github.com/tomasen/realip" // New import
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic).
		defer func() {
			// Use the built-in recover() function to check if a panic occurred.
			// If a panic did happen, recover() will return the panic value. If
			// a panic didn't happen, it will return nil.
			pv := recover()
			if pv != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after the response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type any, so we use
				// fmt.Errorf() with the %v verb to coerce it into an error and
				// call our serverErrorResponse() helper. In turn, this will log the
				// error at the ERROR level and send the client a 500 Internal
				// Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// rate limit is a middleware that limits the number of requests a client can make to our API within a given time period. It uses the golang.org/x/time/rate package to implement a token bucket algorithm for rate limiting. The middleware checks if rate limiting is enabled in the application configuration, and if so, it creates a map of clients (identified by their IP addresses) and their corresponding rate limiters. Each client has a limiter that allows a certain number of requests per second (RPS) and a burst size. If a client exceeds the allowed rate, the middleware responds with a 429 Too Many Requests status code.

func (app *application) rateLimit(next http.Handler) http.Handler {
	// If rate limiting is not enabled, return the next handler in the chain with
	// with no further action.

	if !app.config.limiter.enabled {
		return next
	}

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realip.FromRequest(r)
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				// Use the requests-per-second and burst values from the config
				// struct.
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		// Add the "Vary: Access-Control-Request-Method" header.
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")
		if origin != "" {
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Check if the request has the HTTP method OPTIONS and contains the
					// "Access-Control-Request-Method" header. If it does, then we treat
					// it as a preflight request.
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						// Set the necessary preflight response headers, as discussed
						// previously.
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
						// Write the headers along with a 200 OK status and return from
						// the middleware with no further action.
						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")
		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no such header found.
		authorizationHeader := r.Header.Get("Authorization")
		// If there is no Authorization header found, call the next handler in the chain
		// as normal and then return without executing any of the code below.
		if authorizationHeader == "" {
			next.ServeHTTP(w, r)
			return
		}
		// Otherwise, we expect the value of the Authorization header to be in the format
		// "Bearer <token>". We try to split this into its constituent parts, and if the
		// header isn't in the expected format we return a 401 Unauthorized response
		// using the invalidAuthenticationTokenResponse() helper (which we will create
		// in a moment).
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Extract the actual authentication token from the header parts.
		token := headerParts[1]
		// Validate the token to make sure it is in a sensible format.
		v := validator.New()
		// If the token isn't valid, use the invalidAuthenticationTokenResponse()
		// helper to send a response, rather than the failedValidationResponse() helper
		// that we'd normally use.
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		// Retrieve the details of the user associated with the authentication token,
		// again calling the invalidAuthenticationTokenResponse() helper if no
		// matching record was found. IMPORTANT: Notice that we are using
		// ScopeAuthentication as the first parameter here.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// At this point, we know that the current user has authenticated using
		// a valid token, and we add their user information to the request context
		// with the contextSetAuthenticatedUser() helper that we just made.
		r = app.contextSetAuthenticatedUser(r, user)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

//lint:ignore U1000 middleware currently unused, but might be in the future
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetAuthenticatedUser() helper that we made earlier to retrieve
		// the authenticated user information from the request context. If there
		// is no such information in the request context, the found variable will
		// be false and we call the authenticationRequiredResponse() method to inform
		// the client that they should authenticate before trying again.
		authenticatedUser, found := app.contextGetAuthenticatedUser(r)
		if !found {
			app.authenticationRequiredResponse(w, r)
			return
		}
		// If the user is not activated, use the inactiveAccountResponse() helper to
		// inform them that they need to activate their account.
		if !authenticatedUser.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Notice that the first argument in the requirePermission() function represents the
// permission code that we want the user to have.
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the user is authenticated and activated, just like before.
		authenticatedUser, found := app.contextGetAuthenticatedUser(r)
		if !found {
			app.authenticationRequiredResponse(w, r)
			return
		}
		if !authenticatedUser.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		// Get the slice of permissions for the user.
		permissions, err := app.models.Permissions.GetAllForUser(authenticatedUser.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// Check if the slice includes the required permission. If it doesn't, then
		// return a 403 Forbidden response.
		if !permissions.Include(code) {
			app.missingPermissionResponse(w, r)
			return
		}
		// Otherwise they have the required permission so we call the next handler in
		// the chain.
		next.ServeHTTP(w, r)
	})
}
