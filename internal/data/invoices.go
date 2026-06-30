package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/validator"
	"github.com/lib/pq"
)

type InvoiceModel struct {
	DB *sql.DB
}

type Invoice struct {
	ID            int       `json:"id"`
	BusinessID    int       `json:"business_id"`
	InvoiceNumber string    `json:"invoice_number"`
	InvoiceDate   time.Time `json:"invoice_date"`

	BusinessName    string `json:"business_name"`
	BusinessLogoURL string `json:"business_logo_url"`
	BusinessRCCM    string `json:"business_rccm,omitempty"`

	ClientName    string `json:"client_name"`
	ClientPhone   string `json:"client_phone,omitempty"`
	ClientEmail   string `json:"client_email,omitempty"`
	ClientAddress string `json:"client_address,omitempty"`

	Items       []*InvoiceItem `json:"items"`
	TotalAmount int            `json:"total_amount"`
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
	UnitPrice   int       `json:"unit_price"`
	CreatedAt   time.Time `json:"-"`
}

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

	for i, item := range invoice.Items {
		ValidateInvoiceItem(v, i, item)
	}
}

func ValidateInvoiceItem(v *validator.Validator, index int, item *InvoiceItem) {
	fieldKey := fmt.Sprintf("items[%d].description", index)
	v.Check(strings.TrimSpace(item.Description) != "", fieldKey, "must be provided")

	priceKey := fmt.Sprintf("items[%d].unit_price", index)
	v.Check(item.UnitPrice > 0, priceKey, "must be a positive amount")

	qtyKey := fmt.Sprintf("items[%d].quantity", index)
	v.Check(item.Quantity >= 1, qtyKey, "must be at least 1")
}

func (m InvoiceModel) Insert(invoice *Invoice) error {
	queryInvoice := `
		INSERT INTO invoices (
			invoice_number, invoice_date, business_id, business_name, business_logo_url, business_rccm,
			client_name, client_phone, client_email, client_address,
			total_amount, currency, note_title, note_text, 
			footer_address, footer_phone, footer_email, status
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, version`

	queryItem := `
		INSERT INTO invoice_items (invoice_id, description, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	calculatedTotal := 0
	for _, item := range invoice.Items {
		calculatedTotal += (item.Quantity * item.UnitPrice)
	}
	invoice.TotalAmount = calculatedTotal

	argsInvoice := []any{
		invoice.InvoiceNumber, invoice.InvoiceDate, invoice.BusinessID, invoice.BusinessName, invoice.BusinessLogoURL, invoice.BusinessRCCM,
		invoice.ClientName, invoice.ClientPhone, invoice.ClientEmail, invoice.ClientAddress,
		invoice.TotalAmount, invoice.Currency, invoice.NoteTitle, invoice.NoteText,
		invoice.FooterAddress, invoice.FooterPhone, invoice.FooterEmail, invoice.Status,
	}

	err = tx.QueryRowContext(ctx, queryInvoice, argsInvoice...).Scan(&invoice.ID, &invoice.CreatedAt, &invoice.Version)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrDuplicateInvoiceNumber
		}
		return err
	}

	for _, item := range invoice.Items {
		item.InvoiceID = invoice.ID
		argsItem := []any{item.InvoiceID, item.Description, item.Quantity, item.UnitPrice}
		err = tx.QueryRowContext(ctx, queryItem, argsItem...).Scan(&item.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m InvoiceModel) Get(id int, businessID int) (*Invoice, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT 
			i.id, i.invoice_number,  i.invoice_date, 
			i.business_id, i.business_name, i.business_logo_url, i.business_rccm,
			i.client_name, i.client_phone, i.client_email, i.client_address,
			i.note_title, i.note_text,
			i.total_amount, i.currency, i.status, i.created_at, i.version,
			it.id, it.description, it.quantity, it.unit_price
		FROM invoices i
		LEFT JOIN invoice_items it ON i.id = it.invoice_id
		WHERE i.id = $1 AND i.business_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, id, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoice Invoice
	items := []*InvoiceItem{}
	isFirstRow := true

	for rows.Next() {
		var item InvoiceItem
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.InvoiceDate,
			&invoice.BusinessID, &invoice.BusinessName, &invoice.BusinessLogoURL, &invoice.BusinessRCCM,
			&invoice.ClientName, &invoice.ClientPhone, &invoice.ClientEmail, &invoice.ClientAddress,
			&invoice.NoteTitle, &invoice.NoteText,
			&invoice.TotalAmount, &invoice.Currency, &invoice.Status, &invoice.CreatedAt, &invoice.Version,
			&item.ID, &item.Description, &item.Quantity, &item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		item.InvoiceID = invoice.ID
		items = append(items, &item)
		isFirstRow = false
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if isFirstRow {
		return nil, ErrRecordNotFound
	}

	invoice.Items = items
	return &invoice, nil
}

func (m InvoiceModel) Update(invoice *Invoice) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE invoices 
		SET invoice_number = $1, invoice_date = $2, business_id = $3, client_name = $4, client_phone = $5, 
			client_email = $6, client_address = $7, total_amount = $8, currency = $9, 
			status = $10, note_title = $11, note_text = $12, footer_address = $13, 
			footer_phone = $14, footer_email = $15, version = version + 1
		WHERE id = $16 AND business_id = $17 AND version = $18
		RETURNING version`

	args := []any{
		invoice.InvoiceNumber, invoice.InvoiceDate, invoice.BusinessID, invoice.ClientName, invoice.ClientPhone,
		invoice.ClientEmail, invoice.ClientAddress, invoice.TotalAmount, invoice.Currency,
		invoice.Status, invoice.NoteTitle, invoice.NoteText, invoice.FooterAddress,
		invoice.FooterPhone, invoice.FooterEmail, invoice.ID, invoice.BusinessID, invoice.Version,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&invoice.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	if invoice.Items != nil {
		_, err = tx.ExecContext(ctx, "DELETE FROM invoice_items WHERE invoice_id = $1", invoice.ID)
		if err != nil {
			return err
		}
		insertQuery := `
			INSERT INTO invoice_items (invoice_id, description, quantity, unit_price) 
			VALUES ($1, $2, $3, $4)`
		for _, item := range invoice.Items {
			_, err = tx.ExecContext(ctx, insertQuery, invoice.ID, item.Description, item.Quantity, item.UnitPrice)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (m InvoiceModel) Delete(id int, businessID int) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM invoices WHERE id = $1 AND business_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, businessID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m InvoiceModel) GetAll(businessID int, clientName string, status string, filters Filters) ([]*Invoice, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, invoice_number, invoice_date, business_id, business_name, 
			   client_name, client_phone, client_email, client_address,
			   total_amount, currency, status, created_at, version
		FROM invoices
		WHERE business_id = $1
		AND (client_name ILIKE '%%' || $2 || '%%' OR $2 = '')
		AND (status = UPPER($3) OR $3 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{businessID, clientName, status, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	invoices := []*Invoice{}

	for rows.Next() {
		var i Invoice
		err := rows.Scan(
			&totalRecords,
			&i.ID, &i.InvoiceNumber, &i.InvoiceDate, &i.BusinessID, &i.BusinessName,
			&i.ClientName, &i.ClientPhone, &i.ClientEmail, &i.ClientAddress,
			&i.TotalAmount, &i.Currency, &i.Status, &i.CreatedAt, &i.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		invoices = append(invoices, &i)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return invoices, metadata, nil
}
