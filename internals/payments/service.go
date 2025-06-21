package payments

import (
	"context"
	"fmt"
	"time"

	"mutinydydayvpn/internals/domain")

type Service struct {
	repo     Repository
	provider *CloudPaymentsProvider
}

func NewService(repo Repository, cpProvider *CloudPaymentsProvider) *Service {
	return &Service{
		repo:     repo,
		provider: cpProvider,
	}
}

// CreatePayment создает новый платеж
func (s *Service) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*domain.Payment, error) {
	// Валидация
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid payment request: %w", err)
	}

	// Создаем платеж
	payment := &domain.Payment{
		UserChatID:     req.UserChatID,
		SubscriptionID: req.SubscriptionID,
		Amount:         req.Amount,
		Currency:       "RUB",
		Status:         StatusPending,
		PaymentMethod:  "cloudpayments",
		CreatedAt:      time.Now().Format(time.RFC3339),
	}

	// Сохраняем в БД
	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Создаем платеж в CloudPayments
	providerReq := CloudPaymentsRequest{
		Amount:    req.Amount,
		Currency:  "RUB",
		PaymentID: payment.ID,
		UserID:    req.UserChatID,
		Email:     req.Email,
		Phone:     req.Phone,
	}

	result, err := s.provider.CreatePayment(ctx, providerReq)
	if err != nil {
		// Обновляем статус на failed
		payment.Status = StatusFailed
		s.repo.Update(ctx, payment)
		return nil, fmt.Errorf("cloudpayments failed: %w", err)
	}

	// Обновляем платеж данными от CloudPayments
	payment.ProviderTransactionID = result.TransactionID
	if err := s.repo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return payment, nil
}

// ProcessWebhook обрабатывает webhook от CloudPayments
func (s *Service) ProcessWebhook(ctx context.Context, payload []byte) error {
	webhook, err := s.provider.ParseWebhook(payload)
	if err != nil {
		return fmt.Errorf("failed to parse webhook: %w", err)
	}

	// Находим платеж по ID от CloudPayments
	payment, err := s.repo.GetByProviderTransactionID(ctx, webhook.TransactionID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	// Обновляем статус
	oldStatus := payment.Status
	payment.Status = webhook.Status
	
	if webhook.Status == StatusCompleted {
		payment.ConfirmedAt = time.Now().Format(time.RFC3339)
	}

	if err := s.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Если платеж успешно завершен, активируем подписку
	if oldStatus != StatusCompleted && payment.Status == StatusCompleted {
		if err := s.activateSubscription(ctx, payment); err != nil {
			return fmt.Errorf("failed to activate subscription: %w", err)
		}
	}

	return nil
}

// GetPayment получает платеж по ID
func (s *Service) GetPayment(ctx context.Context, id int) (*domain.Payment, error) {
	return s.repo.GetByID(ctx, id)
}

// GetUserPayments получает все платежи пользователя
func (s *Service) GetUserPayments(ctx context.Context, userChatID int64) ([]*domain.Payment, error) {
	return s.repo.GetByUserChatID(ctx, userChatID)
}

// activateSubscription активирует подписку после успешного платежа
func (s *Service) activateSubscription(ctx context.Context, payment *domain.Payment) error {
	// Здесь должна быть интеграция с сервисом подписок
	// Пока заглушка
	return nil
}