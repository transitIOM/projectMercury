package filesystem

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

type GTFSAdapter struct {
	gtfsPath     string
	messagesPath string
	checksumPath string
}

func NewGTFSAdapter(gtfsPath, messagesPath string) *GTFSAdapter {
	return &GTFSAdapter{
		gtfsPath:     gtfsPath,
		messagesPath: messagesPath,
		checksumPath: gtfsPath + ".sha256",
	}
}

func (a *GTFSAdapter) GetGTFSReader(ctx context.Context) (io.ReadCloser, error) {
	return os.Open(a.gtfsPath)
}

func (a *GTFSAdapter) GetChecksum(ctx context.Context) (string, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Try reading stored checksum
	if data, err := os.ReadFile(a.checksumPath); err == nil {
		return string(data), nil
	}

	// Calculate if not stored
	file, err := os.Open(a.gtfsPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	// Use a small buffer to allow for context cancellation during copy
	buf := make([]byte, 32*1024)
	for range 1000000 { // Large limit but better than infinite loop for security
		if err := ctx.Err(); err != nil {
			return "", err
		}
		n, err := file.Read(buf)
		if n > 0 {
			if _, hashErr := hash.Write(buf[:n]); hashErr != nil {
				return "", hashErr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	// Store it
	_ = os.WriteFile(a.checksumPath, []byte(checksum), 0644)

	return checksum, nil
}
