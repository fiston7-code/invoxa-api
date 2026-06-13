package data

import (
	"database/sql"
	"errors"
)

// Centralisation des erreurs globales du storage pour ton SaaS
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models regroupe tous les modèles de données du SaaS Invoxa.
// C'est cette structure unique que l'on va injecter dans ton application globale.
type Models struct {
	Invoices InvoiceModel
}

// NewModels initialise et renvoie une structure Models contenant notre pool SQL.
func NewModels(db *sql.DB) Models {
	return Models{
		Invoices: InvoiceModel{DB: db},
	}
}
