package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type CloudPaymentsProvider struct {
	publicID  string
	apiSecret string
	client    *http.Client
}

func NewCloudPaymentsProvider(publicID, apiSecret string) *CloudPaymentsProvider {
	return &CloudPaymentsProvider{
		publicID:  publicID,
		apiSecret: apiSecret,
		client:    &http.Client{},
	}
}

type CloudPaymentsRequest struct {
	Amount    float64 `json:"Amount"`
	Currency  string  `json:"Currency"`
	PaymentID int     `json:"InvoiceId"`
	UserID    int64   `json:"AccountId"`
	Email     string  `json:"Email,omitempty"`
	Phone     string  `json:"Phone,omitempty"`
}

type CloudPaymentsResponse struct {
	Success       bool   `json:"Success"`
	TransactionID string `json:"TransactionId"`
	PaymentURL    string `json:"Model"`
	Message       string `json:"Message"`
}

type CloudPaymentsWebhook struct {
	TransactionID string  `json:"TransactionId"`
	Amount        float64 `json:"Amount"`
	Currency      string  `json:"Currency"`
	Status        string  `json:"Status"`
	InvoiceId     string  `json:"InvoiceId"`
	AccountId     string  `json:"AccountId"`
	Email         string  `json:"Email"`
	DateTime      string  `json:"DateTime"`
}

func (p *CloudPaymentsProvider) CreatePayment(ctx context.Context, req CloudPaymentsRequest) (*CloudPaymentsResponse, error) {
	// Создаем платеж через CloudPayments API
	requestData := map[string]interface{}{
		"Amount":      req.Amount,
		"Currency":    req.Currency,
		"InvoiceId":   fmt.Sprintf("%d", req.PaymentID),
		"AccountId":   fmt.Sprintf("%d", req.UserID),
		"Description": "Оплата VPN подписки",
		"RequireConfirmation": false,
	}

	if req.Email != "" {
		requestData["Email"] = req.Email
	}
	if req.Phone != "" {
		requestData["Phone"] = req.Phone
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем HTTP запрос
	httpReq, err := http.NewRequestWithContext(ctx, "POST", 
		"https://api.cloudpayments.ru/payments/cards/charge", 
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(p.publicID, p.apiSecret)

	// Отправляем запрос
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CloudPaymentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("cloudpayments error: %s", result.Message)
	}

	return &result, nil
}

func (p *CloudPaymentsProvider) ParseWebhook(payload []byte) (*CloudPaymentsWebhook, error) {
	var webhook CloudPaymentsWebhook
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Преобразуем статус CloudPayments в наш формат
	status := StatusFailed
	switch webhook.Status {
	case "Completed":
		status = StatusCompleted
	case "Cancelled":
		status = StatusCancelled
	case "Declined":
		status = StatusFailed
	}

	webhook.Status = status
	return &webhook, nil
}

// ValidateWebhook проверяет подпись webhook от CloudPayments
func (p *CloudPaymentsProvider) ValidateWebhook(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(p.apiSecret))
	mac.Write(payload)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}