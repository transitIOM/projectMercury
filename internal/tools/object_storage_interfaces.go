package tools

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/url"
	"time"
)

var (
	KeyNotFound         = "The specified key does not exist."
	NoGTFSScheduleFound = errors.New("no GTFS schedule found")
	NoMessageLogFound   = errors.New("no message log found")
)

// BucketInfo contains information about a storage bucket
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// ObjectInfo contains metadata about a stored object
type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
	ContentType  string
	VersionID    string
}

// UploadInfo contains information about an uploaded object
type UploadInfo struct {
	Bucket    string
	Key       string
	VersionID string
	ETag      string
}

// ObjectAttributes contains detailed attributes of an object
type ObjectAttributes struct {
	VersionID string
	ETag      string
}

// ObjectStorageClient defines the interface for interacting with object storage systems.
// This abstracts the underlying storage implementation (MinIO, S3, etc.) and makes
// the code more testable and flexible.
type ObjectStorageClient interface {
	// BucketExists checks if a bucket exists
	BucketExists(ctx context.Context, bucketName string) (bool, error)

	// MakeBucket creates a new bucket with optional region
	MakeBucket(ctx context.Context, bucketName string, region string) error

	// SetBucketVersioning enables or disables versioning for a bucket
	SetBucketVersioning(ctx context.Context, bucketName string, enabled bool) error

	// ListBuckets lists all buckets
	ListBuckets(ctx context.Context) ([]BucketInfo, error)

	// GetObject retrieves an object from storage
	// The returned io.ReadCloser must be closed by the caller
	GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error)

	// PutObject uploads an object to storage
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (UploadInfo, error)

	// StatObject gets metadata about an object
	StatObject(ctx context.Context, bucketName, objectName string) (ObjectInfo, error)

	// GetObjectAttributes retrieves object attributes including version ID
	GetObjectAttributes(ctx context.Context, bucketName, objectName string) (ObjectAttributes, error)

	// PresignedGetObject generates a presigned URL for downloading an object
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error)
}

// MessageLog represents a single entry in the message log
type MessageLog struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// NewMessage creates a new MessageLog entry with the current UTC timestamp
func NewMessage(msg string) MessageLog {
	return MessageLog{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Message:   msg,
	}
}

// GTFSStorage defines the interface for GTFS schedule storage operations.
// This interface provides high-level operations specific to GTFS data management.
type GTFSStorage interface {
	GetLatestGTFSVersionID() (versionID string, err error)

	GetLatestURL() (downloadURL *url.URL, versionID string, err error)

	PutSchedule(reader io.Reader, fileSize int64) (versionID string, err error)
}

// MessageStorage defines the interface for message log storage operations.
// This interface provides high-level operations for managing message logs.
type MessageStorage interface {
	AppendMessage(message *bytes.Buffer) (versionID string, err error)

	GetLatestLog() (messageLog *bytes.Buffer, err error)

	// GetLatestMessageVersionID returns the version ID of the latest message log
	GetLatestMessageVersionID() (versionID string, err error)
}

type ObjectStorageManager interface {
	GTFSStorage
	MessageStorage

	Initialize() error
	Close() error
}
