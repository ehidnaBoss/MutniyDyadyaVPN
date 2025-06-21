package payments

import (
	"context"
	"database/sql"
	"mutinydydayvpn/internals/domain"
)


type Repository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	Update(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id int) (*domain.Payment, error)
	GetByProviderTxID(ctx context.Context, txID string) (*domain.Payment, error)
	GetByUserChatID(ctx context.Context, userChatID int64) ([]*domain.Payment, error)
	GetByStatus(ctx context.Context, status string) ([]*domain.Payment, error)
}

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
	INSERT INTO payments (user_chat_id, subscription_id, amount, currency, status, payment_method, created_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		payment.UserChatID,
		payment.SubscriptionID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.CreatedAt,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	payment.ID = int(id)
	return nil
}

func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	query := `
	UPDATE payments 
	SET status = ?, provider_transaction_id = ?, confirmed_at = ?
	WHERE id = ?
	`
	
	_, err := r.db.ExecContext(ctx, query,
		payment.Status,
		payment.ProviderTransactionID,
		payment.ConfirmedAt,
		payment.ID,
	)
	
	return err
}

func (r *PaymentRepository) GetByID(ctx context.Context, id int) (*domain.Payment, error) {
	query := `
	SELECT id, user_chat_id, subscription_id, amount, currency, status, 
	       payment_method, provider_transaction_id, created_at, confirmed_at
	FROM payments 
	WHERE id = ?
	`
	
	payment := &domain.Payment{}
	var providerTxID, confirmedAt sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.UserChatID,
		&payment.SubscriptionID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&providerTxID,
		&payment.CreatedAt,
		&confirmedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	payment.ProviderTransactionID = providerTxID.String
	payment.ConfirmedAt = confirmedAt.String
	
	return payment, nil
}

func (r *PaymentRepository) GetByProviderTransactionID(ctx context.Context, txID string) (*domain.Payment, error) {
	query := `
	SELECT id, user_chat_id, subscription_id, amount, currency, status, 
	       payment_method, provider_transaction_id, created_at, confirmed_at
	FROM payments 
	WHERE provider_transaction_id = ?
	`
	
	payment := &domain.Payment{}
	var providerTxID, confirmedAt sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, txID).Scan(
		&payment.ID,
		&payment.UserChatID,
		&payment.SubscriptionID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.PaymentMethod,
		&providerTxID,
		&payment.CreatedAt,
		&confirmedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	payment.ProviderTransactionID = providerTxID.String
	payment.ConfirmedAt = confirmedAt.String
	
	return payment, nil
}

func (r *PaymentRepository) GetByUserChatID(ctx context.Context, userChatID int64) ([]*domain.Payment, error) {
	query := `
	SELECT id, user_chat_id, subscription_id, amount, currency, status, 
	       payment_method, provider_transaction_id, created_at, confirmed_at
	FROM payments 
	WHERE user_chat_id = ?
	ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userChatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var providerTxID, confirmedAt sql.NullString
		
		err := rows.Scan(
			&payment.ID,
			&payment.UserChatID,
			&payment.SubscriptionID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentMethod,
			&providerTxID,
			&payment.CreatedAt,
			&confirmedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		payment.ProviderTransactionID = providerTxID.String
		payment.ConfirmedAt = confirmedAt.String
		
		payments = append(payments, payment)
	}
	
	return payments, nil
}

func (r *PaymentRepository) GetByStatus(ctx context.Context, status string) ([]*domain.Payment, error) {
	query := `
	SELECT id, user_chat_id, subscription_id, amount, currency, status, 
	       payment_method, provider_transaction_id, created_at, confirmed_at
	FROM payments 
	WHERE status = ?
	ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		payment := &domain.Payment{}
		var providerTxID, confirmedAt sql.NullString
		
		err := rows.Scan(
			&payment.ID,
			&payment.UserChatID,
			&payment.SubscriptionID,
			&payment.Amount,
			&payment.Currency,
			&payment.Status,
			&payment.PaymentMethod,
			&providerTxID,
			&payment.CreatedAt,
			&confirmedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		payment.ProviderTransactionID = providerTxID.String
		payment.ConfirmedAt = confirmedAt.String
		
		payments = append(payments, payment)
	}
	
	return payments, nil
}