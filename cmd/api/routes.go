package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Set custom error handlers for 404 Not Found and 405 Method Not Allowed responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// --- ROUTES PUBLIC ---
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// private routes : RECEIPT (Protected by Permissions) ---
	// read (need to read  invoices:read)
	router.HandlerFunc(http.MethodGet, "/v1/invoices", app.requirePermission("invoices:read", app.listInvoicesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id", app.requirePermission("invoices:read", app.showInvoiceHandler))
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id/pdf", app.requirePermission("invoices:read", app.downloadInvoicePDFHandler))

	// write (need  invoices:write)
	router.HandlerFunc(http.MethodPost, "/v1/invoices", app.requirePermission("invoices:write", app.createInvoiceHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/invoices/:id", app.requirePermission("invoices:write", app.updateInvoiceHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/invoices/:id", app.requirePermission("invoices:write", app.deleteInvoiceHandler))

	// Dans routes.go
	// fecth business profile (GET) and update it (PATCH)
	router.HandlerFunc(http.MethodGet, "/v1/business", app.requireActivatedUser(app.getBusinessProfileHandler))

	// for creating a business profile (POST) and updating it (PATCH)
	router.HandlerFunc(http.MethodPost, "/v1/business", app.requireActivatedUser(app.createBusinessProfileHandler))

	// PATCH avec :id pour modifier un profil spécifique (optionnel)
	router.HandlerFunc(http.MethodPatch, "/v1/business/:id", app.requireActivatedUser(app.updateBusinessProfileHandler))
	router.Handler(http.MethodPost, "/v1/business/logo", app.requireActivatedUser(http.HandlerFunc(app.uploadLogoHandler)))

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Chaîne globale des middlewares applicatifs
	// return app.recoverPanic(app.rateLimit(app.enableCORS(app.authenticate(router))))
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
