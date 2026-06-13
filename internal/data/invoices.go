package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/validator"
)

// InvoiceModel encapsule le pool de connexions à PostgreSQL.
type InvoiceModel struct {
	DB *sql.DB
}

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

	Items       []*InvoiceItem `json:"items"`
	TotalAmount int            `json:"total_amount"` // Changé en int (centimes)
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
	UnitPrice   int       `json:"unit_price"` // Changé en int (centimes)
	CreatedAt   time.Time `json:"-"`
}

// ValidateInvoice checks all business and structural rules for an invoice instance.
func ValidateInvoice(v *validator.Validator, invoice *Invoice) {
	// Perform validation checks on the main invoice fields.
	v.Check(strings.TrimSpace(invoice.InvoiceNumber) != "", "invoice_number", "must be provided")
	v.Check(strings.TrimSpace(invoice.ClientName) != "", "client_name", "must be provided")

	// Validate email format only if an email address was provided.
	if invoice.ClientEmail != "" {
		v.Check(validator.Matches(invoice.ClientEmail, validator.EmailRX), "client_email", "must be a valid email address")
	}

	// Financial safety checks.
	v.Check(invoice.TotalAmount > 0, "total_amount", "must be greater than zero")
	v.Check(validator.PermittedValue(invoice.Currency, "USD", "CDF"), "currency", "must be either USD or CDF")

	// Ensure that at least one invoice item line is provided.
	v.Check(invoice.Items != nil, "items", "must be provided")
	v.Check(len(invoice.Items) >= 1, "items", "must contain at least 1 item line")

	//  Loop through and automatically validate each individual item row right here!
	for i, item := range invoice.Items {
		ValidateInvoiceItem(v, i, item)
	}
}

// ValidateInvoiceItem checks validation constraints for an individual line item.
func ValidateInvoiceItem(v *validator.Validator, index int, item *InvoiceItem) {
	fieldKey := fmt.Sprintf("items[%d].description", index)
	v.Check(strings.TrimSpace(item.Description) != "", fieldKey, "must be provided")

	priceKey := fmt.Sprintf("items[%d].unit_price", index)
	v.Check(item.UnitPrice > 0, priceKey, "must be a positive amount")

	qtyKey := fmt.Sprintf("items[%d].quantity", index)
	v.Check(item.Quantity >= 1, qtyKey, "must be at least 1")
}

// Insert écrit une nouvelle facture et ses articles associés dans PostgreSQL.
func (m InvoiceModel) Insert(invoice *Invoice) error {
	// 1. Définir la requête SQL pour l'en-tête de la facture
	queryInvoice := `
		INSERT INTO invoices (
			invoice_number, invoice_date, business_name, business_logo_url, business_rccm,
			client_name, client_phone, client_email, client_address,
			total_amount, currency, note_title, note_text, 
			footer_address, footer_phone, footer_email, status
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, version`

	// 2. Définir la requête SQL pour les articles de la facture
	queryItem := `
		INSERT INTO invoice_items (invoice_id, description, quantity, unit_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	// 3. Créer un contexte avec un timeout de 3 secondes
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 4. Démarrer la transaction SQL
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Au cas où le code panique ou s'arrête brutalement, on s'assure d'annuler les modifs
	defer tx.Rollback()

	// 5. Étape A : Insérer l'en-tête de la facture
	argsInvoice := []any{
		invoice.InvoiceNumber, invoice.InvoiceDate, invoice.BusinessName, invoice.BusinessLogoURL, invoice.BusinessRCCM,
		invoice.ClientName, invoice.ClientPhone, invoice.ClientEmail, invoice.ClientAddress,
		invoice.TotalAmount, invoice.Currency, invoice.NoteTitle, invoice.NoteText,
		invoice.FooterAddress, invoice.FooterPhone, invoice.FooterEmail, invoice.Status,
	}

	// On exécute et on récupère directement les valeurs générées par Postgres (ID, CreatedAt, Version)
	err = tx.QueryRowContext(ctx, queryInvoice, argsInvoice...).Scan(&invoice.ID, &invoice.CreatedAt, &invoice.Version)
	if err != nil {
		return err
	}

	// 6. Étape B : Insérer chaque article lié à cette facture
	for _, item := range invoice.Items {
		// On injecte l'ID de la facture tout juste créée dans la clé étrangère de l'article
		item.InvoiceID = invoice.ID

		argsItem := []any{item.InvoiceID, item.Description, item.Quantity, item.UnitPrice}

		// On exécute l'insertion de la ligne et on récupère son ID généré
		err = tx.QueryRowContext(ctx, queryItem, argsItem...).Scan(&item.ID)
		if err != nil {
			return err
		}
	}

	// 7. Étape C : Si tout s'est bien passé, on valide définitivement la transaction
	return tx.Commit()
}

func (m InvoiceModel) Get(id int) (*Invoice, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Requête SQL avec une jointure pour ramener la facture ET ses articles d'un coup
	query := `
        SELECT 
            i.id, i.invoice_number, i.invoice_date, 
            i.business_name, i.business_logo_url, i.business_rccm,
            i.client_name, i.client_phone, i.client_email, i.client_address,
            i.total_amount, i.currency, i.status, i.created_at, i.version,
            it.id, it.description, it.quantity, it.unit_price
        FROM invoices i
        LEFT JOIN invoice_items it ON i.id = it.invoice_id
        WHERE i.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Contrairement à QueryRow, on utilise Query ici car la jointure va retourner
	// autant de lignes qu'il y a d'articles dans la facture.
	rows, err := m.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoice Invoice
	// Slice temporaire pour accumuler les articles
	items := []*InvoiceItem{}

	isFirstRow := true

	for rows.Next() {
		var item InvoiceItem
		// On scanne à la fois les champs de la facture et ceux de l'article courant
		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &invoice.InvoiceDate,
			&invoice.BusinessName, &invoice.BusinessLogoURL, &invoice.BusinessRCCM,
			&invoice.ClientName, &invoice.ClientPhone, &invoice.ClientEmail, &invoice.ClientAddress,
			&invoice.TotalAmount, &invoice.Currency, &invoice.Status, &invoice.CreatedAt, &invoice.Version,
			&item.ID, &item.Description, &item.Quantity, &item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}

		// On lie l'article à la facture
		item.InvoiceID = invoice.ID
		items = append(items, &item)

		isFirstRow = false
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Si la boucle n'a jamais tourné, c'est que la facture n'existe pas
	if isFirstRow {
		return nil, ErrRecordNotFound
	}

	// On injecte la slice d'articles dans notre structure Invoice
	invoice.Items = items

	return &invoice, nil
}

// Update met à jour les informations d'une facture et gère le verrouillage optimiste.
func (i InvoiceModel) Update(invoice *Invoice) error {
	// TODO: Implémenter la mise à jour SQL avec vérification de version
	return nil
}

func (i InvoiceModel) Delete(id int) error {
	// Si l'ID n'est pas valide, on s'arrête tout de suite
	if id < 1 {
		return ErrRecordNotFound
	}

	// Requête SQL de suppression ciblée
	query := `
        DELETE FROM invoices
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// On utilise ExecContext car on ne s'attend pas à recevoir des lignes de données en retour
	result, err := i.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// On vérifie combien de lignes ont été affectées par la requête
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// Si aucune ligne n'a été modifiée, c'est que la facture n'existait pas en BDD
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
