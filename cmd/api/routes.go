package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	// route health
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// new invoices
	router.HandlerFunc(http.MethodPost, "/v1/invoices", app.createInvoiceHandler)
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id", app.showInvoiceHandler)

	return router
}
