package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
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
		app.notFoundResponse(w, r)
		return
	}

	invoice, err := app.models.Invoices.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"invoice": invoice}, nil)
	if err != nil {
		// Use the new serverErrorResponse() helper.
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extraire l'ID de la facture depuis l'URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// 2. Récupérer la facture existante en BDD
	invoice, err := app.models.Invoices.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 3. Vérification du Header de version (Optimistic Concurrency Control préventif)
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.Itoa(int(invoice.Version)) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	// 4. Déclarer la structure d'input avec des POINTEURS pour le mode PATCH
	// 4. Structure INPUT avec TOUS les champs modifiables d'une facture
	var input struct {
		InvoiceNumber *string    `json:"invoice_number"`
		InvoiceDate   *time.Time `json:"invoice_date"`
		ClientName    *string    `json:"client_name"`
		ClientPhone   *string    `json:"client_phone"`
		ClientEmail   *string    `json:"client_email"`
		ClientAddress *string    `json:"client_address"`
		TotalAmount   *int       `json:"total_amount"`
		Currency      *string    `json:"currency"`
		Status        *string    `json:"status"`

		// Ajout des notes et footers manquants
		NoteTitle     *string `json:"note_title"`
		NoteText      *string `json:"note_text"`
		FooterAddress *string `json:"footer_address"`
		FooterPhone   *string `json:"footer_phone"`
		FooterEmail   *string `json:"footer_email"`

		// Ajout du pointeur vers la slice d'items pour le PATCH
		Items *[]struct {
			Description string `json:"description"`
			Quantity    int    `json:"quantity"`
			UnitPrice   int    `json:"unit_price"`
		} `json:"items"`
	}

	// 5. Décoder le JSON de la requête
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 6. Application conditionnelle des modifications (Déréférencement des pointeurs)
	if input.InvoiceNumber != nil {
		invoice.InvoiceNumber = *input.InvoiceNumber
	}
	if input.InvoiceDate != nil {
		invoice.InvoiceDate = *input.InvoiceDate
	}
	if input.ClientName != nil {
		invoice.ClientName = *input.ClientName
	}
	if input.ClientPhone != nil {
		invoice.ClientPhone = *input.ClientPhone
	}
	if input.ClientEmail != nil {
		invoice.ClientEmail = *input.ClientEmail
	}
	if input.ClientAddress != nil {
		invoice.ClientAddress = *input.ClientAddress
	}
	if input.TotalAmount != nil {
		invoice.TotalAmount = *input.TotalAmount
	}
	if input.Currency != nil {
		invoice.Currency = *input.Currency
	}
	if input.Status != nil {
		invoice.Status = *input.Status
	}

	// 7. Validation des données de la facture mise à jour
	v := validator.New()
	if data.ValidateInvoice(v, invoice); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 8. Exécuter la mise à jour en base de données
	err = app.models.Invoices.Update(invoice)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 9. Renvoyer la facture modifiée et sa nouvelle version
	err = app.writeJSON(w, http.StatusOK, envelope{"invoice": invoice}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteInvoiceHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Extraire l'ID de la facture (en int64 grâce à ton helper)
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// 2. Appeler la méthode Delete() de ton modèle
	err = app.models.Invoices.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 3. Renvoyer la confirmation de suppression en JSON
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "invoice and its items successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listInvoicesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ClientName string // Plus utile que BusinessName
		Status     string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	// 1. On filtre sur ce qui change vraiment
	input.ClientName = app.readString(qs, "client_name", "")
	input.Status = app.readString(qs, "status", "")

	// 2. Pagination et tri
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "-invoice_date") // Tri par défaut plus logique : le plus récent

	// 3. Liste blanche adaptée
	input.Filters.SortSafelist = []string{
		"id", "invoice_number", "invoice_date", "total_amount", "client_name",
		"-id", "-invoice_number", "-invoice_date", "-total_amount", "-client_name",
	}

	// 4. Validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 5. Appel au modèle (à implémenter dans invoices.go)
	invoices, metadata, err := app.models.Invoices.GetAll(input.ClientName, input.Status, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 6. Réponse avec metadata
	err = app.writeJSON(w, http.StatusOK, envelope{"invoices": invoices, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
