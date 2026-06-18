package main

import (
	"net/http"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/fiston7-code/invoxa-api/internal/validator"
)

// GET /v1/business/:id - Récupérer les infos pour le formulaire
func (app *application) getBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	profile, err := app.models.BusinessProfiles.Get(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"business": profile}, nil)
}

// POST /v1/business - Création initiale
func (app *application) createBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		LogoURL string `json:"logo_url"`
		RCCM    string `json:"rccm"`
		Address string `json:"address"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	profile := &data.BusinessProfile{
		Name:    input.Name,
		LogoURL: input.LogoURL,
		RCCM:    input.RCCM,
		Address: input.Address,
		Phone:   input.Phone,
		Email:   input.Email,
	}

	// Validation stricte
	v := validator.New()
	if data.ValidateBusinessProfile(v, profile); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insertion
	if err := app.models.BusinessProfiles.Insert(profile); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"business": profile}, nil)
}

// PUT /v1/business/:id - Mise à jour
func (app *application) updateBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	profile, err := app.models.BusinessProfiles.Get(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	var input struct {
		Name    *string `json:"name"`
		LogoURL *string `json:"logo_url"`
		RCCM    *string `json:"rccm"`
		Address *string `json:"address"`
		Phone   *string `json:"phone"`
		Email   *string `json:"email"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Mise à jour partielle (si le champ est présent dans le JSON)
	if input.Name != nil {
		profile.Name = *input.Name
	}
	if input.LogoURL != nil {
		profile.LogoURL = *input.LogoURL
	}
	if input.RCCM != nil {
		profile.RCCM = *input.RCCM
	}
	if input.Address != nil {
		profile.Address = *input.Address
	}
	if input.Phone != nil {
		profile.Phone = *input.Phone
	}
	if input.Email != nil {
		profile.Email = *input.Email
	}

	v := validator.New()
	if data.ValidateBusinessProfile(v, profile); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.models.BusinessProfiles.Update(profile); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"business": profile}, nil)
}
