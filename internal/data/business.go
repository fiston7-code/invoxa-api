package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/validator"
)

type BusinessProfile struct {
	ID int `json:"id"`
	// UserID    int64     `json:"-"` // Utilisateur propriétaire
	Name      string    `json:"name"`
	LogoURL   string    `json:"logo_url"`
	RCCM      string    `json:"rccm"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BusinessProfileModel struct {
	DB *sql.DB
}

// Insert crée le profil lors de la première configuration
func (m BusinessProfileModel) Insert(p *BusinessProfile) error {
	query := `
		INSERT INTO business_profiles (name, logo_url, rccm, address, phone, email)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	args := []any{p.Name, p.LogoURL, p.RCCM, p.Address, p.Phone, p.Email}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&p.ID, &p.CreatedAt)
}

// GetByUserID récupère le profil de l'utilisateur connecté
// func (m BusinessProfileModel) GetByUserID(userID int64) (*BusinessProfile, error) {
// 	query := `
// 		SELECT id, name, logo_url, rccm, address, phone, email, created_at, updated_at
// 		FROM business_profiles
// 		WHERE user_id = $1`

// 	var p BusinessProfile
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
// 		&p.ID, &p.Name, &p.LogoURL, &p.RCCM, &p.Address, &p.Phone, &p.Email, &p.CreatedAt, &p.UpdatedAt,
// 	)

// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, errors.New("profil non trouvé")
// 		}
// 		return nil, err
// 	}
// 	p.UserID = userID
// 	return &p, nil
// }

// Get récupère un profil business par son ID
func (m BusinessProfileModel) Get(id int) (*BusinessProfile, error) {
	query := `
        SELECT id, name, logo_url, rccm, address, phone, email, created_at, updated_at
        FROM business_profiles
        WHERE id = $1`

	var p BusinessProfile
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.LogoURL, &p.RCCM, &p.Address, &p.Phone, &p.Email, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("profil non trouvé")
		}
		return nil, err
	}

	return &p, nil
}

// Update met à jour le profil de l'utilisateur
func (m BusinessProfileModel) Update(p *BusinessProfile) error {
	query := `
		UPDATE business_profiles 
		SET name = $1, logo_url = $2, rccm = $3, address = $4, phone = $5, email = $6, updated_at = NOW()
		WHERE id = $7`

	args := []any{p.Name, p.LogoURL, p.RCCM, p.Address, p.Phone, p.Email, p.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func ValidateBusinessProfile(v *validator.Validator, profile *BusinessProfile) {
	v.Check(strings.TrimSpace(profile.Name) != "", "name", "must be provided")
	v.Check(len(profile.Name) <= 100, "name", "must not be more than 100 bytes long")

	v.Check(len(profile.LogoURL) <= 500, "logo_url", "must not be more than 500 bytes long")
	v.Check(len(profile.RCCM) <= 100, "rccm", "must not be more than 100 bytes long")
	v.Check(len(profile.Address) <= 500, "address", "must not be more than 500 bytes long")
	v.Check(len(profile.Phone) <= 30, "phone", "must not be more than 30 bytes long")

	if profile.Email != "" {
		v.Check(len(profile.Email) <= 255, "email", "must not be more than 255 bytes long")
		v.Check(validator.Matches(profile.Email, validator.EmailRX), "email", "must be a valid email address")
	}
}
