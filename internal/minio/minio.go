package minio

import (
	"context"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

var (
	c   *minio.Client
	ctx context.Context
)

func init() {

	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	endpoint := os.Getenv("MINIO_ENDPOINT")

	ctx = context.Background()

	var err error
	c, err = minio.New(endpoint, &minio.Options{
		Secure: false,
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatal(err)
	}

	bucketOpt := bucketOptions{
		name:              "timetables",
		makeBucketOptions: minio.MakeBucketOptions{},
		versioningConfig:  minio.BucketVersioningConfiguration{Status: "Enabled"},
	}
	makeBucket(bucketOpt)
}

type bucketOptions struct {
	name              string
	ctx               context.Context
	makeBucketOptions minio.MakeBucketOptions
	versioningConfig  minio.BucketVersioningConfiguration
}

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

func GetLatestVersionID(bucketName string, objectName string) (versionID string, err error) {

	attributes, err := c.GetObjectAttributes(ctx, bucketName, objectName, minio.ObjectAttributesOptions{})
	if err != nil {
		return "", err
	}

	return attributes.VersionID, nil
}
