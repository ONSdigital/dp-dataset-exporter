package file

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

// Store provides file storage via S3.
type Store struct {
	config *aws.Config
	bucket string
}

// NewStore returns a new store instance for the given AWS region and S3 bucket name.
func NewStore(region string, bucket string) *Store {

	config := aws.NewConfig().WithRegion(region)

	return &Store{
		config: config,
		bucket: bucket,
	}
}

// PutFile stores the contents of the given reader to the given filename.
func (store *Store) PutFile(reader io.Reader, filename string) error {

	session, err := session.NewSession(store.config)
	uploader := s3manager.NewUploader(session)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: &store.bucket,
		Key:    &filename,
	})
	if err != nil {
		return err
	}

	log.Debug("upload successful", log.Data{
		"upload_location": result.Location,
	})

	return nil
}
