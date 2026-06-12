package data

import (
	"time"
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

	// 📦 CONTENU ET TOTAL (Relation One-to-Many)
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
