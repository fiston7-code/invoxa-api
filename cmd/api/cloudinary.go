package main

import (
	"context"
	"mime/multipart"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// initCloudinary remplace ta fonction credentials()
// Elle utilise proprement la variable d'environnement de ton fichier .env
func initCloudinary() (*cloudinary.Cloudinary, error) {
	// Le SDK cherche automatiquement la variable CLOUDINARY_URL
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return nil, err
	}

	// Sécurise toutes les URLs générées en HTTPS pour éviter les bugs d'affichage PDF
	cld.Config.URL.Secure = true

	return cld, nil
}

// uploadBusinessLogo remplace ton ancienne fonction uploadImage()
// Elle est maintenant rattachée à ta structure globale "application"
func (app *application) uploadBusinessLogo(fileHeader *multipart.FileHeader) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Ouvrir le fichier multipart envoyé par Next.js
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 2. Upload dynamique vers Cloudinary
	resp, err := app.cloudinary.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "invoices_logos", // Range tes images dans un dossier propre
	})
	if err != nil {
		return "", err
	}

	// Retourne l'URL sécurisée finale
	return resp.SecureURL, nil
}
