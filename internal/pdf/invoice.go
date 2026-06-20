package pdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/go-pdf/fpdf"
)

// GenerateInvoicePDF orchestre la création complète du document PDF stylisé SaaS premium
func GenerateInvoicePDF(invoice *data.Invoice, businessProfile *data.BusinessProfile) (*bytes.Buffer, error) {
	// Création du document A4 en mm
	pdf := fpdf.New("P", "mm", "A4", "")

	// Configuration automatique du footer global
	setupFooter(pdf, invoice)

	pdf.AddPage()

	// --- DESIGN DE L'ARRIÈRE-PLAN DE LA PAGE ---
	pdf.SetFillColor(247, 249, 252) // Fond gris/bleu très clair, moderne (#F7F9FC)
	pdf.Rect(0, 0, 210, 297, "F")

	// Carte blanche principale conteneur (Effet de feuille surélevée)
	pdf.SetFillColor(255, 255, 255)
	pdf.SetDrawColor(226, 232, 240) // Bordure fine grise (#E2E8F0)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(12, 12, 186, 273, 4, "1234", "DF")

	// Marges intérieures à la carte blanche
	pdf.SetMargins(22, 22, 22)
	pdf.SetAutoPageBreak(true, 30)

	var logoReader io.Reader
	var logoFormat string
	var logoErr error

	if invoice.BusinessLogoURL != "" {
		logoReader, logoFormat, logoErr = fetchImageFromURL(invoice.BusinessLogoURL)
		if logoErr != nil {
			fmt.Printf("Erreur lors de la récupération du logo: %v\n", logoErr)
		} else {
			options := fpdf.ImageOptions{ImageType: logoFormat}
			pdf.RegisterImageOptionsReader("business_logo", options, logoReader)
		}
	}

	// Construction des sections graphiques
	addHeader(pdf, invoice, businessProfile, logoErr)
	addClientSection(pdf, invoice)
	addStatusBanner(pdf, invoice)
	addItemsTable(pdf, invoice)
	addTotalsAndStamp(pdf, invoice)
	addNotes(pdf, invoice)
	addNotesFooter(pdf, businessProfile)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func setupFooter(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	pdf.SetFooterFunc(func() {
		pdf.SetY(-22)
		pdf.SetFont("Arial", "", 8) // Repassage sur Arial standard
		pdf.SetTextColor(160, 174, 192)

		// Ligne de séparation fine
		pdf.SetDrawColor(226, 232, 240)
		pdf.SetLineWidth(0.2)
		pdf.Line(22, pdf.GetY(), 188, pdf.GetY())
		pdf.Ln(4)

		pdf.SetX(22)
		tr := pdf.UnicodeTranslatorFromDescriptor("") // Traducteur natif
		footerText := tr("Ce document est généré automatiquement et constitue une preuve de paiement officielle.")
		pdf.CellFormat(120, 4, footerText, "0", 0, "L", false, 0, "")

		pdf.SetX(148)
	})
}

func addHeader(pdf *fpdf.Fpdf, invoice *data.Invoice, businessProfile *data.BusinessProfile, logoErr error) {
	tr := pdf.UnicodeTranslatorFromDescriptor("") // Traducteur natif

	pdf.SetXY(22, 22)
	if invoice.BusinessLogoURL != "" && logoErr == nil {
		pdf.ImageOptions("business_logo", 22, 22, 35, 0, false, fpdf.ImageOptions{}, 0, "")
	} else {
		pdf.SetFont("Arial", "B", 13)
		pdf.SetTextColor(26, 54, 93)
		pdf.Cell(100, 5, "SILIKIN VILLAGE")
		pdf.Ln(4.5)
		pdf.SetX(22)
		pdf.SetFont("Arial", "", 7.5)
		pdf.SetTextColor(113, 128, 150)
		pdf.Cell(100, 4, "by TEXAF BILEMBO")
	}

	// X=118 + Largeur=70 -> Arrive exactement à la marge droite de 188 mm
	pdf.SetXY(118, 22)
	pdf.SetFont("Arial", "B", 15)
	pdf.SetTextColor(26, 54, 93)

	bName := invoice.BusinessName
	if bName == "" {
		bName = "Acoriss"
	}
	pdf.CellFormat(70, 6, tr(bName), "0", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(74, 85, 104)

	pdf.SetX(118)
	pdf.CellFormat(70, 4.5, tr(fmt.Sprintf("Reçu N° : %s", invoice.InvoiceNumber)), "0", 1, "R", false, 0, "") // [cite: 3, 12]
	pdf.SetX(118)
	pdf.CellFormat(70, 4.5, tr(fmt.Sprintf("Email : %s", businessProfile.Email)), "0", 1, "R", false, 0, "") //
	pdf.SetX(118)
	pdf.CellFormat(70, 4.5, tr(fmt.Sprintf("Tél : %s", businessProfile.Phone)), "0", 1, "R", false, 0, "") //

	pdf.SetY(50)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(26, 54, 93)
	pdf.CellFormat(166, 10, tr("REÇU"), "0", 1, "C", false, 0, "") // [cite: 6]
	pdf.Ln(2)
}

func addClientSection(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	startY := pdf.GetY()

	pdf.SetFillColor(248, 250, 252)
	pdf.SetDrawColor(237, 242, 247)
	pdf.SetLineWidth(0.3)
	pdf.RoundedRect(22, startY, 166, 24, 3, "1234", "DF")

	pdf.SetTextColor(74, 85, 104)

	pdf.SetXY(26, startY+3)
	pdf.SetFont("Arial", "B", 9.5)
	pdf.SetTextColor(26, 54, 93)
	pdf.Cell(80, 5, tr("Détails du Reçu :")) // [cite: 2]

	pdf.SetFont("Arial", "", 9.5)
	pdf.SetTextColor(74, 85, 104)
	pdf.SetXY(26, startY+9)
	pdf.Cell(80, 5, tr(fmt.Sprintf("Reçu N° : %s", invoice.InvoiceNumber))) // [cite: 3, 12]
	pdf.SetXY(26, startY+14)
	pdf.Cell(80, 5, tr(fmt.Sprintf("Client : %s", invoice.ClientName))) // [cite: 4]
	pdf.SetXY(26, startY+19)
	pdf.Cell(80, 5, tr(fmt.Sprintf("Motif : %s", invoice.ClientAddress))) // [cite: 5]

	// Alignement à droite sur l'axe 188 mm (X=118 + Largeur=70 avec option "R")
	pdf.SetXY(118, startY+9)
	pdf.CellFormat(70, 5, tr(fmt.Sprintf("Date : %s", invoice.InvoiceDate.Format("02/01/2006"))), "0", 0, "R", false, 0, "") //
	pdf.SetXY(118, startY+14)
	pdf.CellFormat(70, 5, tr(fmt.Sprintf("Référence : %s", invoice.InvoiceNumber)), "0", 0, "R", false, 0, "") //

	pdf.SetY(startY + 29)
}

func addStatusBanner(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	status := strings.ToUpper(invoice.Status)
	if status == "PAID" || status == "PAYE" || status == "PAYÉ" { // [cite: 11]
		pdf.SetFillColor(240, 253, 244)
		pdf.SetDrawColor(56, 161, 105)
		pdf.Rect(22, pdf.GetY(), 166, 11, "F")

		pdf.SetLineWidth(0.8)
		pdf.Line(22, pdf.GetY(), 22, pdf.GetY()+11)

		pdf.SetXY(26, pdf.GetY()+1.5)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(39, 103, 73)
		pdf.Cell(160, 4, tr("Transaction validée avec succès")) // [cite: 8]

		pdf.SetXY(26, pdf.GetY()+5.5)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(74, 85, 104)
		pdf.Cell(160, 4, tr("Ce reçu confirme le paiement de votre transaction. Conservez-le pour vos archives.")) // [cite: 9]
	} else {
		pdf.SetFillColor(254, 242, 242)
		pdf.SetDrawColor(220, 38, 38)
		pdf.Rect(22, pdf.GetY(), 166, 11, "F")

		pdf.SetLineWidth(0.8)
		pdf.Line(22, pdf.GetY(), 22, pdf.GetY()+11)

		pdf.SetXY(26, pdf.GetY()+3.5)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(153, 27, 27)
		pdf.Cell(160, 4, tr(fmt.Sprintf("Statut actuel : %s. En attente de régularisation.", status)))
	}
	pdf.SetY(pdf.GetY() + 17)
}

func addItemsTable(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetFillColor(26, 54, 93)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9.5)

	pdf.SetX(22)
	pdf.CellFormat(96, 9, "Description", "0", 0, "L", true, 0, "")   //
	pdf.CellFormat(35, 9, "Prix Unitaire", "0", 0, "C", true, 0, "") //
	pdf.CellFormat(35, 9, "Total", "0", 1, "R", true, 0, "")         //

	pdf.SetTextColor(45, 55, 72)
	pdf.SetFont("Arial", "", 9.5)

	for i, item := range invoice.Items {
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(248, 250, 252)
		}

		x, y := 22.0, pdf.GetY()
		pdf.SetXY(x, y)

		uPrice := float64(item.UnitPrice) / 100.0
		lTotal := (float64(item.UnitPrice) * float64(item.Quantity)) / 100.0

		pdf.SetDrawColor(237, 242, 247)
		pdf.SetLineWidth(0.3)
		pdf.MultiCell(96, 8, tr(item.Description), "B", "L", true) //
		nextY := pdf.GetY()

		pdf.SetXY(x+96, y)
		pdf.CellFormat(35, nextY-y, tr(fmt.Sprintf("%.2f %s", uPrice, invoice.Currency)), "B", 0, "C", true, 0, "") //
		pdf.CellFormat(35, nextY-y, tr(fmt.Sprintf("%.2f %s", lTotal, invoice.Currency)), "B", 1, "R", true, 0, "") //

		pdf.SetY(nextY)
	}
}

func addTotalsAndStamp(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.Ln(5)
	currentY := pdf.GetY()

	status := strings.ToUpper(invoice.Status)
	if status == "PAID" || status == "PAYE" || status == "PAYÉ" { //
		pdf.TransformBegin()
		pdf.SetAlpha(0.35, "Normal")
		pdf.SetTextColor(56, 161, 105)
		pdf.SetDrawColor(56, 161, 105)

		centerX := 55.0
		centerY := currentY + 12.0

		// MODIFICATION ICI : Change -10.0 par 0.0 pour le rendre horizontal
		pdf.TransformRotate(0.0, centerX, centerY)

		pdf.SetLineWidth(1.0)
		pdf.RoundedRect(centerX-18, centerY-6, 36, 11, 1.5, "1234", "D")
		pdf.SetLineWidth(0.3)
		pdf.RoundedRect(centerX-16.5, centerY-4.8, 33, 8.6, 1, "1234", "D")

		pdf.SetFont("Arial", "B", 11)
		pdf.SetXY(centerX-18, centerY-4.5)
		pdf.CellFormat(36, 9, tr("PAYÉ"), "0", 0, "C", false, 0, "") //

		pdf.TransformEnd()
		pdf.SetAlpha(1.0, "Normal")
	}

	finalTotal := float64(invoice.TotalAmount) / 100.0

	// Aligné pour finir exactement à X=188 (100 + 50 + 38 = 188)
	pdf.SetXY(100, currentY+4)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(74, 85, 104)
	pdf.CellFormat(50, 8, tr("MONTANT TOTAL :"), "", 0, "R", false, 0, "") // [cite: 13]

	pdf.SetFont("Arial", "B", 15)
	pdf.SetTextColor(56, 161, 105)
	pdf.CellFormat(38, 8, tr(fmt.Sprintf("%.2f %s", finalTotal, invoice.Currency)), "", 1, "R", false, 0, "") //

	pdf.SetY(currentY + 22)
}

func addNotes(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// 1. Titre de remerciement ou Titre de la note personnalisé
	pdf.SetX(22)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(26, 54, 93)

	noteTitle := invoice.NoteTitle
	if invoice.NoteTitle != "" {
		noteTitle = invoice.NoteTitle
	}
	pdf.CellFormat(166, 5, tr(noteTitle), "", 1, "C", false, 0, "") //

	pdf.Ln(2)
	pdf.SetX(22)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(113, 128, 150)

	noteText := invoice.NoteText
	if noteText == "" {
		noteText = invoice.NoteText // Utilisation du texte par défaut si aucun texte personnalisé n'est fourni
	}
	content := fmt.Sprintf("%s\n\n%s", noteText, invoice.NoteText)

	// MultiCell est parfait ici pour éviter que le texte ne dépasse si la note est longue
	pdf.MultiCell(166, 4, tr(content), "", "C", false)
}

func addNotesFooter(pdf *fpdf.Fpdf, businessProfile *data.BusinessProfile) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetX(22)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(26, 54, 93)
	pdf.CellFormat(166, 5, tr("Merci pour votre confiance !"), "", 1, "C", false, 0, "") // [cite: 15]

	pdf.Ln(2)
	pdf.SetX(22)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(113, 128, 150)

	contactPrompt := fmt.Sprintf("Pour toute question ou réclamation, contactez-nous sur : %s  |  %s", businessProfile.Email, businessProfile.Phone) // [cite: 16]
	pdf.CellFormat(166, 4, tr(contactPrompt), "", 1, "C", false, 0, "")
}
