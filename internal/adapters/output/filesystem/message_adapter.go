package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/transitIOM/projectMercury/internal/domain/models"
)

func (a *GTFSAdapter) GetMessages(ctx context.Context) ([]models.Message, string, error) {
	if _, err := os.Stat(a.messagesPath); os.IsNotExist(err) {
		return []models.Message{}, "", nil
	}

	data, err := os.ReadFile(a.messagesPath)
	if err != nil {
		return nil, "", err
	}

	var messages []models.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, "", err
	}

	info, _ := os.Stat(a.messagesPath)
	version := info.ModTime().Format(time.RFC3339)

	return messages, version, nil
}

func (a *GTFSAdapter) SaveMessage(ctx context.Context, message models.Message) (string, error) {
	messages, _, err := a.GetMessages(ctx)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if message.ID == "" {
		message.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	messages = append(messages, message)

	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(a.messagesPath), 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(a.messagesPath, data, 0644); err != nil {
		return "", err
	}

	return message.Timestamp.Format(time.RFC3339), nil
}
