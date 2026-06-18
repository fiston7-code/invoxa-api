package pdf

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// fetchImageFromURL télécharge une image distante et retourne un Reader prêt pour fpdf.
// Il gère également la détection du format (PNG/JPG).
func fetchImageFromURL(url string) (io.Reader, string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 1. Création de la requête avec un User-Agent pour passer les blocages 403
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("erreur de création de requête: %w", err)
	}

	// On simule un navigateur pour éviter les erreurs 403 Forbidden
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Safari/537.36")

	// 2. Exécution de la requête
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	// 3. Vérification du statut
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("impossible de récupérer l'image, statut: %d", resp.StatusCode)
	}

	// 4. Détection du format
	contentType := resp.Header.Get("Content-Type")
	var format string
	switch contentType {
	case "image/png":
		format = "PNG"
	case "image/jpeg", "image/jpg":
		format = "JPG"
	default:
		return nil, "", fmt.Errorf("format d'image non supporté: %s", contentType)
	}

	// 5. Lecture des données
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("erreur lors de la lecture des données d'image: %w", err)
	}

	return bytes.NewReader(data), format, nil
}
