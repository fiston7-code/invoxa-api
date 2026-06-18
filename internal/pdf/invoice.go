package pdf

import (
	"bytes"
	"fmt"
	"strconv"

	"io"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/go-pdf/fpdf"
)

// GenerateInvoicePDF orchestre la création complète du document PDF
func GenerateInvoicePDF(invoice *data.Invoice) (*bytes.Buffer, error) {
	// Configuration initiale : Format A4, Unité : mm, Orientation : Portrait
	pdf := fpdf.New("P", "mm", "A4", "")

	// Configuration du footer automatique (Numérotation des pages et mentions légales)
	setupFooter(pdf, invoice)

	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)
	// Activer le saut de page automatique à 30mm du bas pour ne pas chevaucher le footer
	pdf.SetAutoPageBreak(true, 30)

	var logoReader io.Reader
	var logoFormat string
	var logoErr error

	// Si l'URL du logo est fournie, on tente de la récupérer
	if invoice.BusinessLogoURL != "" {
		logoReader, logoFormat, logoErr = fetchImageFromURL(invoice.BusinessLogoURL)
		if logoErr != nil {
			// En production, on loggue l'erreur mais on NE FAIT PAS planter l'API.
			// La facture sera générée sans logo. C'est de la "dégradation gracieuse".
			fmt.Printf("Erreur lors de la récupération du logo pour la facture %s: %v\n", invoice.InvoiceNumber, logoErr)
		} else {
			// Si récupéré avec succès, on l'enregistre dans le moteur fpdf
			// On définit uniquement le ImageType, sans le champ ReadFromFormat
			options := fpdf.ImageOptions{ImageType: logoFormat}
			pdf.RegisterImageOptionsReader("business_logo", options, logoReader)
		}
	}

	// Construction des différentes sections
	addHeader(pdf, invoice, logoErr)
	addClientSection(pdf, invoice)
	addItemsTable(pdf, invoice)
	addTotals(pdf, invoice)
	addNotes(pdf, invoice)

	// Extraction du flux binaire dans un buffer de mémoire
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// --- SOUS-FONCTIONS DE RENDU ---

// setupFooter définit le comportement automatique du bas de page sur CHAQUE page
func setupFooter(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	pdf.SetFooterFunc(func() {
		// Positionnement à 20 mm du bas
		pdf.SetY(-20)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(120, 120, 120)

		// Ligne séparatrice discrète
		pdf.SetDrawColor(220, 220, 220)
		pdf.SetLineWidth(0.3)
		pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
		pdf.Ln(2)

		// Construction de la chaîne de contact du footer
		var contactInfo string
		if invoice.FooterAddress != "" {
			contactInfo += invoice.FooterAddress
		}
		if invoice.FooterPhone != "" {
			if contactInfo != "" {
				contactInfo += "  |  "
			}
			contactInfo += fmt.Sprintf("Tel: %s", invoice.FooterPhone)
		}
		if invoice.FooterEmail != "" {
			if contactInfo != "" {
				contactInfo += "  |  "
			}
			contactInfo += fmt.Sprintf("Email: %s", invoice.FooterEmail)
		}

		// Si aucun footer personnalisé n'est fourni, on met une valeur par défaut pro
		if contactInfo == "" {
			contactInfo = fmt.Sprintf("%s - Merci pour votre confiance.", invoice.BusinessName)
		}

		// Affichage des infos à gauche et du numéro de page à droite
		pdf.CellFormat(130, 6, contactInfo, "0", 0, "L", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("Page %d", pdf.PageNo()), "0", 0, "R", false, 0, "")
	})
}

// addHeader gère l'identité visuelle de l'émetteur (Logo ou Texte) et le statut
func addHeader(pdf *fpdf.Fpdf, invoice *data.Invoice, logoErr error) {
	// 1. Initialisation : Si logo, on décale le texte vers le bas
	textY := 20.0

	// Si logo présent et aucune erreur de téléchargement
	if invoice.BusinessLogoURL != "" && logoErr == nil {
		// Le logo occupe environ 30-35mm en hauteur
		// On utilise des options vides {} pour éviter l'erreur sur le champ inconnu
		pdf.ImageOptions("business_logo", 20, 15, 35, 0, false, fpdf.ImageOptions{}, 0, "")
		textY = 40.0 // Le texte commence plus bas
	}

	// 2. Nom de l'entreprise
	pdf.SetXY(20, textY)
	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(59, 30, 93)

	businessName := invoice.BusinessName
	if businessName == "" {
		businessName = "KALK"
	}
	pdf.Cell(100, 10, businessName)

	// 3. Titre "INVOICE" (Fixe en haut à droite)
	pdf.SetFont("Arial", "B", 26)
	pdf.MoveTo(140, 20)
	pdf.CellFormat(50, 10, "RECU", "0", 0, "R", false, 0, "")

	// 4. Métadonnées (Fixe en haut à droite)
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.MoveTo(140, 31)
	pdf.CellFormat(50, 5, fmt.Sprintf("N#: %s", invoice.InvoiceNumber), "0", 0, "R", false, 0, "")
	pdf.MoveTo(140, 36)
	pdf.CellFormat(50, 5, fmt.Sprintf("Date: %s", invoice.InvoiceDate.Format("2006-01-02")), "0", 0, "R", false, 0, "")

	// 5. Statut
	if invoice.Status != "" {
		pdf.MoveTo(140, 42)
		pdf.SetFont("Arial", "B", 9)
		if invoice.Status == "PAID" || invoice.Status == "payé" {
			pdf.SetTextColor(40, 167, 69)
		} else {
			pdf.SetTextColor(220, 53, 69)
		}
		pdf.CellFormat(50, 5, fmt.Sprintf("STATUS: %s", invoice.Status), "0", 0, "R", false, 0, "")
	}

	// 6. Code RCCM (Positionné dynamiquement sous le nom de l'entreprise)
	if invoice.BusinessRCCM != "" {
		pdf.MoveTo(20, textY+10)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(120, 120, 120)
		pdf.MultiCell(90, 4, fmt.Sprintf("RCCM: %s", invoice.BusinessRCCM), "", "L", false)
	}

	// On force le curseur pour la suite du document
	pdf.SetY(85)
}

// addClientSection affiche les coordonnées complètes du client (Billed To)
func addClientSection(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	// Au lieu de fixer Y à 55, on récupère la position actuelle du curseur
	// Cela garantit que la section client s'affiche toujours après l'en-tête
	startY := pdf.GetY()
	if startY < 85 {
		startY = 85
	} // On s'assure d'un minimum d'espace

	// Ligne décorative verticale
	pdf.SetDrawColor(123, 75, 183)
	pdf.SetLineWidth(0.6)
	pdf.Line(20, startY, 20, startY+27)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(123, 75, 183)
	pdf.MoveTo(24, startY)
	pdf.Cell(100, 4, "BILLED TO:")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(59, 30, 93)
	pdf.MoveTo(24, 60)
	pdf.Cell(100, 5, invoice.ClientName)
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(80, 80, 80)
	currentY := 66.0

	if invoice.ClientPhone != "" {
		pdf.MoveTo(24, currentY)
		pdf.Cell(100, 4, fmt.Sprintf("Phone: %s", invoice.ClientPhone))
		currentY += 4.5
	}
	if invoice.ClientEmail != "" {
		pdf.MoveTo(24, currentY)
		pdf.Cell(100, 4, fmt.Sprintf("Email: %s", invoice.ClientEmail))
		currentY += 4.5
	}
	if invoice.ClientAddress != "" {
		pdf.MoveTo(24, currentY)
		pdf.Cell(100, 4, fmt.Sprintf("Address: %s", invoice.ClientAddress))
	}
}

// addItemsTable construit la grille dynamique des items
func addItemsTable(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	pdf.SetY(92)
	pdf.SetFillColor(59, 30, 93)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 10)

	// En-têtes
	pdf.CellFormat(110, 8, "Description", "0", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "QTY", "0", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "AMOUNT", "0", 1, "R", true, 0, "")

	// Lignes du tableau
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Arial", "", 10)

	for _, item := range invoice.Items {
		// RAPPEL : UnitPrice est stocké en centimes (int).
		// Le montant total de la ligne = (UnitPrice * Quantité) / 100.0
		lineTotal := float64(int64(item.UnitPrice)*int64(item.Quantity)) / 100.0

		x, y := pdf.GetX(), pdf.GetY()

		// Rendu de la description avec retour à la ligne automatique
		pdf.MultiCell(110, 6, item.Description, "B", "L", false)
		nextY := pdf.GetY()

		// Alignement horizontal des cellules suivantes
		pdf.SetXY(x+110, y)
		pdf.CellFormat(25, nextY-y, strconv.Itoa(item.Quantity), "B", 0, "C", false, 0, "")
		pdf.CellFormat(35, nextY-y, fmt.Sprintf("%.2f %s", lineTotal, invoice.Currency), "B", 1, "R", false, 0, "")

		pdf.SetY(nextY)
	}
}

// addTotals affiche la ligne finale de facturation convertie depuis les centimes
func addTotals(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	pdf.Ln(4)

	// Conversion propre des centimes (int) en float64 pour l'affichage décimal
	finalTotal := float64(invoice.TotalAmount) / 100.0

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(123, 75, 183)
	pdf.CellFormat(135, 8, "Total TTC", "", 0, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(59, 30, 93)
	pdf.CellFormat(35, 8, fmt.Sprintf("%.2f %s", finalTotal, invoice.Currency), "", 1, "R", false, 0, "")
}

// addNotes ajoute le bloc d'instructions de paiement ou notes optionnelles
func addNotes(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	if invoice.NoteTitle == "" && invoice.NoteText == "" {
		return
	}

	pdf.Ln(8)
	pdf.SetFillColor(245, 242, 239) // Fond beige clair texturé
	pdf.SetDrawColor(231, 217, 196)

	currentX, currentY := pdf.GetX(), pdf.GetY()

	// Dessine l'encadré
	pdf.Rect(currentX, currentY, 170, 22, "DF")

	pdf.SetTextColor(50, 50, 50)
	if invoice.NoteTitle != "" {
		pdf.SetFont("Arial", "B", 9)
		pdf.MoveTo(currentX+4, currentY+4)
		pdf.Cell(160, 4, invoice.NoteTitle)
		currentY += 4
	}
	if invoice.NoteText != "" {
		pdf.SetFont("Arial", "I", 9)
		pdf.MoveTo(currentX+4, currentY+5)
		pdf.MultiCell(160, 4, invoice.NoteText, "", "L", false)
	}
}
