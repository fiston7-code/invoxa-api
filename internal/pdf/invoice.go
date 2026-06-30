package pdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fiston7-code/invoxa-api/internal/data"
	"github.com/go-pdf/fpdf"
)

// GenerateInvoicePDF orchestre la création complète du document PDF stylisé SaaS premium
func GenerateInvoicePDF(invoice *data.Invoice, businessProfile *data.BusinessProfile) (*bytes.Buffer, error) {
	// Création du document A4 en mm
	pdf := fpdf.New("P", "mm", "A4", "")

	// Configuration automatique du footer global
	setupFooter(pdf, invoice, businessProfile)

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

	// ✅ DYNAMIQUE: Récupère le logo du business profile
	if businessProfile.LogoURL != "" {
		logoReader, logoFormat, logoErr = fetchImageFromURL(businessProfile.LogoURL)
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

func setupFooter(pdf *fpdf.Fpdf, invoice *data.Invoice, businessProfile *data.BusinessProfile) {
	pdf.SetFooterFunc(func() {
		pdf.SetY(-22)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(160, 174, 192)

		// Ligne de séparation fine
		pdf.SetDrawColor(226, 232, 240)
		pdf.SetLineWidth(0.2)
		pdf.Line(22, pdf.GetY(), 188, pdf.GetY())
		pdf.Ln(4)

		pdf.SetX(22)
		tr := pdf.UnicodeTranslatorFromDescriptor("")
		footerText := tr("Ce document est généré automatiquement et constitue une preuve de paiement officielle.")
		pdf.CellFormat(120, 4, footerText, "0", 0, "L", false, 0, "")

		pdf.SetX(148)
	})
}

// func addHeader(pdf *fpdf.Fpdf, invoice *data.Invoice, businessProfile *data.BusinessProfile, logoErr error) {
// 	tr := pdf.UnicodeTranslatorFromDescriptor("")

// 	pdf.SetXY(22, 22)

// 	// ✅ DYNAMIQUE: Logo du business profile
// 	if businessProfile.LogoURL != "" && logoErr == nil {
// 		pdf.ImageOptions("business_logo", 22, 22, 35, 0, false, fpdf.ImageOptions{}, 0, "")
// 	} else {
// 		// ✅ Fallback: Initialles du nom du business
// 		pdf.SetFont("Arial", "B", 13)
// 		pdf.SetTextColor(26, 54, 93)

// 		initials := ""
// 		if len(businessProfile.Name) > 0 {
// 			initials = strings.ToUpper(string(businessProfile.Name[0]))
// 			if len(businessProfile.Name) > 1 {
// 				for i := 1; i < len(businessProfile.Name); i++ {
// 					if businessProfile.Name[i] == ' ' && i+1 < len(businessProfile.Name) {
// 						initials += strings.ToUpper(string(businessProfile.Name[i+1]))
// 						break
// 					}
// 				}
// 			}
// 		}

// 		pdf.Cell(100, 5, initials)
// 		pdf.Ln(4.5)
// 		pdf.SetX(22)
// 		pdf.SetFont("Arial", "", 7.5)
// 		pdf.SetTextColor(113, 128, 150)
// 		pdf.Cell(100, 4, tr(businessProfile.Name)) // ✅ DYNAMIQUE
// 	}

// 	// ✅ DYNAMIQUE: Infos du business à droite
// 	pdf.SetXY(118, 22)
// 	pdf.SetFont("Arial", "B", 15)
// 	pdf.SetTextColor(26, 54, 93)
// 	pdf.CellFormat(70, 6, tr(businessProfile.Name), "0", 1, "R", false, 0, "")

// 	pdf.SetFont("Arial", "", 8.5)
// 	pdf.SetTextColor(74, 85, 104)

// 	pdf.SetX(118)
// 	// ✅ DYNAMIQUE: Numéro de reçu
// 	pdf.CellFormat(70, 4.5, tr(fmt.Sprintf("Reçu N° : %s", invoice.InvoiceNumber)), "0", 1, "R", false, 0, "")
// 	pdf.SetX(118)
// 	pdf.CellFormat(70, 4.5, tr(fmt.Sprintf("Email : %s", businessProfile.Email)), "0", 1, "R", false, 0, "")
// 	pdf.SetX(118)
// 	pdf.CellFormat(70, 4.5, tr(fmt.Sprintf("Tél : %s", businessProfile.Phone)), "0", 1, "R", false, 0, "")

// 	pdf.SetY(50)
// 	pdf.SetFont("Arial", "B", 20)
// 	pdf.SetTextColor(26, 54, 93)
// 	pdf.CellFormat(166, 10, tr("REÇU"), "0", 1, "C", false, 0, "")
// 	pdf.Ln(2)
// }

func addHeader(pdf *fpdf.Fpdf, invoice *data.Invoice, businessProfile *data.BusinessProfile, logoErr error) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.SetXY(22, 22)

	// ✅ DYNAMIQUE: Logo du business profile
	if businessProfile.LogoURL != "" && logoErr == nil {
		pdf.ImageOptions("business_logo", 22, 22, 35, 0, false, fpdf.ImageOptions{}, 0, "")
	} else {
		// ✅ Fallback: Initialles du nom du business
		pdf.SetFont("Arial", "B", 13)
		pdf.SetTextColor(26, 54, 93)

		initials := ""
		if len(businessProfile.Name) > 0 {
			initials = strings.ToUpper(string(businessProfile.Name[0]))
			if len(businessProfile.Name) > 1 {
				for i := 1; i < len(businessProfile.Name); i++ {
					if businessProfile.Name[i] == ' ' && i+1 < len(businessProfile.Name) {
						initials += strings.ToUpper(string(businessProfile.Name[i+1]))
						break
					}
				}
			}
		}

		pdf.Cell(100, 5, initials)
		pdf.Ln(4.5)
		pdf.SetX(22)
		pdf.SetFont("Arial", "", 7.5)
		pdf.SetTextColor(113, 128, 150)

		businessName := businessProfile.Name
		if businessName == "" {
			businessName = "Entreprise"
		}
		pdf.Cell(100, 4, tr(businessName))
	}

	// ✅ CORRIGÉ: Infos du business à droite avec meilleur spacing
	pdf.SetXY(118, 22)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(26, 54, 93)

	businessName := businessProfile.Name
	if businessName == "" {
		businessName = "Entreprise"
	}
	pdf.CellFormat(68, 6, tr(businessName), "0", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(74, 85, 104)

	// Numéro de reçu
	pdf.SetXY(118, 30)
	pdf.CellFormat(68, 4, tr(fmt.Sprintf("Reçu N° : %s", invoice.InvoiceNumber)), "0", 1, "R", false, 0, "")

	// Email
	pdf.SetXY(118, 34.5)
	email := businessProfile.Email
	if email == "" {
		email = "—"
	}
	pdf.CellFormat(68, 4, tr(fmt.Sprintf("Email : %s", businessProfile.Email)), "0", 1, "R", false, 0, "")

	// Téléphone
	pdf.SetXY(118, 39)
	phone := businessProfile.Phone
	if phone == "" {
		phone = "—"
	}
	pdf.CellFormat(68, 4, tr(fmt.Sprintf("Tél : %s", businessProfile.Phone)), "0", 1, "R", false, 0, "")

	pdf.SetY(50)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(26, 54, 93)
	pdf.CellFormat(166, 10, tr("REÇU"), "0", 1, "C", false, 0, "")
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
	pdf.Cell(80, 5, tr("Détails du Reçu :"))

	pdf.SetFont("Arial", "", 9.5)
	pdf.SetTextColor(74, 85, 104)
	pdf.SetXY(26, startY+9)
	// ✅ DYNAMIQUE: Tous les champs du client
	pdf.Cell(80, 5, tr(fmt.Sprintf("Reçu N° : %s", invoice.InvoiceNumber)))
	pdf.SetXY(26, startY+14)
	pdf.Cell(80, 5, tr(fmt.Sprintf("Client : %s", invoice.ClientName)))
	pdf.SetXY(26, startY+19)
	pdf.Cell(80, 5, tr(fmt.Sprintf("Adresse : %s", invoice.ClientAddress)))

	// À droite: Date et références
	pdf.SetXY(118, startY+9)
	pdf.CellFormat(70, 5, tr(fmt.Sprintf("Date : %s", invoice.InvoiceDate.Format("02/01/2006"))), "0", 0, "R", false, 0, "")
	pdf.SetXY(118, startY+14)
	pdf.CellFormat(70, 5, tr(fmt.Sprintf("Statut : %s", strings.ToUpper(invoice.Status))), "0", 0, "R", false, 0, "")

	if invoice.ClientEmail != "" {
		pdf.SetXY(118, startY+19)
		pdf.CellFormat(70, 5, tr(fmt.Sprintf("Email : %s", invoice.ClientEmail)), "0", 0, "R", false, 0, "")
	}

	pdf.SetY(startY + 29)
}

func addStatusBanner(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	status := strings.ToUpper(invoice.Status)

	// ✅ DYNAMIQUE: Statut dynamique du reçu
	if status == "PAID" || status == "PAYE" || status == "PAYÉ" {
		pdf.SetFillColor(240, 253, 244)
		pdf.SetDrawColor(56, 161, 105)
		pdf.Rect(22, pdf.GetY(), 166, 11, "F")

		pdf.SetLineWidth(0.8)
		pdf.Line(22, pdf.GetY(), 22, pdf.GetY()+11)

		pdf.SetXY(26, pdf.GetY()+1.5)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(39, 103, 73)
		pdf.Cell(160, 4, tr("✓ Transaction validée avec succès"))

		pdf.SetXY(26, pdf.GetY()+5.5)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(74, 85, 104)
		pdf.Cell(160, 4, tr("Ce reçu confirme le paiement de votre transaction. Conservez-le pour vos archives."))
	} else if status == "SENT" || status == "ENVOYÉ" {
		pdf.SetFillColor(254, 249, 240)
		pdf.SetDrawColor(217, 119, 6)
		pdf.Rect(22, pdf.GetY(), 166, 11, "F")

		pdf.SetLineWidth(0.8)
		pdf.Line(22, pdf.GetY(), 22, pdf.GetY()+11)

		pdf.SetXY(26, pdf.GetY()+1.5)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(120, 53, 15)
		pdf.Cell(160, 4, tr("⏳ Reçu envoyé - En attente de paiement"))

		pdf.SetXY(26, pdf.GetY()+5.5)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(74, 85, 104)
		pdf.Cell(160, 4, tr("Veuillez procéder au paiement dans les délais impartis."))
	} else {
		pdf.SetFillColor(254, 242, 242)
		pdf.SetDrawColor(220, 38, 38)
		pdf.Rect(22, pdf.GetY(), 166, 11, "F")

		pdf.SetLineWidth(0.8)
		pdf.Line(22, pdf.GetY(), 22, pdf.GetY()+11)

		pdf.SetXY(26, pdf.GetY()+3.5)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(153, 27, 27)
		pdf.Cell(160, 4, tr(fmt.Sprintf("Statut : %s", status)))
	}
	pdf.SetY(pdf.GetY() + 17)
}

func addItemsTable(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetFillColor(26, 54, 93)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9.5)

	pdf.SetX(22)
	pdf.CellFormat(96, 9, "Description", "0", 0, "L", true, 0, "")
	pdf.CellFormat(35, 9, "Prix Unitaire", "0", 0, "C", true, 0, "")
	pdf.CellFormat(35, 9, "Total", "0", 1, "R", true, 0, "")

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

		// ✅ DYNAMIQUE: Calcul automatique des prix
		uPrice := float64(item.UnitPrice) / 100.0
		lTotal := (float64(item.UnitPrice) * float64(item.Quantity)) / 100.0

		pdf.SetDrawColor(237, 242, 247)
		pdf.SetLineWidth(0.3)
		pdf.MultiCell(96, 8, tr(item.Description), "B", "L", true)
		nextY := pdf.GetY()

		pdf.SetXY(x+96, y)
		pdf.CellFormat(35, nextY-y, tr(fmt.Sprintf("%.2f %s", uPrice, invoice.Currency)), "B", 0, "C", true, 0, "")
		pdf.CellFormat(35, nextY-y, tr(fmt.Sprintf("%.2f %s", lTotal, invoice.Currency)), "B", 1, "R", true, 0, "")

		pdf.SetY(nextY)
	}
}

func addTotalsAndStamp(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.Ln(5)
	currentY := pdf.GetY()

	status := strings.ToUpper(invoice.Status)

	// ✅ DYNAMIQUE: Affiche le tampon PAYÉ si le statut est paid
	if status == "PAID" || status == "PAYE" || status == "PAYÉ" {
		pdf.TransformBegin()
		pdf.SetAlpha(0.35, "Normal")
		pdf.SetTextColor(56, 161, 105)
		pdf.SetDrawColor(56, 161, 105)

		centerX := 55.0
		centerY := currentY + 12.0

		pdf.TransformRotate(0.0, centerX, centerY)

		pdf.SetLineWidth(1.0)
		pdf.RoundedRect(centerX-18, centerY-6, 36, 11, 1.5, "1234", "D")
		pdf.SetLineWidth(0.3)
		pdf.RoundedRect(centerX-16.5, centerY-4.8, 33, 8.6, 1, "1234", "D")

		pdf.SetFont("Arial", "B", 11)
		pdf.SetXY(centerX-18, centerY-4.5)
		pdf.CellFormat(36, 9, tr("PAYÉ"), "0", 0, "C", false, 0, "")

		pdf.TransformEnd()
		pdf.SetAlpha(1.0, "Normal")
	}

	// ✅ DYNAMIQUE: Montant total
	finalTotal := float64(invoice.TotalAmount) / 100.0

	pdf.SetXY(100, currentY+4)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(74, 85, 104)
	pdf.CellFormat(50, 8, tr("MONTANT TOTAL :"), "", 0, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 15)
	pdf.SetTextColor(56, 161, 105)
	pdf.CellFormat(38, 8, tr(fmt.Sprintf("%.2f %s", finalTotal, invoice.Currency)), "", 1, "R", false, 0, "")

	pdf.SetY(currentY + 22)
}

func addNotes(pdf *fpdf.Fpdf, invoice *data.Invoice) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	if invoice.NoteTitle == "" && invoice.NoteText == "" {
		return
	}

	// ✅ DYNAMIQUE: Titre de la note
	pdf.SetX(22)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(26, 54, 93)

	noteTitle := invoice.NoteTitle
	if noteTitle == "" {
		noteTitle = "Notes supplémentaires"
	}
	pdf.CellFormat(166, 5, tr(noteTitle), "", 1, "C", false, 0, "")

	pdf.Ln(2)
	pdf.SetX(22)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(113, 128, 150)

	// ✅ DYNAMIQUE: Texte de la note
	if invoice.NoteText != "" {
		pdf.MultiCell(166, 4, tr(invoice.NoteText), "", "C", false)
	}
}

func addNotesFooter(pdf *fpdf.Fpdf, businessProfile *data.BusinessProfile) {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.SetX(22)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(26, 54, 93)
	pdf.CellFormat(166, 5, tr("Merci pour votre confiance !"), "", 1, "C", false, 0, "")

	pdf.Ln(2)
	pdf.SetX(22)
	pdf.SetFont("Arial", "", 8.5)
	pdf.SetTextColor(113, 128, 150)

	// ✅ DYNAMIQUE: Coordonnées du business
	contactPrompt := fmt.Sprintf("Pour toute question, contactez-nous : %s  |  %s  |  %s",
		businessProfile.Email, businessProfile.Phone, businessProfile.Address)
	pdf.CellFormat(166, 4, tr(contactPrompt), "", 1, "C", false, 0, "")

	// ✅ NOUVEAU: Date de génération
	pdf.Ln(2)
	pdf.SetX(22)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(160, 174, 192)
	pdf.CellFormat(166, 3, tr(fmt.Sprintf("Généré le : %s", time.Now().Format("02/01/2006 15:04"))), "", 1, "C", false, 0, "")
}
