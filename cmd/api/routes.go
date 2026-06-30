package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Configurer les gestionnaires d'erreurs personnalisés
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// --- ROUTES PUBLIQUES ---
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// --- ROUTES PRIVÉES : FACTURES (Protégées par Permissions) ---
	// Lecture (Besoin de invoices:read)
	router.HandlerFunc(http.MethodGet, "/v1/invoices", app.requirePermission("invoices:read", app.listInvoicesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id", app.requirePermission("invoices:read", app.showInvoiceHandler))
	router.HandlerFunc(http.MethodGet, "/v1/invoices/:id/pdf", app.requirePermission("invoices:read", app.downloadInvoicePDFHandler))

	// Écritures / Modifications / Suppressions (Alignés sur invoices:write en BDD)
	router.HandlerFunc(http.MethodPost, "/v1/invoices", app.requirePermission("invoices:write", app.createInvoiceHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/invoices/:id", app.requirePermission("invoices:write", app.updateInvoiceHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/invoices/:id", app.requirePermission("invoices:write", app.deleteInvoiceHandler))

	// --- ROUTES PRIVÉES : BUSINESS (Sécurisées au moins par Authentification + Activation) ---
	// router.HandlerFunc(http.MethodGet, "/v1/business/:id", app.requireActivatedUser(app.getBusinessProfileHandler))
	// router.HandlerFunc(http.MethodPost, "/v1/business", app.requireActivatedUser(app.createBusinessProfileHandler))
	// router.HandlerFunc(http.MethodPatch, "/v1/business/:id", app.requireActivatedUser(app.updateBusinessProfileHandler))

	// Dans routes.go
	// ✅ CORRECT: GET sans :id pour récupérer LE profil de l'user connecté
	router.HandlerFunc(http.MethodGet, "/v1/business", app.requireActivatedUser(app.getBusinessProfileHandler))

	// POST pour créer
	router.HandlerFunc(http.MethodPost, "/v1/business", app.requireActivatedUser(app.createBusinessProfileHandler))

	// PATCH avec :id pour modifier un profil spécifique (optionnel)
	router.HandlerFunc(http.MethodPatch, "/v1/business/:id", app.requireActivatedUser(app.updateBusinessProfileHandler))
	// Remplace HandlerFunc par Handler
	// Dans routes.go
	router.Handler(http.MethodPost, "/v1/test-upload", app.requireActivatedUser(http.HandlerFunc(app.testUploadHandler)))

	// Chaîne globale des middlewares applicatifs
	return app.recoverPanic(app.rateLimit(app.enableCORS(app.authenticate(router))))
}
