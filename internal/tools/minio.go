package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"sync"
	"time"

	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
)

// MinIOClient wraps the minio.Client to implement the ObjectStorageClient interface.
type MinIOClient struct {
	client *minio.Client
}

// NewMinIOClient creates a new MinIOClient wrapper around an existing minio.Client.
func NewMinIOClient(client *minio.Client) *MinIOClient {
	return &MinIOClient{client: client}
}

// BucketExists checks if a bucket exists in the object storage.
func (m *MinIOClient) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.client.BucketExists(ctx, bucketName)
}

// MakeBucket creates a new bucket with the given region.
func (m *MinIOClient) MakeBucket(ctx context.Context, bucketName string, region string) error {
	return m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region})
}

// SetBucketVersioning enables or disables versioning for a bucket.
// Versioning allows you to preserve, retrieve, and restore every version of every object.
func (m *MinIOClient) SetBucketVersioning(ctx context.Context, bucketName string, enabled bool) error {
	status := "Suspended"
	if enabled {
		status = "Enabled"
	}
	config := minio.BucketVersioningConfiguration{Status: status}
	return m.client.SetBucketVersioning(ctx, bucketName, config)
}

// ListBuckets returns a list of all buckets owned by the authenticated user.
func (m *MinIOClient) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	buckets, err := m.client.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	// Convert MinIO bucket info to our generic type
	result := make([]BucketInfo, len(buckets))
	for i, b := range buckets {
		result[i] = BucketInfo{
			Name:         b.Name,
			CreationDate: b.CreationDate,
		}
	}
	return result, nil
}

// GetObject retrieves an object from storage.
// Returned io.ReadCloser must be closed after use to release resources.
func (m *MinIOClient) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}

// PutObject uploads an object to storage.
// Returns upload information including the version ID if versioning is enabled.
func (m *MinIOClient) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (UploadInfo, error) {
	opts := minio.PutObjectOptions{ContentType: contentType}
	uploadInfo, err := m.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
	if err != nil {
		return UploadInfo{}, err
	}

	// Convert MinIO upload info to generic type
	return UploadInfo{
		Bucket:    uploadInfo.Bucket,
		Key:       uploadInfo.Key,
		VersionID: uploadInfo.VersionID,
		ETag:      uploadInfo.ETag,
	}, nil
}

// StatObject retrieves metadata about an object without downloading it.
func (m *MinIOClient) StatObject(ctx context.Context, bucketName, objectName string) (ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return ObjectInfo{}, err
	}

	// Convert MinIO object info to generic type
	return ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		ETag:         info.ETag,
		LastModified: info.LastModified,
		ContentType:  info.ContentType,
		VersionID:    info.VersionID,
	}, nil
}

// GetObjectAttributes retrieves object attributes including version ID.
// This is particularly useful when versioning is enabled.
func (m *MinIOClient) GetObjectAttributes(ctx context.Context, bucketName, objectName string) (ObjectAttributes, error) {
	attrs, err := m.client.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return ObjectAttributes{}, err
	}

	// Convert MinIO attributes to generic type
	return ObjectAttributes{
		VersionID: attrs.VersionID,
		ETag:      attrs.ETag,
	}, nil
}

// PresignedGetObject generates a presigned URL for downloading an object.
// The URL is valid for the specified expiry duration and can include custom request parameters.
func (m *MinIOClient) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error) {
	return m.client.PresignedGetObject(ctx, bucketName, objectName, expiry, reqParams)
}

// MinIOStorageManager implements the ObjectStorageManager interface.
// It provides a complete storage solution for GTFS schedules and message logs.
type MinIOStorageManager struct {
	client ObjectStorageClient
	ctx    context.Context

	// GTFS-specific fields
	gtfsBucketName string
	gtfsObjectName string
	gtfsMutex      sync.RWMutex

	// Messaging-specific fields
	messagingBucketName string
	messagingObjectName string
	messagingMutex      sync.RWMutex
}

// NewMinIOStorageManager creates a new storage manager with the given client.
// The manager uses default bucket and object names, which can be customized if needed.
func NewMinIOStorageManager(client ObjectStorageClient, ctx context.Context) *MinIOStorageManager {
	return &MinIOStorageManager{
		client:              client,
		ctx:                 ctx,
		gtfsBucketName:      "gtfs",
		gtfsObjectName:      "GTFSSchedule.zip",
		messagingBucketName: "messages",
		messagingObjectName: "messages.jsonl",
	}
}

// Initialize sets up the storage client and creates necessary buckets.
// It creates both the GTFS and messaging buckets with versioning enabled.
// It includes a retry mechanism for connection failures.
func (m *MinIOStorageManager) Initialize() error {
	maxAttempts := 10
	if attemptsStr := os.Getenv("MINIO_RETRY_ATTEMPTS"); attemptsStr != "" {
		if a, err := strconv.Atoi(attemptsStr); err == nil {
			maxAttempts = a
		}
	}

	retryInterval := 60 * time.Second
	if intervalStr := os.Getenv("MINIO_RETRY_INTERVAL"); intervalStr != "" {
		if i, err := strconv.Atoi(intervalStr); err == nil {
			retryInterval = time.Duration(i) * time.Second
		}
	}

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Infof("initializing storage (attempt %d/%d)...", attempt, maxAttempts)

		err = m.initBuckets()
		if err == nil {
			log.Info("storage initialized successfully")
			return nil
		}

		log.Warnf("failed to initialize storage on attempt %d: %v", attempt, err)

		if attempt < maxAttempts {
			log.Infof("retrying in %v", retryInterval)
			select {
			case <-m.ctx.Done():
				return m.ctx.Err()
			case <-time.After(retryInterval):
			}
		}
	}

	return fmt.Errorf("failed to initialize storage after %d attempts: %w", maxAttempts, err)
}

// initBuckets performs the actual bucket creation logic.
func (m *MinIOStorageManager) initBuckets() error {
	// Create GTFS bucket
	if err := m.createBucket(m.gtfsBucketName); err != nil {
		return fmt.Errorf("failed to create GTFS bucket: %w", err)
	}

	// Create messaging bucket
	if err := m.createBucket(m.messagingBucketName); err != nil {
		return fmt.Errorf("failed to create messaging bucket: %w", err)
	}

	return nil
}

// createBucket is a helper method that creates a bucket if it doesn't exist
// and enables versioning on it.
func (m *MinIOStorageManager) createBucket(bucketName string) error {
	// Check if bucket exists
	exists, err := m.client.BucketExists(m.ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// Create bucket if it doesn't exist
	if !exists {
		log.Debugf("Creating bucket: %s", bucketName)
		err = m.client.MakeBucket(m.ctx, bucketName, "")
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	} else {
		log.Debugf("Bucket already exists: %s", bucketName)
	}

	// Enable versioning
	log.Debugf("Setting versioning for bucket: %s", bucketName)
	err = m.client.SetBucketVersioning(m.ctx, bucketName, true)
	if err != nil {
		return fmt.Errorf("failed to set bucket versioning: %w", err)
	}

	return nil
}

// Close cleans up any resources.
// Currently, there are no resources to clean up, but this method is here
// for future extensibility and to satisfy the ObjectStorageManager interface.
func (m *MinIOStorageManager) Close() error {
	// No resources to clean up currently
	return nil
}

// ------------------------------------
// GTFSStorage Interface Implementation
// ------------------------------------

// GetLatestGTFSVersionID returns the version ID of the latest GTFS schedule.
// It uses GetObjectAttributes to retrieve the version ID without downloading the file.
func (m *MinIOStorageManager) GetLatestGTFSVersionID() (versionID string, err error) {
	m.gtfsMutex.RLock()
	defer m.gtfsMutex.RUnlock()

	log.Debugf("Getting attributes for %s/%s", m.gtfsBucketName, m.gtfsObjectName)
	attributes, err := m.client.GetObjectAttributes(m.ctx, m.gtfsBucketName, m.gtfsObjectName)
	if err != nil {
		// Check if the error is because the object doesn't exist
		if err.Error() == KeyNotFound {
			log.Debug("No GTFS schedule found on server")
			return "", NoGTFSScheduleFound
		}
		return "", err
	}

	log.Debugf("Latest GTFS version ID: %s", attributes.VersionID)
	return attributes.VersionID, nil
}

// GetLatestURL returns a presigned URL to download the latest GTFS schedule.
// The URL is valid for 5 minutes and includes headers to force download with the correct filename.
func (m *MinIOStorageManager) GetLatestURL() (downloadURL *url.URL, versionID string, err error) {
	// Set up presigned URL parameters
	expiryTime := 5 * time.Minute
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%s", m.gtfsObjectName))
	reqParams.Set("response-content-type", "application/zip")

	// Generate presigned URL
	m.gtfsMutex.RLock()
	downloadURL, err = m.client.PresignedGetObject(m.ctx, m.gtfsBucketName, m.gtfsObjectName, expiryTime, reqParams)
	m.gtfsMutex.RUnlock()

	if err != nil {
		if err.Error() == "The specified key does not exist." {
			log.Debug("No GTFS schedule found on server")
			return nil, "", NoGTFSScheduleFound
		}
		return nil, "", err
	}

	// Get the version ID
	versionID, err = m.GetLatestGTFSVersionID()
	if err != nil {
		return nil, "", err
	}

	return downloadURL, versionID, nil
}

// PutSchedule uploads a new GTFS schedule.
// It returns the version ID of the newly uploaded schedule.
func (m *MinIOStorageManager) PutSchedule(reader io.Reader, fileSize int64) (versionID string, err error) {
	m.gtfsMutex.Lock()
	defer m.gtfsMutex.Unlock()

	log.Debugf("Uploading %s to %s, size: %d", m.gtfsObjectName, m.gtfsBucketName, fileSize)
	uploadInfo, err := m.client.PutObject(
		m.ctx,
		m.gtfsBucketName,
		m.gtfsObjectName,
		reader,
		fileSize,
		"application/zip",
	)
	if err != nil {
		return "", err
	}

	log.Debugf("Successfully uploaded %s, version ID: %s", m.gtfsObjectName, uploadInfo.VersionID)
	return uploadInfo.VersionID, nil
}

// ---------------------------------------
// MessageStorage Interface Implementation
// ---------------------------------------

// AppendMessage appends a new message to the message log.
// It retrieves the existing log, appends the new message, and uploads the combined result.
func (m *MinIOStorageManager) AppendMessage(message *bytes.Buffer) (versionID string, err error) {
	m.messagingMutex.Lock()
	defer m.messagingMutex.Unlock()

	existingData := bytes.Buffer{}

	// Check if the message log already exists
	log.Debugf("Checking if %s exists in %s", m.messagingObjectName, m.messagingBucketName)
	_, err = m.client.StatObject(m.ctx, m.messagingBucketName, m.messagingObjectName)

	if err == nil {
		// File exists, retrieve it
		log.Debugf("Retrieving existing %s", m.messagingObjectName)
		r, err := m.client.GetObject(m.ctx, m.messagingBucketName, m.messagingObjectName)
		if err != nil {
			return "", fmt.Errorf("failed to get existing message log: %w", err)
		}
		defer func(r io.ReadCloser) {
			if closeErr := r.Close(); closeErr != nil {
				log.Error(closeErr)
			}
		}(r)

		_, err = existingData.ReadFrom(r)
		if err != nil {
			return "", fmt.Errorf("failed to read existing message log: %w", err)
		}
	} else {
		// error checking is storage-implementation specific
		if err.Error() != "The specified key does not exist." {
			return "", fmt.Errorf("failed to check message log existence: %w", err)
		}
		log.Debug("Message log does not exist, starting new one")
	}

	// Append the new message
	log.Debug("Appending new message to existing data")
	_, err = existingData.ReadFrom(message)
	if err != nil {
		return "", err
	}

	// Upload the combined data
	updatedReader := bytes.NewReader(existingData.Bytes())
	log.Debugf("Uploading updated %s to %s, total size: %d", m.messagingObjectName, m.messagingBucketName, existingData.Len())
	uploadInfo, err := m.client.PutObject(
		m.ctx,
		m.messagingBucketName,
		m.messagingObjectName,
		updatedReader,
		int64(existingData.Len()),
		"text/jsonl",
	)

	if err != nil {
		return "", err
	}

	log.Debugf("Successfully updated %s, version ID: %s", m.messagingObjectName, uploadInfo.VersionID)
	return uploadInfo.VersionID, nil
}

// GetLatestLog retrieves the latest message log.
// It downloads the entire log file and returns it as a buffer.
func (m *MinIOStorageManager) GetLatestLog() (messageLog *bytes.Buffer, err error) {
	m.messagingMutex.RLock()
	defer m.messagingMutex.RUnlock()

	log.Debugf("Retrieving %s from %s", m.messagingObjectName, m.messagingBucketName)
	r, err := m.client.GetObject(m.ctx, m.messagingBucketName, m.messagingObjectName)
	if err != nil {
		return nil, err
	}
	defer func(r io.ReadCloser) {
		if closeErr := r.Close(); closeErr != nil {
			log.Error(closeErr)
		}
	}(r)

	messageLog = &bytes.Buffer{}
	_, err = messageLog.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	log.Debugf("Successfully retrieved message log, size: %d bytes", messageLog.Len())
	return messageLog, nil
}

// GetLatestMessageVersionID returns the version ID of the latest message log.
// It uses GetObjectAttributes to retrieve the version ID without downloading the file.
func (m *MinIOStorageManager) GetLatestMessageVersionID() (versionID string, err error) {
	m.messagingMutex.RLock()
	defer m.messagingMutex.RUnlock()

	log.Debugf("Getting attributes for %s/%s", m.messagingBucketName, m.messagingObjectName)
	attributes, err := m.client.GetObjectAttributes(m.ctx, m.messagingBucketName, m.messagingObjectName)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			log.Debug("No message log found on server")
			return "", NoMessageLogFound
		}
		return "", err
	}

	log.Debugf("Latest message log version ID: %s", attributes.VersionID)
	return attributes.VersionID, nil
}
