package payments



type CreatePaymentRequest struct {
	UserChatID     int64   `json:"user_chat_id" validate:"required"`
	SubscriptionID int     `json:"subscription_id" validate:"required"`
	Amount         float64 `json:"amount" validate:"required,gt=0"`
	Email          string  `json:"email,omitempty"`
	Phone          string  `json:"phone,omitempty"`
}

func (r *CreatePaymentRequest) Validate() error {
	if r.UserChatID == 0 {
		return ErrInvalidUserChatID
	}
	if r.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

type PaymentResponse struct {
	ID             int     `json:"id"`
	Status         string  `json:"status"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentURL     string  `json:"payment_url,omitempty"`
	CreatedAt      string  `json:"created_at"`
}