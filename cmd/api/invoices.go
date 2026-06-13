package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/fiston7-code/invoxa-api/internal/validator"
)

func (app *application) createInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the expected input data from the client.
	var input struct {
		InvoiceNumber string    `json:"invoice_number"`
		InvoiceDate   time.Time `json:"invoice_date"`

		// AJOUTE CES 3 CHAMPS BUSINESS ICI :
		BusinessName    string `json:"business_name"`
		BusinessLogoURL string `json:"business_logo_url"`
		BusinessRCCM    string `json:"business_rccm"`

		ClientName    string `json:"client_name"`
		ClientPhone   string `json:"client_phone"`
		ClientEmail   string `json:"client_email"`
		ClientAddress string `json:"client_address"`

		Items []struct {
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
			UnitPrice   int    `json:"unit_price"` // Aligné en int (centimes) !
		} `json:"items"`

		TotalAmount int    `json:"total_amount"` // Aligné en int (centimes) !
		Currency    string `json:"currency"`

		NoteTitle     string `json:"note_title"`
		NoteText      string `json:"note_text"`
		FooterAddress string `json:"footer_address"`
		FooterPhone   string `json:"footer_phone"`
		FooterEmail   string `json:"footer_email"`

		// AJOUTE LE STATUS ICI :
		Status string `json:"status"`
	}

	// Decode the request body into the input struct.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new domain data.Invoice struct.
	invoice := &data.Invoice{
		InvoiceNumber:   input.InvoiceNumber,
		InvoiceDate:     input.InvoiceDate,     // AJOUTÉ
		BusinessName:    input.BusinessName,    // AJOUTÉ
		BusinessLogoURL: input.BusinessLogoURL, // AJOUTÉ
		BusinessRCCM:    input.BusinessRCCM,    // AJOUTÉ
		ClientName:      input.ClientName,
		ClientPhone:     input.ClientPhone,
		ClientEmail:     input.ClientEmail,
		ClientAddress:   input.ClientAddress,
		TotalAmount:     input.TotalAmount,
		Currency:        input.Currency,
		NoteTitle:       input.NoteTitle,
		NoteText:        input.NoteText,
		FooterAddress:   input.FooterAddress,
		FooterPhone:     input.FooterPhone,
		FooterEmail:     input.FooterEmail,
		Status:          input.Status, // AJOUTÉ
	}

	// Copy the nested input items into the domain model slice.
	for _, item := range input.Items {
		invoice.Items = append(invoice.Items, &data.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Call the updated ValidateInvoice() which automatically handles sub-items loop.
	if data.ValidateInvoice(v, invoice); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//  Call the Insert() method on our Invoices model, passing in the validated pointer.
	// This will write the data to your PostgreSQL instance and return an updated object.

	err = app.models.Invoices.Insert(invoice)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/invoices/%d", invoice.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"invoice": invoice}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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
