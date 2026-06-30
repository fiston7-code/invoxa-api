package data

import (
	"database/sql"
	"errors"
)

// Centralisation des erreurs globales du storage pour ton SaaS
var (
	ErrRecordNotFound         = errors.New("record not found")
	ErrEditConflict           = errors.New("edit conflict")
	ErrDuplicateInvoiceNumber = errors.New("duplicate invoice number")
)

// Models regroupe tous les modèles de données du SaaS Invoxa.
// C'est cette structure unique que l'on va injecter dans ton application globale.
type Models struct {
	Invoices         InvoiceModel
	Permissions      PermissionModel
	BusinessProfiles BusinessProfileModel
	Tokens           TokenModel
	Users            UserModel
}

// NewModels initialise et renvoie une structure Models contenant notre pool SQL.
func NewModels(db *sql.DB) Models {
	return Models{
		Invoices:         InvoiceModel{DB: db},
		Permissions:      PermissionModel{DB: db},
		BusinessProfiles: BusinessProfileModel{DB: db},
		Tokens:           TokenModel{DB: db},
		Users:            UserModel{DB: db},
	}
}
