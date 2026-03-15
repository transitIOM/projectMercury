package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/transitIOM/projectMercury/internal/domain/models"
)

func TestMessageAdapter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mercury-msg-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	gtfsFile := filepath.Join(tempDir, "GTFS.zip")
	messagesFile := filepath.Join(tempDir, "messages.json")

	adapter := NewGTFSAdapter(gtfsFile, messagesFile)

	t.Run("GetMessages Empty", func(t *testing.T) {
		messages, version, err := adapter.GetMessages(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, messages)
		assert.Empty(t, version)
	})

	t.Run("Save and Get Messages", func(t *testing.T) {
		msg := models.Message{
			Content: "Test message",
		}
		version, err := adapter.SaveMessage(context.Background(), msg)
		assert.NoError(t, err)
		assert.NotEmpty(t, version)

		messages, v2, err := adapter.GetMessages(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(messages))
		assert.Equal(t, "Test message", messages[0].Content)
		assert.NotEmpty(t, messages[0].ID)
		assert.False(t, messages[0].Timestamp.IsZero())
		assert.Equal(t, version, v2)

		// Save another
		msg2 := models.Message{
			ID:        "manual-id",
			Content:   "Second message",
			Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		}
		_, err = adapter.SaveMessage(context.Background(), msg2)
		assert.NoError(t, err)

		messages, _, err = adapter.GetMessages(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 2, len(messages))
		assert.Equal(t, "manual-id", messages[1].ID)
		assert.True(t, messages[1].Timestamp.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	})
}
