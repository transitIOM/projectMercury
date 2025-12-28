package tools

import (
	"bytes"
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
	c                   *minio.Client
	ctx                 context.Context
	gtfsBucketName      = "gtfs"
	gtfsObjectName      = "GTFSSchedule.zip"
	messagingBucketName = "messages"
	messagingObjectName = "messages.json"
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

	downloadURL, err = c.PresignedGetObject(ctx, gtfsBucketName, gtfsObjectName, expiryTime, reqParams)
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

	uploadInfo, err := c.PutObject(ctx, gtfsBucketName, gtfsObjectName, reader, fileSize, minio.PutObjectOptions{ContentType: "application/zip"})
	if err != nil {
		return "", err
	}
	return uploadInfo.VersionID, nil
}

func AppendMessage(b *bytes.Buffer, fileSize int64) (versionID string, err error) {

	r := bytes.NewReader(b.Bytes())

	uploadInfo, err := c.AppendObject(ctx, messagingBucketName, messagingObjectName, r, fileSize, minio.AppendObjectOptions{})
	if err != nil {
		return "", err
	}
	return uploadInfo.VersionID, nil
}

func GetLatestMessageLog() (messageLog bytes.Buffer, err error) {

	bucketName := messagingBucketName
	objectName := messagingObjectName

	r, err := c.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return bytes.Buffer{}, err
	}

	messageLog.ReadFrom(r)

	return messageLog, nil
}

func GetLatestMessageLogVersionID() (versionID string, err error) {

	bucketName := messagingBucketName
	objectName := messagingObjectName

	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	return attributes.VersionID, nil
}
