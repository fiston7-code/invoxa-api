package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	// Convert the notFoundResponse() helper to a http.Handler using the
	// http.HandlerFunc() adapter, and then set it as the custom error handler for 404
	// Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	// Likewise, convert the methodNotAllowedResponse() helper to a http.Handler and set
	// it as the custom error handler for 405 Method Not Allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/invoices", app.listInvoicesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id/pdf", app.downloadInvoicePDFHandler)
	router.HandlerFunc(http.MethodPost, "/v1/invoices", app.createInvoiceHandler)
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id", app.showInvoiceHandler)

	router.HandlerFunc(http.MethodPatch, "/v1/invoices/:id", app.updateInvoiceHandler)

	// Add the route for the DELETE /v1/invoices/:id endpoint.
	router.HandlerFunc(http.MethodDelete, "/v1/invoices/:id", app.deleteInvoiceHandler)

	router.HandlerFunc(http.MethodGet, "/v1/business/:id", app.getBusinessProfileHandler)
	router.HandlerFunc(http.MethodPost, "/v1/business", app.createBusinessProfileHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/business/:id", app.updateBusinessProfileHandler)
	// router.HandlerFunc(http.MethodPost, "/v1/test-upload", app.testUploadHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// Wrap the router with CORS, rateLimit, and recoverPanic middlewares.
	return app.recoverPanic(app.rateLimit(app.enableCORS(router)))
}
