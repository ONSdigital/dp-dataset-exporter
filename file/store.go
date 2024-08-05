package file

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out filetest/uploader.go -pkg filetest . Uploader
//go:generate moq -out filetest/vault.go -pkg filetest . VaultClient

// Uploader represents the methods required to upload to s3 with and without encryption
type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	Session() *session.Session
	BucketName() string
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// VaultClient is an interface to represent methods called to action upon vault
type VaultClient interface {
	WriteKey(path, key, value string) error
}

// Store provides file storage via S3.
type Store struct {
	Uploader        Uploader
	PrivateUploader Uploader
	PublicURL       string
	PublicBucket    string
	privateBucket   string
}

// NewStore returns a new store instance for the given AWS region and S3 bucket name.
func NewStore(
	region,
	publicURL,
	publicBucket,
	privateBucket string,
	localstackHost string,
) (*Store, error) {

	uploader, err := s3client.NewUploader(region, publicBucket)
	if err != nil {
		return nil, err
	}

	var privateUploader *s3client.Uploader

	if localstackHost != "" {
		s, err := session.NewSession(&aws.Config{
			Endpoint:         aws.String(localstackHost),
			Region:           aws.String(region),
			S3ForcePathStyle: aws.Bool(true),
			Credentials:      credentials.NewStaticCredentials("test", "test", ""),
		})

		if err != nil {
			return nil, err
		}

		privateUploader = s3client.NewUploaderWithSession(privateBucket, s)
	} else {
		privateUploader = s3client.NewUploaderWithSession(privateBucket, uploader.Session())
	}

	if err != nil {
		return nil, err
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
	var result *s3manager.UploadOutput

	if isPublished {
		log.Info(ctx, "uploading public file to S3", log.Data{
			"bucket": store.PublicBucket,
			"name":   filename,
		})

		result, err = store.Uploader.Upload(&s3manager.UploadInput{
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

		result, err = store.PrivateUploader.Upload(&s3manager.UploadInput{
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
