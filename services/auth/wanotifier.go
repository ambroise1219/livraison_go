package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ambroise1219/livraison_go/config"
)

type WanotifierService struct {
	webhookURL string
	httpClient *http.Client
}

func NewWanotifierService(cfg *config.Config) *WanotifierService {
	return &WanotifierService{
		webhookURL: cfg.WanotifierWebhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type wanotifierPayload struct {
	Data struct {
		BodyVariables []string `json:"body_variables"`
	} `json:"data"`
	Recipients []struct {
		WhatsAppNumber string                 `json:"whatsapp_number"`
		FirstName      string                 `json:"first_name,omitempty"`
		LastName       string                 `json:"last_name,omitempty"`
		Attributes     map[string]interface{} `json:"attributes,omitempty"`
		Lists          []string               `json:"lists,omitempty"`
		Tags           []string               `json:"tags,omitempty"`
		Replace        bool                   `json:"replace,omitempty"`
	} `json:"recipients"`
}

func (w *WanotifierService) SendOTP(phone, code, firstName, lastName string) error {
	if w.webhookURL == "" {
		return fmt.Errorf("WANOTIFIER_WEBHOOK_URL non configuré")
	}
	payload := wanotifierPayload{}
	payload.Data.BodyVariables = []string{code}
	recipient := struct {
		WhatsAppNumber string                 `json:"whatsapp_number"`
		FirstName      string                 `json:"first_name,omitempty"`
		LastName       string                 `json:"last_name,omitempty"`
		Attributes     map[string]interface{} `json:"attributes,omitempty"`
		Lists          []string               `json:"lists,omitempty"`
		Tags           []string               `json:"tags,omitempty"`
		Replace        bool                   `json:"replace,omitempty"`
	}{
		WhatsAppNumber: phone,
		FirstName:      firstName,
		LastName:       lastName,
		Lists:          []string{"OTP"},
		Tags:           []string{"otp_verification", "auth"},
		Replace:        false,
	}
	payload.Recipients = []struct {
		WhatsAppNumber string                 `json:"whatsapp_number"`
		FirstName      string                 `json:"first_name,omitempty"`
		LastName       string                 `json:"last_name,omitempty"`
		Attributes     map[string]interface{} `json:"attributes,omitempty"`
		Lists          []string               `json:"lists,omitempty"`
		Tags           []string               `json:"tags,omitempty"`
		Replace        bool                   `json:"replace,omitempty"`
	}{recipient}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, w.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("wanotifier status %d: %s", resp.StatusCode, string(b))
	}
	// optional debug logs
	if config.GetConfig().WanotifierDebug {
		fmt.Printf("wanotifier ok: %s\n", string(b))
	}
	return nil
}
