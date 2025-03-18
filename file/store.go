package file

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3/v3"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out filetest/uploader.go -pkg filetest . Uploader
//go:generate moq -out filetest/vault.go -pkg filetest . VaultClient

// Uploader represents the methods required to upload to s3 with and without encryption
type Uploader interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, options ...func(*manager.Uploader)) (*manager.UploadOutput, error)
	Config() aws.Config
	BucketName() string
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// VaultClient is an interface to represent methods called to action upon vault
type VaultClient interface {
	WriteKey(path, key, value string) error
}

// Store provides file storage via S3.
type Store struct {
	Uploader        *s3client.Client
	PrivateUploader *s3client.Client
	PublicURL       string
	PublicBucket    string
	privateBucket   string
}

// NewStore returns a new store instance for the given AWS region and S3 bucket name.
func NewStore(
	ctx context.Context,
	region,
	publicURL,
	publicBucket,
	privateBucket,
	localstackHost string,
) (*Store, error) {

	uploader, err := s3client.NewClient(ctx, region, publicBucket)

	if err != nil {
		return nil, err
	}

	var privateUploader *s3client.Client

	if localstackHost != "" {
		awsConfig, err := awsConfig.LoadDefaultConfig(ctx,
			awsConfig.WithRegion(region),
			awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		)
		if err != nil {
			return nil, err
		}

		privateUploader = s3client.NewClientWithConfig(privateBucket, awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(localstackHost)
			o.UsePathStyle = true
		})
	} else {
		privateUploader = s3client.NewClientWithConfig(privateBucket, uploader.Config())
	}

	return &Store{
		Uploader:        uploader,
		PrivateUploader: privateUploader,
		PublicURL:       publicURL,
		PublicBucket:    publicBucket,
		privateBucket:   privateBucket,
	}, nil
}

// PutFile stores the contents of the given reader to a csv file of given the supplied name.
func (store *Store) PutFile(ctx context.Context, reader io.Reader, filename string, isPublished bool) (uploadedFileURL string, err error) {
	var result *manager.UploadOutput

	if isPublished {
		log.Info(ctx, "uploading public file to S3", log.Data{
			"bucket": store.PublicBucket,
			"name":   filename,
		})

		result, err = store.Uploader.Upload(ctx, &s3.PutObjectInput{
			Body:   reader,
			Bucket: &store.PublicBucket,
			Key:    &filename,
		})
		if err != nil {
			return "", err
		}
		if store.PublicURL != "" {
			return fmt.Sprintf("%s/%s", store.PublicURL, filename), nil
		}
	} else {
		log.Info(ctx, "uploading private file to S3", log.Data{
			"bucket": store.privateBucket,
			"name":   filename,
		})

		result, err = store.PrivateUploader.Upload(ctx, &s3.PutObjectInput{
			Body:   reader,
			Bucket: &store.privateBucket,
			Key:    &filename,
		})
		if err != nil {
			return "", err
		}
	}

	return url.PathUnescape(result.Location)
}
