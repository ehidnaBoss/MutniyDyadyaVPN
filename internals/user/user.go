package user

import (
	"context"
	"database/sql"
	"mutinydydayvpn/internals/domain"
)

type Service struct {
	repo Repository
}

type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByChatID(ctx context.Context, chatID int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

type UserRepository struct {
	db *sql.DB
}

type CreateUserRequest struct {
	ChatID    int64  `json:"chat_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (s *Service) CreateOrUpdateUser(ctx context.Context, req *CreateUserRequest) (*domain.User, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.repo.GetByChatID(ctx, req.ChatID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if existingUser != nil {
		// Обновляем существующего пользователя
		existingUser.Username = req.Username
		existingUser.FirstName = req.FirstName
		existingUser.LastName = req.LastName
		
		if err := s.repo.Update(ctx, existingUser); err != nil {
			return nil, err
		}
		return existingUser, nil
	}

	// Создаем нового пользователя
	user := &domain.User{
		ChatID:    req.ChatID,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Repository implementation
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (chat_id, username, first_name, last_name) 
	VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.ExecContext(ctx, query,
		user.ChatID,
		user.Username,
		user.FirstName,
		user.LastName,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	user.ID = int(id)
	return nil
}

func (r *UserRepository) GetByChatID(ctx context.Context, chatID int64) (*domain.User, error) {
	query := `
	SELECT id, chat_id, username, first_name, last_name, created_at
	FROM users 
	WHERE chat_id = ?
	`
	
	user := &domain.User{}
	var username, firstName, lastName sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&user.ID,
		&user.ChatID,
		&username,
		&firstName,
		&lastName,
		&user.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	user.Username = username.String
	user.FirstName = firstName.String
	user.LastName = lastName.String
	
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
	UPDATE users 
	SET username = ?, first_name = ?, last_name = ?
	WHERE id = ?
	`
	
	_, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.ID,
	)
	
	return err
}