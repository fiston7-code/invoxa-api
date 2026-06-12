package validator

import (
	"regexp"
	"slices"
)

var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// Validator contient la map des erreurs de validation (ex: "client_name": "must be provided")
type Validator struct {
	Errors map[string]string
}

// New crée une nouvelle instance de Validator avec une map d'erreurs vide.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid renvoie true si aucune erreur n'a été détectée.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError ajoute un message d'erreur à la map si la clé n'existe pas encore.
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check ajoute une erreur uniquement si la condition 'ok' est fausse (false).
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue vérifie de manière générique si une valeur fait partie des choix autorisés.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Matches renvoie true si une chaîne correspond à l'expression régulière donnée.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique renvoie true si toutes les valeurs du tableau (slice) sont uniques.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
