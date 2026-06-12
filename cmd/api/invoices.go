package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/data"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.
func (app *application) createInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new an invoice")
}

func (app *application) showInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// 1. On récupère l'ID depuis l'URL grâce à ton helper
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 2. On simule (Mock) une facture basée sur ton premier modèle visuel
	invoice := data.Invoice{
		ID:              id,
		InvoiceNumber:   "009/26",
		InvoiceDate:     time.Now(),
		BusinessName:    "JA Agence de Placement",
		BusinessLogoURL: "https://invoxa.storage/logos/ja-agency.png",
		BusinessRCCM:    "CD/KNG/RCCM/22-A-03015 du 04/07/2022",
		ClientName:      "OLYMPIC HOSPITAL",
		ClientPhone:     "+243 000 000 000",
		TotalAmount:     288.00,
		Currency:        "USD",
		Status:          "pending", // ◄ En changeant ça en "paid", ton front affichera le Reçu !
		CreatedAt:       time.Now(),
		Version:         1,
		Items: []*data.InvoiceItem{
			{
				ID:          1,
				Description: "Salaire agent placé – Mars 2026",
				Quantity:    1,
				UnitPrice:   288.00,
			},
		},
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"invoice": invoice}, nil)
	if err != nil {
		// Use the new serverErrorResponse() helper.
		app.serverErrorResponse(w, r, err)
	}
}
