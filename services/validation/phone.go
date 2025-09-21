package validation

import (
	"errors"
	"regexp"
	"strings"
)

// PhoneValidator gère la validation des numéros de téléphone
type PhoneValidator struct {
	// Regex pour différents formats de numéros de téléphone
	internationalRegex *regexp.Regexp
	localRegex         *regexp.Regexp
	ivoirianRegex      *regexp.Regexp
}

// NewPhoneValidator crée une nouvelle instance du validateur de téléphone
func NewPhoneValidator() *PhoneValidator {
	return &PhoneValidator{
		// Format international : +225XXXXXXXX (Côte d'Ivoire)
		internationalRegex: regexp.MustCompile(`^\+225\d{8,10}$`),
		// Format local : 0X XX XX XX XX ou 0XXXXXXXXX
		localRegex: regexp.MustCompile(`^0\d{8,9}$`),
		// Format ivoirien spécifique : +225 suivi de 07, 05, 01, etc.
		ivoirianRegex: regexp.MustCompile(`^\+225(0[157]|07|05|01)\d{7,8}$`),
	}
}

// ValidatePhone valide un numéro de téléphone selon les règles ivoiriennes
func (pv *PhoneValidator) ValidatePhone(phone string) error {
	if phone == "" {
		return errors.New("numéro de téléphone requis")
	}

	// Supprimer les espaces et tirets
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	// Vérifier la longueur minimale et maximale
	if len(cleanPhone) < 8 || len(cleanPhone) > 15 {
		return errors.New("numéro de téléphone doit contenir entre 8 et 15 caractères")
	}

	// Vérifier si c'est un format international valide
	if strings.HasPrefix(cleanPhone, "+") {
		return pv.validateInternationalPhone(cleanPhone)
	}

	// Vérifier si c'est un format local valide
	if strings.HasPrefix(cleanPhone, "0") {
		return pv.validateLocalPhone(cleanPhone)
	}

	// Si ce n'est ni international ni local, c'est invalide
	return errors.New("format de numéro de téléphone invalide")
}

// validateInternationalPhone valide un numéro international
func (pv *PhoneValidator) validateInternationalPhone(phone string) error {
	// Vérifier le format général international
	if !pv.internationalRegex.MatchString(phone) {
		return errors.New("format international invalide. Format attendu: +225XXXXXXXX")
	}

	// Vérification spécifique pour la Côte d'Ivoire
	if strings.HasPrefix(phone, "+225") {
		// Extraire la partie après +225
		localPart := phone[4:] // Enlever "+225"
		
		// Vérifier que la partie locale est valide
		if len(localPart) < 8 || len(localPart) > 10 {
			return errors.New("numéro ivoirien invalide après l'indicatif +225")
		}

		// Vérifier les préfixes d'opérateurs ivoiriens
		validPrefixes := []string{"07", "05", "01", "08", "09"}
		isValidPrefix := false
		
		for _, prefix := range validPrefixes {
			if strings.HasPrefix(localPart, prefix) {
				isValidPrefix = true
				break
			}
		}

		if !isValidPrefix {
			return errors.New("préfixe d'opérateur ivoirien invalide (doit commencer par 01, 05, 07, 08, ou 09)")
		}
	}

	return nil
}

// validateLocalPhone valide un numéro local
func (pv *PhoneValidator) validateLocalPhone(phone string) error {
	if !pv.localRegex.MatchString(phone) {
		return errors.New("format local invalide. Format attendu: 0XXXXXXXX")
	}

	// Vérifier les préfixes locaux valides (01, 05, 07, 08, 09)
	validLocalPrefixes := []string{"001", "005", "007", "008", "009"}
	isValidLocalPrefix := false
	
	for _, prefix := range validLocalPrefixes {
		if strings.HasPrefix(phone, prefix) {
			isValidLocalPrefix = true
			break
		}
	}

	if !isValidLocalPrefix {
		return errors.New("préfixe local invalide (doit commencer par 001, 005, 007, 008, ou 009)")
	}

	return nil
}

// NormalizePhone normalise un numéro de téléphone au format international
func (pv *PhoneValidator) NormalizePhone(phone string) (string, error) {
	// D'abord valider
	if err := pv.ValidatePhone(phone); err != nil {
		return "", err
	}

	// Supprimer les espaces et tirets
	cleanPhone := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")

	// Si c'est déjà international, retourner tel quel
	if strings.HasPrefix(cleanPhone, "+225") {
		return cleanPhone, nil
	}

	// Si c'est local (commence par 0), convertir en international
	if strings.HasPrefix(cleanPhone, "0") {
		// Remplacer le 0 par +225
		return "+225" + cleanPhone[1:], nil
	}

	return cleanPhone, nil
}

// IsValidPhoneFormat vérifie rapidement si le format est valide sans validation complète
func (pv *PhoneValidator) IsValidPhoneFormat(phone string) bool {
	return pv.ValidatePhone(phone) == nil
}