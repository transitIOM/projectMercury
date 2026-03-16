package cura

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Adapter struct {
	owner      string
	repo       string
	targetPath string
	client     *http.Client
}

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func NewAdapter(owner, repo, targetPath string) *Adapter {
	return &Adapter{
		owner:      owner,
		repo:       repo,
		targetPath: targetPath,
		client:     &http.Client{Timeout: 5 * time.Minute},
	}
}

func (a *Adapter) FetchLatestSchedule(ctx context.Context) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", a.owner, a.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cura api returned status: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		slog.ErrorContext(ctx, "failed to decode cura release response", "error", err)
		return err
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "cura-gtfs-v1.0.1.zip" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		slog.WarnContext(ctx, "GTFS.zip not found in latest Cura release assets", "tag", release.TagName)
		return fmt.Errorf("GTFS.zip not found in latest release assets")
	}

	slog.InfoContext(ctx, "Downloading latest GTFS release", "tag", release.TagName)
	return a.downloadFile(ctx, downloadURL)
}

func (a *Adapter) downloadFile(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, status: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(a.targetPath), 0755); err != nil {
		return err
	}

	// Use a temporary file to ensure atomic update
	tmpFile := a.targetPath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.Remove(tmpFile)

	if _, err = io.Copy(out, resp.Body); err != nil {
		slog.ErrorContext(ctx, "failed to copy GTFS to temporary file", "error", err)
		return err
	}
	out.Close()

	// Atomic rename
	if err := os.Rename(tmpFile, a.targetPath); err != nil {
		slog.ErrorContext(ctx, "failed to perform atomic swap of GTFS file", "error", err)
		return err
	}
	slog.InfoContext(ctx, "Successfully updated local GTFS schedule", "path", a.targetPath)
	return nil
}
