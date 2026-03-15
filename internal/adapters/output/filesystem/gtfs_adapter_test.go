package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGTFSAdapter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mercury-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	gtfsFile := filepath.Join(tempDir, "GTFS.zip")
	err = os.WriteFile(gtfsFile, []byte("fake gtfs content"), 0644)
	assert.NoError(t, err)

	messagesFile := filepath.Join(tempDir, "messages.json")

	adapter := NewGTFSAdapter(gtfsFile, messagesFile)

	t.Run("GetGTFSReader", func(t *testing.T) {
		reader, err := adapter.GetGTFSReader(context.Background())
		assert.NoError(t, err)
		defer reader.Close()
		content, err := os.ReadFile(gtfsFile)
		assert.NoError(t, err)
		assert.Equal(t, "fake gtfs content", string(content))
	})

	t.Run("GetChecksum", func(t *testing.T) {
		checksum, err := adapter.GetChecksum(context.Background())
		assert.NoError(t, err)
		assert.NotEmpty(t, checksum)

		// Verify stored checksum
		storedChecksum, err := os.ReadFile(gtfsFile + ".sha256")
		assert.NoError(t, err)
		assert.Equal(t, checksum, string(storedChecksum))

		// Check context cancellation
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = adapter.GetChecksum(ctx)
		assert.Error(t, err)
	})
}
