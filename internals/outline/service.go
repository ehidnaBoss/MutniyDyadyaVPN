package outline

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"mutinydydayvpn/internals/domain"
)

type Service struct {
	client *Client
	config *Config
}

type Config struct {
	ServerURL     string
	ApiKey        string
	DefaultLimitMB int64  // Лимит по умолчанию в MB
	KeyPrefix     string  // Префикс для имен ключей
}

func NewService(config *Config) *Service {
	client := NewClient(config.ServerURL, config.ApiKey)
	
	return &Service{
		client: client,
		config: config,
	}
}

// CreateUserKey создает ключ для пользователя
func (s *Service) CreateUserKey(ctx context.Context, userChatID int64, planMonths int) (*domain.VPNKey, error) {
	// Генерируем имя ключа
	keyName := fmt.Sprintf("%suser_%d_%dmonths", s.config.KeyPrefix, userChatID, planMonths)
	
	// Создаем ключ в Outline
	outlineKey, err := s.client.CreateKey(ctx, keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to create outline key: %w", err)
	}
	
	// Устанавливаем лимит трафика если настроен
	if s.config.DefaultLimitMB > 0 {
		limitBytes := s.config.DefaultLimitMB * 1024 * 1024 // Переводим MB в байты
		if err := s.client.SetDataLimit(ctx, outlineKey.ID, limitBytes); err != nil {
			log.Printf("Warning: failed to set data limit for key %s: %v", outlineKey.ID, err)
		}
	}
	
	// Создаем объект VPN ключа
	vpnKey := &domain.VPNKey{
		OutlineKeyID: outlineKey.ID,
		UserChatID:   userChatID,
		KeyName:      keyName,
		AccessURL:    outlineKey.AccessURL,
		PlanMonths:   planMonths,
		IsActive:     true,
		CreatedAt:    time.Now().Format(time.RFC3339),
		ExpiresAt:    time.Now().AddDate(0, planMonths, 0).Format(time.RFC3339),
	}
	
	return vpnKey, nil
}

// DeactivateKey деактивирует ключ (удаляет из Outline)
func (s *Service) DeactivateKey(ctx context.Context, outlineKeyID string) error {
	if err := s.client.DeleteKey(ctx, outlineKeyID); err != nil {
		return fmt.Errorf("failed to delete outline key: %w", err)
	}
	
	return nil
}

// GetKeyUsage получает статистику использования ключа
func (s *Service) GetKeyUsage(ctx context.Context, outlineKeyID string) (int64, error) {
	usage, err := s.client.GetKeyUsage(ctx, outlineKeyID)
	if err != nil {
		return 0, fmt.Errorf("failed to get key usage: %w", err)
	}
	
	return usage, nil
}

// UpdateKeyLimit обновляет лимит трафика для ключа
func (s *Service) UpdateKeyLimit(ctx context.Context, outlineKeyID string, limitMB int64) error {
	limitBytes := limitMB * 1024 * 1024
	
	if err := s.client.SetDataLimit(ctx, outlineKeyID, limitBytes); err != nil {
		return fmt.Errorf("failed to update key limit: %w", err)
	}
	
	return nil
}

// RemoveKeyLimit удаляет лимит трафика для ключа
func (s *Service) RemoveKeyLimit(ctx context.Context, outlineKeyID string) error {
	if err := s.client.RemoveDataLimit(ctx, outlineKeyID); err != nil {
		return fmt.Errorf("failed to remove key limit: %w", err)
	}
	
	return nil
}

// GetAllKeys получает все ключи с сервера
func (s *Service) GetAllKeys(ctx context.Context) ([]OutlineKey, error) {
	keys, err := s.client.GetKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}
	
	return keys, nil
}

// GetServerInfo получает информацию о сервере
func (s *Service) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	info, err := s.client.GetServerInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	
	return info, nil
}

// GetServerStats получает статистику сервера
func (s *Service) GetServerStats(ctx context.Context) (*domain.ServerStats, error) {
	// Получаем общую информацию
	info, err := s.client.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}
	
	// Получаем все ключи
	keys, err := s.client.GetKeys(ctx)
	if err != nil {
		return nil, err
	}
	
	// Получаем статистику использования
	usage, err := s.client.GetDataUsage(ctx)
	if err != nil {
		return nil, err
	}
	
	// Подсчитываем общую статистику
	var totalUsage int64
	activeKeys := 0
	
	for _, key := range keys {
		if keyUsage, exists := usage.BytesTransferredByUserId[key.ID]; exists {
			totalUsage += keyUsage
		}
		// Считаем ключ активным, если он существует (в Outline нет явного статуса)
		activeKeys++
	}
	
	stats := &domain.ServerStats{
		ServerName:    info.Name,
		ServerID:      info.ServerID,
		TotalKeys:     len(keys),
		ActiveKeys:    activeKeys,
		TotalUsageGB:  float64(totalUsage) / (1024 * 1024 * 1024),
		CreatedAt:     time.Unix(info.CreatedTimestampMs/1000, 0).Format(time.RFC3339),
	}
	
	return stats, nil
}

// TestConnection проверяет подключение к серверу
func (s *Service) TestConnection(ctx context.Context) error {
	return s.client.TestConnection(ctx)
}

// ExtendKey продлевает срок действия ключа (логически, не на уровне Outline)
func (s *Service) ExtendKey(ctx context.Context, vpnKey *domain.VPNKey, additionalMonths int) error {
	// Парсим текущую дату истечения
	expiresAt, err := time.Parse(time.RFC3339, vpnKey.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to parse expires_at: %w", err)
	}
	
	// Продлеваем на указанное количество месяцев
	newExpiresAt := expiresAt.AddDate(0, additionalMonths, 0)
	vpnKey.ExpiresAt = newExpiresAt.Format(time.RFC3339)
	
	return nil
}

// IsKeyExpired проверяет, истек ли ключ
func (s *Service) IsKeyExpired(vpnKey *domain.VPNKey) bool {
	expiresAt, err := time.Parse(time.RFC3339, vpnKey.ExpiresAt)
	if err != nil {
		return true // Считаем истекшим если не можем распарсить дату
	}
	
	return time.Now().After(expiresAt)
}

// GetExpiredKeys получает список истекших ключей
func (s *Service) GetExpiredKeys(ctx context.Context, vpnKeys []*domain.VPNKey) []*domain.VPNKey {
	var expiredKeys []*domain.VPNKey
	
	for _, key := range vpnKeys {
		if s.IsKeyExpired(key) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	return expiredKeys
}

// CleanupExpiredKeys удаляет истекшие ключи из Outline
func (s *Service) CleanupExpiredKeys(ctx context.Context, expiredKeys []*domain.VPNKey) error {
	for _, key := range expiredKeys {
		if err := s.DeactivateKey(ctx, key.OutlineKeyID); err != nil {
			log.Printf("Failed to cleanup expired key %s: %v", key.OutlineKeyID, err)
			continue
		}
		log.Printf("Cleaned up expired key: %s", key.OutlineKeyID)
	}
	
	return nil
}