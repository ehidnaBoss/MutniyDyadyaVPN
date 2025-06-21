package domain

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

type Payment struct {
	ID                      int     `db:"id" json:"id"`
	UserChatID              int64   `db:"user_chat_id" json:"user_chat_id"`
	SubscriptionID          int     `db:"subscription_id" json:"subscription_id"`
	Amount                  float64 `db:"amount" json:"amount"`
	Currency                string  `db:"currency" json:"currency"`
	Status                  string  `db:"status" json:"status"`
	PaymentMethod           string  `db:"payment_method" json:"payment_method"`
	ProviderTransactionID string  `json:"provider_transaction_id" db:"provider_transaction_id"`
	CreatedAt               string  `db:"created_at" json:"created_at"`
	ConfirmedAt             string  `db:"confirmed_at" json:"confirmed_at"`
}