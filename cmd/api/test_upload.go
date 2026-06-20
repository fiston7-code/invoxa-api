package main

import (
	"net/http"
)

func (app *application) testUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Limiter la taille du fichier à 5 Mo pour protéger ton serveur
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		app.logger.Error("Erreur lecture formulaire", "error", err.Error())
		http.Error(w, "Fichier trop volumineux ou invalide", http.StatusBadRequest)
		return
	}

	// 2. Récupérer le fichier depuis la clé "logo" du formulaire
	file, header, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "Clé 'logo' manquante dans le formulaire", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. Appeler ta fonction Cloudinary dédiée
	app.logger.Info("Téléversement vers Cloudinary en cours...")
	secureURL, err := app.uploadBusinessLogo(header)
	if err != nil {
		app.logger.Error("Échec Cloudinary", "error", err.Error())
		http.Error(w, "Erreur lors de l'envoi à Cloudinary", http.StatusInternalServerError)
		return
	}

	// 4. Retourner l'URL magique reçue
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "url": "` + secureURL + `"}`))
}
