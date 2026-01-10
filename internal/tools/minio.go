package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

var (
	c                   *minio.Client
	ctx                 context.Context
	gtfsBucketName      = "gtfs"
	gtfsObjectName      = "GTFSSchedule.zip"
	gtfsMutex           = sync.RWMutex{}
	messagingBucketName = "messages"
	messagingObjectName = "messages.jsonl"
	messagingMutex      = sync.RWMutex{}
)

func InitializeMinio() {
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	endpoint := os.Getenv("MINIO_ENDPOINT")

	ctx = context.Background()

	for {
		var err error
		c, err = minio.New(endpoint, &minio.Options{
			Secure: false,
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		})

		if err == nil {
			_, err = c.ListBuckets(ctx)
		}

		if err != nil {
			log.Errorf("Minio connection failed: %v. Retrying in 1 minute...", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		log.Info("Successfully connected to Minio")
		break
	}

	gtfsBucketOpt := bucketOptions{
		name:              gtfsBucketName,
		makeBucketOptions: minio.MakeBucketOptions{},
		versioningConfig:  minio.BucketVersioningConfiguration{Status: "Enabled"},
	}
	makeBucket(gtfsBucketOpt)

	messagingBucketOpt := bucketOptions{
		name:              messagingBucketName,
		makeBucketOptions: minio.MakeBucketOptions{},
		versioningConfig:  minio.BucketVersioningConfiguration{Status: "Enabled"},
	}
	makeBucket(messagingBucketOpt)
}

type bucketOptions struct {
	name              string
	makeBucketOptions minio.MakeBucketOptions
	versioningConfig  minio.BucketVersioningConfiguration
}

func makeBucket(options bucketOptions) {

	exists, err := c.BucketExists(ctx, options.name)
	if err != nil {
		log.Error(err)
	}
	if !exists {
		log.Debugf("Creating bucket: %s", options.name)
		err = c.MakeBucket(ctx, options.name, options.makeBucketOptions)
		if err != nil {
			log.Error(err)
		}
	} else {
		log.Debugf("Bucket already exists: %s", options.name)
	}

	log.Debugf("Setting versioning for bucket: %s", options.name)
	err = c.SetBucketVersioning(ctx, options.name, options.versioningConfig)
	if err != nil {
		log.Error(err)
	}
}

func GetLatestGTFSScheduleVersionID() (versionID string, err error) {

	bucketName := gtfsBucketName
	objectName := gtfsObjectName

	gtfsMutex.RLock()
	defer gtfsMutex.RUnlock()

	log.Debugf("Getting attributes for %s/%s", bucketName, objectName)
	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	log.Debugf("Latest GTFS version ID: %s", attributes.VersionID)
	return attributes.VersionID, nil
}

func GetLatestGTFSScheduleURL() (downloadURL *url.URL, versionID string, err error) {

	expiryTime := 5 * time.Minute
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=%s", gtfsObjectName))
	reqParams.Set("response-content-type", "application/zip")

	gtfsMutex.RLock()
	downloadURL, err = c.PresignedGetObject(ctx, gtfsBucketName, gtfsObjectName, expiryTime, reqParams)
	gtfsMutex.RUnlock()

	if err != nil {
		return nil, "", err
	}

	ID, err := GetLatestGTFSScheduleVersionID()
	if err != nil {
		return nil, "", err
	}

	return downloadURL, ID, nil
}

func PutLatestGTFSSchedule(reader io.Reader, fileSize int64) (versionID string, err error) {

	gtfsMutex.Lock()
	defer gtfsMutex.Unlock()

	log.Debugf("Uploading %s to %s, size: %d", gtfsObjectName, gtfsBucketName, fileSize)
	uploadInfo, err := c.PutObject(ctx, gtfsBucketName, gtfsObjectName, reader, fileSize, minio.PutObjectOptions{ContentType: "application/zip"})
	if err != nil {
		return "", err
	}
	log.Debugf("Successfully uploaded %s, version ID: %s", gtfsObjectName, uploadInfo.VersionID)
	return uploadInfo.VersionID, nil
}

func AppendMessage(b *bytes.Buffer) (versionID string, err error) {
	messagingMutex.Lock()
	defer messagingMutex.Unlock()
	existingData := bytes.Buffer{}

	log.Debugf("Checking if %s exists in %s", messagingObjectName, messagingBucketName)
	_, err = c.StatObject(ctx, messagingBucketName, messagingObjectName, minio.StatObjectOptions{})
	if err == nil {
		log.Debugf("Retrieving existing %s", messagingObjectName)
		r, err := c.GetObject(ctx, messagingBucketName, messagingObjectName, minio.GetObjectOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to get existing message log: %w", err)
		}
		defer func(r *minio.Object) {
			err := r.Close()
			if err != nil {
				log.Error(err)
			}
		}(r)
		_, err = existingData.ReadFrom(r)
		if err != nil {
			return "", fmt.Errorf("failed to read existing message log: %w", err)
		}
	} else {
		if minio.ToErrorResponse(err).Code != "NoSuchKey" {
			return "", fmt.Errorf("failed to check message log existence: %w", err)
		}
		log.Debug("Message log does not exist, starting new one")
	}

	log.Debug("Appending new message to existing data")
	_, err = existingData.ReadFrom(b)
	if err != nil {
		return "", err
	}

	updatedReader := bytes.NewReader(existingData.Bytes())
	log.Debugf("Uploading updated %s to %s, total size: %d", messagingObjectName, messagingBucketName, existingData.Len())
	uploadInfo, err := c.PutObject(ctx, messagingBucketName, messagingObjectName, updatedReader, int64(existingData.Len()), minio.PutObjectOptions{
		ContentType: "text/jsonl",
	})

	if err != nil {
		return "", err
	}
	log.Debugf("Successfully updated %s, version ID: %s", messagingObjectName, uploadInfo.VersionID)
	return uploadInfo.VersionID, nil
}

func GetLatestMessageLog() (messageLog bytes.Buffer, err error) {

	bucketName := messagingBucketName
	objectName := messagingObjectName

	messagingMutex.RLock()
	defer messagingMutex.RUnlock()

	log.Debugf("Retrieving %s from %s", objectName, bucketName)
	r, err := c.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return bytes.Buffer{}, err
	}
	defer func(r *minio.Object) {
		err := r.Close()
		if err != nil {
			log.Error(err)
		}
	}(r)

	_, err = messageLog.ReadFrom(r)
	if err != nil {
		return bytes.Buffer{}, err
	}

	log.Debugf("Successfully retrieved message log, size: %d bytes", messageLog.Len())
	return messageLog, nil
}

func GetLatestMessageLogVersionID() (versionID string, err error) {

	bucketName := messagingBucketName
	objectName := messagingObjectName

	messagingMutex.RLock()
	defer messagingMutex.RUnlock()

	log.Debugf("Getting attributes for %s/%s", bucketName, objectName)
	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	log.Debugf("Latest message log version ID: %s", attributes.VersionID)
	return attributes.VersionID, nil
}
