package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/validator"
)

type Invoice struct {
	ID            int       `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceDate   time.Time `json:"invoice_date"`

	BusinessName    string `json:"business_name"`
	BusinessLogoURL string `json:"business_logo_url"`
	BusinessRCCM    string `json:"business_rccm,omitempty"`

	ClientName    string `json:"client_name"`
	ClientPhone   string `json:"client_phone,omitempty"`
	ClientEmail   string `json:"client_email,omitempty"`
	ClientAddress string `json:"client_address,omitempty"`

	//  CONTENU ET TOTAL (Relation One-to-Many)
	Items       []*InvoiceItem `json:"items"`
	TotalAmount float64        `json:"total_amount"`
	Currency    string         `json:"currency"`

	NoteTitle     string `json:"note_title,omitempty"`
	NoteText      string `json:"note_text,omitempty"`
	FooterAddress string `json:"footer_address,omitempty"`
	FooterPhone   string `json:"footer_phone,omitempty"`
	FooterEmail   string `json:"footer_email,omitempty"`

	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type InvoiceItem struct {
	ID          int       `json:"id"`
	InvoiceID   int       `json:"invoice_id,omitempty"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	CreatedAt   time.Time `json:"-"`
}

// ValidateInvoice checks all business rules for an invoice and its nested line items.
func ValidateInvoice(v *validator.Validator, invoice *Invoice) {
	v.Check(strings.TrimSpace(invoice.InvoiceNumber) != "", "invoice_number", "must be provided")
	v.Check(strings.TrimSpace(invoice.ClientName) != "", "client_name", "must be provided")

	if invoice.ClientEmail != "" {
		v.Check(validator.Matches(invoice.ClientEmail, validator.EmailRX), "client_email", "must be a valid email address")
	}

	v.Check(invoice.TotalAmount > 0, "total_amount", "must be greater than zero")
	v.Check(validator.PermittedValue(invoice.Currency, "USD", "CDF"), "currency", "must be either USD or CDF")

	v.Check(invoice.Items != nil, "items", "must be provided")
	v.Check(len(invoice.Items) >= 1, "items", "must contain at least 1 item line")

	// 🚀 The loop runs safely inside the data layer now!
	for i, item := range invoice.Items {
		ValidateInvoiceItem(v, i, item)
	}
}

// ValidateInvoiceItem checks rules for an individual line item.
func ValidateInvoiceItem(v *validator.Validator, index int, description string, unitPrice float64, quantity int) {
	fieldKey := fmt.Sprintf("items[%d].description", index)
	v.Check(strings.TrimSpace(description) != "", fieldKey, "must be provided")

	priceKey := fmt.Sprintf("items[%d].unit_price", index)
	v.Check(unitPrice > 0, priceKey, "must be a positive amount")

	qtyKey := fmt.Sprintf("items[%d].quantity", index)
	v.Check(quantity >= 1, qtyKey, "must be at least 1")
}
