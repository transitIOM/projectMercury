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

	"github.com/joho/godotenv"
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

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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

func GetLatestGTFSScheduleVersionID() (versionID string, err error) {

	bucketName := gtfsBucketName
	objectName := gtfsObjectName

	gtfsMutex.RLock()
	defer gtfsMutex.RUnlock()

	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

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

	uploadInfo, err := c.PutObject(ctx, gtfsBucketName, gtfsObjectName, reader, fileSize, minio.PutObjectOptions{ContentType: "application/zip"})
	if err != nil {
		return "", err
	}
	return uploadInfo.VersionID, nil
}

func AppendMessage(b *bytes.Buffer) (versionID string, err error) {
	messagingMutex.Lock()
	defer messagingMutex.Unlock()
	existingData := bytes.Buffer{}

	_, err = c.StatObject(ctx, messagingBucketName, messagingObjectName, minio.StatObjectOptions{})
	if err == nil {
		r, err := c.GetObject(ctx, messagingBucketName, messagingObjectName, minio.GetObjectOptions{})
		if err == nil {
			defer r.Close()
			_, err = existingData.ReadFrom(r)
			if err != nil {
				return "", fmt.Errorf("failed to read existing message log: %w", err)
			}
		}
	} else {
		if minio.ToErrorResponse(err).Code != "NoSuchKey" {
			return "", fmt.Errorf("failed to check message log existence: %w", err)
		}
	}

	_, err = existingData.ReadFrom(b)
	if err != nil {
		return "", err
	}

	updatedReader := bytes.NewReader(existingData.Bytes())
	uploadInfo, err := c.PutObject(ctx, messagingBucketName, messagingObjectName, updatedReader, int64(existingData.Len()), minio.PutObjectOptions{
		ContentType: "text/jsonl",
	})

	if err != nil {
		return "", err
	}
	return uploadInfo.VersionID, nil
}

func GetLatestMessageLog() (messageLog bytes.Buffer, err error) {

	bucketName := messagingBucketName
	objectName := messagingObjectName

	messagingMutex.RLock()
	defer messagingMutex.RUnlock()

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

	return messageLog, nil
}

func GetLatestMessageLogVersionID() (versionID string, err error) {

	bucketName := messagingBucketName
	objectName := messagingObjectName

	messagingMutex.RLock()
	defer messagingMutex.RUnlock()

	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	return attributes.VersionID, nil
}
