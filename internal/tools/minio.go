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

	bucketName := "gtfs"
	objectName := "GTFSSchedule.zip"

	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	return attributes.VersionID, nil
}

func GetLatestGTFSScheduleURL() (downloadURL *url.URL, versionID string, err error) {

	expiryTime := 5 * time.Minute
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

func PutLatestGTFSSchedule(reader io.Reader, fileSize int64) (versionID string, err error) {

	uploadInfo, err := c.PutObject(ctx, bucketName, objectName, reader, fileSize, minio.PutObjectOptions{ContentType: "application/zip"})
	if err != nil {
		return "", err
	}
	return uploadInfo.VersionID, nil
}
