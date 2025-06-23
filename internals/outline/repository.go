package outline

import (
	"context"
	"database/sql"
	"mutinydydayvpn/internals/domain"
)

type VPNKeyRepository struct {
	db *sql.DB
}

func NewVPNKeyRepository(db *sql.DB) *VPNKeyRepository {
	return &VPNKeyRepository{db: db}
}

func (r *VPNKeyRepository) Create(ctx context.Context, key *domain.VPNKey) error {
	query := `
	INSERT INTO vpn_keys (outline_key_id, user_chat_id, key_name, access_url, plan_months, is_active, created_at, expires_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		key.OutlineKeyID,
		key.UserChatID,
		key.KeyName,
		key.AccessURL,
		key.PlanMonths,
		key.IsActive,
		key.CreatedAt,
		key.ExpiresAt,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	key.ID = int(id)
	return nil
}

func (r *VPNKeyRepository) Update(ctx context.Context, key *domain.VPNKey) error {
	query := `
	UPDATE vpn_keys 
	SET key_name = ?, access_url = ?, is_active = ?, expires_at = ?
	WHERE id = ?
	`
	
	_, err := r.db.ExecContext(ctx, query,
		key.KeyName,
		key.AccessURL,
		key.IsActive,
		key.ExpiresAt,
		key.ID,
	)
	
	return err
}

func (r *VPNKeyRepository) GetByID(ctx context.Context, id int) (*domain.VPNKey, error) {
	query := `
	SELECT id, outline_key_id, user_chat_id, key_name, access_url, plan_months, is_active, created_at, expires_at
	FROM vpn_keys 
	WHERE id = ?
	`
	
	key := &domain.VPNKey{}
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&key.ID,
		&key.OutlineKeyID,
		&key.UserChatID,
		&key.KeyName,
		&key.AccessURL,
		&key.PlanMonths,
		&key.IsActive,
		&key.CreatedAt,
		&key.ExpiresAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return key, nil
}

func (r *VPNKeyRepository) GetByUserChatID(ctx context.Context, userChatID int64) ([]*domain.VPNKey, error) {
	query := `
	SELECT id, outline_key_id, user_chat_id, key_name, access_url, plan_months, is_active, created_at, expires_at
	FROM vpn_keys 
	WHERE user_chat_id = ?
	ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userChatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.VPNKey
	for rows.Next() {
		key := &domain.VPNKey{}
		
		err := rows.Scan(
			&key.ID,
			&key.OutlineKeyID,
			&key.UserChatID,
			&key.KeyName,
			&key.AccessURL,
			&key.PlanMonths,
			&key.IsActive,
			&key.CreatedAt,
			&key.ExpiresAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		keys = append(keys, key)
	}
	
	return keys, nil
}

func (r *VPNKeyRepository) GetByOutlineKeyID(ctx context.Context, outlineKeyID string) (*domain.VPNKey, error) {
	query := `
	SELECT id, outline_key_id, user_chat_id, key_name, access_url, plan_months, is_active, created_at, expires_at
	FROM vpn_keys 
	WHERE outline_key_id = ?
	`
	
	key := &domain.VPNKey{}
	
	err := r.db.QueryRowContext(ctx, query, outlineKeyID).Scan(
		&key.ID,
		&key.OutlineKeyID,
		&key.UserChatID,
		&key.KeyName,
		&key.AccessURL,
		&key.PlanMonths,
		&key.IsActive,
		&key.CreatedAt,
		&key.ExpiresAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return key, nil
}

func (r *VPNKeyRepository) GetActiveKeys(ctx context.Context) ([]*domain.VPNKey, error) {
	query := `
	SELECT id, outline_key_id, user_chat_id, key_name, access_url, plan_months, is_active, created_at, expires_at
	FROM vpn_keys 
	WHERE is_active = true
	ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.VPNKey
	for rows.Next() {
		key := &domain.VPNKey{}
		
		err := rows.Scan(
			&key.ID,
			&key.OutlineKeyID,
			&key.UserChatID,
			&key.KeyName,
			&key.AccessURL,
			&key.PlanMonths,
			&key.IsActive,
			&key.CreatedAt,
			&key.ExpiresAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		keys = append(keys, key)
	}
	
	return keys, nil
}

func (r *VPNKeyRepository) GetExpiredKeys(ctx context.Context) ([]*domain.VPNKey, error) {
	query := `
	SELECT id, outline_key_id, user_chat_id, key_name, access_url, plan_months, is_active, created_at, expires_at
	FROM vpn_keys 
	WHERE expires_at < datetime('now') AND is_active = true
	ORDER BY expires_at ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.VPNKey
	for rows.Next() {
		key := &domain.VPNKey{}
		
		err := rows.Scan(
			&key.ID,
			&key.OutlineKeyID,
			&key.UserChatID,
			&key.KeyName,
			&key.AccessURL,
			&key.PlanMonths,
			&key.IsActive,
			&key.CreatedAt,
			&key.ExpiresAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		keys = append(keys, key)
	}
	
	return keys, nil
}

func (r *VPNKeyRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM vpn_keys WHERE id = ?`
	
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}