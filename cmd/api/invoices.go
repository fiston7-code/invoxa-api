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
	// 1. On crée le struct anonyme complet pour tout capturer en toute sécurité
	var input struct {
		InvoiceNumber string `json:"invoice_number"`

		// Infos Client
		ClientName    string `json:"client_name"`
		ClientPhone   string `json:"client_phone"`
		ClientEmail   string `json:"client_email"`
		ClientAddress string `json:"client_address"`

		// 📦 Le tableau de lignes (on utilise un struct anonyme imbriqué)
		Items []struct {
			Description string  `json:"description"`
			Quantity    int     `json:"quantity"`
			UnitPrice   float64 `json:"unit_price"`
		} `json:"items"`

		TotalAmount float64 `json:"total_amount"`
		Currency    string  `json:"currency"`

		// Notes et Pied de page
		NoteTitle     string `json:"note_title"`
		NoteText      string `json:"note_text"`
		FooterAddress string `json:"footer_address"`
		FooterPhone   string `json:"footer_phone"`
		FooterEmail   string `json:"footer_email"`
	}

	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &input)
	if err != nil {
		// Use the new badRequestResponse() helper.
		app.badRequestResponse(w, r, err)
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
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
