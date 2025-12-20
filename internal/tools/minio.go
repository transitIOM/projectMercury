package tools

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

var (
	c          *minio.Client
	ctx        context.Context
	bucketName = "gtfs"
	objectName = "GTFSSchedule.zip"
)

// init loads environment variables from a local .env file, reads MINIO_ACCESS_KEY, MINIO_SECRET_KEY, and MINIO_ENDPOINT, initializes the package context and MinIO client, and ensures the configured bucket exists with versioning enabled.
// It logs a fatal error and exits the process if loading the .env file or creating the MinIO client fails.
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	endpoint := os.Getenv("MINIO_ENDPOINT")

	ctx = context.Background()

	c, err = minio.New(endpoint, &minio.Options{
		Secure: false,
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatal(err)
	}

	bucketOpt := bucketOptions{
		name:              bucketName,
		makeBucketOptions: minio.MakeBucketOptions{},
		versioningConfig:  minio.BucketVersioningConfiguration{Status: "Enabled"},
	}
	makeBucket(bucketOpt)
}

type bucketOptions struct {
	name              string
	makeBucketOptions minio.MakeBucketOptions
	versioningConfig  minio.BucketVersioningConfiguration
}

// makeBucket ensures a bucket exists and applies the provided versioning configuration.
// It checks for a bucket named by options.name, creates it if it does not exist, and sets
// its versioning according to options.versioningConfig. Errors encountered while creating
// the bucket or setting versioning are logged but not returned.
func makeBucket(options bucketOptions) {

	exists, err := c.BucketExists(ctx, options.name)
	if exists == false {
		err = c.MakeBucket(ctx, options.name, options.makeBucketOptions)
		if err != nil {
			log.Error(err)
		}
	}

	err = c.SetBucketVersioning(ctx, options.name, options.versioningConfig)
	if err != nil {
		log.Error(err)
	}
}

// GetLatestGTFSScheduleVersionID retrieves the version ID of the "GTFSSchedule.zip" object in the "gtfs" bucket.
// It returns the object's VersionID, or an error if the object's attributes cannot be obtained.
func GetLatestGTFSScheduleVersionID() (versionID string, err error) {

	bucketName := "gtfs"
	objectName := "GTFSSchedule.zip"

	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	return attributes.VersionID, nil
}

// GetLatestGTFSScheduleURL returns a presigned GET URL for downloading the latest GTFSSchedule.zip and the object's version ID.
// The returned downloadURL is a presigned URL valid for 30 seconds that forces a ZIP file download named "GTFSSchedule.zip". The returned versionID is the latest VersionID of that object in the "gtfs" bucket. On failure, a non-nil error is returned.
func GetLatestGTFSScheduleURL() (downloadURL *url.URL, versionID string, err error) {

	bucketName := "gtfs"
	objectName := "GTFSSchedule.zip"
	expiryTime := 30 * time.Second
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%s", objectName))
	reqParams.Set("response-content-type", "application/zip")

	downloadURL, err = c.PresignedGetObject(ctx, bucketName, objectName, expiryTime, reqParams)
	if err != nil {
		return nil, "", err
	}

	ID, err := GetLatestGTFSScheduleVersionID()
	if err != nil {
		return nil, "", err
	}

	return downloadURL, ID, nil
}

// PutLatestGTFSSchedule uploads the provided reader as the latest GTFSSchedule.zip to the gtfs bucket.
// It returns the uploaded object's VersionID on success or an error if the upload fails.
func PutLatestGTFSSchedule(reader io.Reader, fileSize int64) (versionID string, err error) {

	uploadInfo, err := c.PutObject(ctx, bucketName, objectName, reader, fileSize, minio.PutObjectOptions{ContentType: "application/zip"})
	if err != nil {
		return "", err
	}
	return uploadInfo.VersionID, nil
}