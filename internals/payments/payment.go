package payments

type Payment struct {
	ID                      int     `json:"id"`
	UserChatID              int64   `json:"user_chat_id"`
	SubscriptionID          int     `json:"subscription_id"`
	Amount                  float64 `json:"amount"`
	Currency                string  `json:"currency"`
	Status                  string  `json:"status"`
	PaymentMethod           string  `json:"payment_method"`
	ProviderTransactionID   string  `json:"provider_transaction_id"`
	CreatedAt               string  `json:"created_at"`
	ConfirmedAt             string  `json:"confirmed_at"`
}

type User struct {
	ID        int    `json:"id"`
	ChatID    int64  `json:"chat_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt string `json:"created_at"`
}

type Subscription struct {
	ID            int    `json:"id"`
	UserChatID    int64  `json:"user_chat_id"`
	PlanType      string `json:"plan_type"`
	ActivationKey string `json:"activation_key"`
	IsActive      bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
	ExpiresAt     string `json:"expires_at"`
}