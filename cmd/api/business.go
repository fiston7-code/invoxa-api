package main

import (
	"net/http"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/fiston7-code/invoxa-api/internal/validator"
)

// GET /v1/business - Récupérer automatiquement LE profil de l'utilisateur connecté
func (app *application) getBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
	// 1. On récupère l'user connecté
	user, ok := app.contextGetAuthenticatedUser(r)
	if !ok {
		app.authenticationRequiredResponse(w, r)
		return
	}

	// 2. On va chercher le profil directement via l'ID de cet utilisateur
	profile, err := app.models.BusinessProfiles.GetByUserID(user.ID)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"business": profile}, nil)
}

// POST /v1/business - Création initiale
func (app *application) createBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := app.contextGetAuthenticatedUser(r)
	if !ok {
		app.authenticationRequiredResponse(w, r)
		return
	}

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
		UserID:  user.ID,
		Name:    input.Name,
		LogoURL: input.LogoURL,
		RCCM:    input.RCCM,
		Address: input.Address,
		Phone:   input.Phone,
		Email:   input.Email,
	}

	v := validator.New()
	if data.ValidateBusinessProfile(v, profile); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.models.BusinessProfiles.Insert(profile); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"business": profile}, nil)
}

// PUT /v1/business - Mise à jour du profil de l'utilisateur connecté (Sans ID dans l'URL)
func (app *application) updateBusinessProfileHandler(w http.ResponseWriter, r *http.Request) {
	// 1. On identifie l'user connecté
	user, ok := app.contextGetAuthenticatedUser(r)
	if !ok {
		app.authenticationRequiredResponse(w, r)
		return
	}

	// 2. On récupère SON profil d'entreprise existant
	profile, err := app.models.BusinessProfiles.GetByUserID(user.ID)
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

func (app *application) uploadLogoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parser le formulaire multipart avec une limite de 5 Mo (5 << 20 octets)
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// 2. Récupérer le fichier via le champ "logo" envoyé par Next.js
	file, header, err := r.FormFile("logo")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer file.Close() // Toujours fermer le fichier après utilisation

	// 3. Appeler votre fonction d'upload (celle que vous avez définie précédemment)
	url, err := app.uploadBusinessLogo(header)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 4. Retourner l'URL au format JSON pour que Next.js puisse l'utiliser
	app.writeJSON(w, http.StatusOK, envelope{"url": url}, nil)
}
