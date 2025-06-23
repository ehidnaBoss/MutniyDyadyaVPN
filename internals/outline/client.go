// internals/outline/client.go
package outline

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(serverURL, apiKey string) *Client {
	// Создаем HTTP клиент с отключенной проверкой SSL (для самоподписанных сертификатов)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	
	return &Client{
		baseURL: serverURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// Структуры для работы с Outline API

type OutlineKey struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Password   string `json:"password"`
	Port       int    `json:"port"`
	Method     string `json:"method"`
	AccessURL  string `json:"accessUrl"`
	UsedBytes  int64  `json:"usedBytes,omitempty"`
}

type CreateKeyRequest struct {
	Name   string `json:"name,omitempty"`
	Method string `json:"method,omitempty"`
	Port   int    `json:"port,omitempty"`
}

type RenameKeyRequest struct {
	Name string `json:"name"`
}

type SetDataLimitRequest struct {
	Limit struct {
		Bytes int64 `json:"bytes"`
	} `json:"limit"`
}

type ServerInfo struct {
	Name                   string `json:"name"`
	ServerID               string `json:"serverId"`
	MetricsEnabled         bool   `json:"metricsEnabled"`
	CreatedTimestampMs     int64  `json:"createdTimestampMs"`
	Version                string `json:"version"`
	AccessKeyDataLimit     int64  `json:"accessKeyDataLimit,omitempty"`
	PortForNewAccessKeys   int    `json:"portForNewAccessKeys"`
	HostnameForAccessKeys  string `json:"hostnameForAccessKeys"`
}

type DataUsage struct {
	BytesTransferredByUserId map[string]int64 `json:"bytesTransferredByUserId"`
}

// API методы

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}
	
	return respBody, nil
}

// CreateKey создает новый ключ доступа
func (c *Client) CreateKey(ctx context.Context, name string) (*OutlineKey, error) {
	req := CreateKeyRequest{
		Name: name,
	}
	
	respBody, err := c.makeRequest(ctx, "POST", "/access-keys", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}
	
	var key OutlineKey
	if err := json.Unmarshal(respBody, &key); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &key, nil
}

// GetKeys получает список всех ключей
func (c *Client) GetKeys(ctx context.Context) ([]OutlineKey, error) {
	respBody, err := c.makeRequest(ctx, "GET", "/access-keys", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}
	
	var response struct {
		AccessKeys []OutlineKey `json:"accessKeys"`
	}
	
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return response.AccessKeys, nil
}

// GetKey получает информацию о конкретном ключе
func (c *Client) GetKey(ctx context.Context, keyID string) (*OutlineKey, error) {
	keys, err := c.GetKeys(ctx)
	if err != nil {
		return nil, err
	}
	
	for _, key := range keys {
		if key.ID == keyID {
			return &key, nil
		}
	}
	
	return nil, fmt.Errorf("key not found: %s", keyID)
}

// DeleteKey удаляет ключ доступа
func (c *Client) DeleteKey(ctx context.Context, keyID string) error {
	_, err := c.makeRequest(ctx, "DELETE", "/access-keys/"+keyID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	
	return nil
}

// RenameKey переименовывает ключ
func (c *Client) RenameKey(ctx context.Context, keyID, newName string) error {
	req := RenameKeyRequest{
		Name: newName,
	}
	
	_, err := c.makeRequest(ctx, "PUT", "/access-keys/"+keyID+"/name", req)
	if err != nil {
		return fmt.Errorf("failed to rename key: %w", err)
	}
	
	return nil
}

// SetDataLimit устанавливает лимит трафика для ключа
func (c *Client) SetDataLimit(ctx context.Context, keyID string, limitBytes int64) error {
	req := SetDataLimitRequest{}
	req.Limit.Bytes = limitBytes
	
	_, err := c.makeRequest(ctx, "PUT", "/access-keys/"+keyID+"/data-limit", req)
	if err != nil {
		return fmt.Errorf("failed to set data limit: %w", err)
	}
	
	return nil
}

// RemoveDataLimit удаляет лимит трафика для ключа
func (c *Client) RemoveDataLimit(ctx context.Context, keyID string) error {
	_, err := c.makeRequest(ctx, "DELETE", "/access-keys/"+keyID+"/data-limit", nil)
	if err != nil {
		return fmt.Errorf("failed to remove data limit: %w", err)
	}
	
	return nil
}

// GetServerInfo получает информацию о сервере
func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	respBody, err := c.makeRequest(ctx, "GET", "/server", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}
	
	var info ServerInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &info, nil
}

// GetDataUsage получает статистику использования трафика
func (c *Client) GetDataUsage(ctx context.Context) (*DataUsage, error) {
	respBody, err := c.makeRequest(ctx, "GET", "/metrics/transfer", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get data usage: %w", err)
	}
	
	var usage DataUsage
	if err := json.Unmarshal(respBody, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &usage, nil
}

// GetKeyUsage получает статистику использования конкретного ключа
func (c *Client) GetKeyUsage(ctx context.Context, keyID string) (int64, error) {
	usage, err := c.GetDataUsage(ctx)
	if err != nil {
		return 0, err
	}
	
	if bytes, exists := usage.BytesTransferredByUserId[keyID]; exists {
		return bytes, nil
	}
	
	return 0, nil
}

// SetServerName устанавливает имя сервера
func (c *Client) SetServerName(ctx context.Context, name string) error {
	req := map[string]string{"name": name}
	
	_, err := c.makeRequest(ctx, "PUT", "/name", req)
	if err != nil {
		return fmt.Errorf("failed to set server name: %w", err)
	}
	
	return nil
}

// SetHostnameForAccessKeys устанавливает hostname для ключей доступа
func (c *Client) SetHostnameForAccessKeys(ctx context.Context, hostname string) error {
	req := map[string]string{"hostname": hostname}
	
	_, err := c.makeRequest(ctx, "PUT", "/server/hostname-for-access-keys", req)
	if err != nil {
		return fmt.Errorf("failed to set hostname: %w", err)
	}
	
	return nil
}

// SetPortForNewAccessKeys устанавливает порт по умолчанию для новых ключей
func (c *Client) SetPortForNewAccessKeys(ctx context.Context, port int) error {
	req := map[string]int{"port": port}
	
	_, err := c.makeRequest(ctx, "PUT", "/server/port-for-new-access-keys", req)
	if err != nil {
		return fmt.Errorf("failed to set port: %w", err)
	}
	
	return nil
}



