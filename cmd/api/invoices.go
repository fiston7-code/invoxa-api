package main

import (
	"fmt"
	"net/http"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (app *application) createInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new an invoice")
}

func (app *application) showInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "show the details of invoice %d\n", id)
}
