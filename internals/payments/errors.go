package payments

import "errors"

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidUserChatID    = errors.New("invalid user chat id")
	ErrWebhookValidationFailed = errors.New("webhook validation failed")
)
