package file

import (
	"github.com/ONSdigital/dp-dataset-exporter/observation"
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
func (store *Store) PutFile(reader io.Reader, filter *observation.Filter) (url string, err error) {

	session, err := session.NewSession(store.config)
	if err != nil {
		return "", err
	}

	uploader := s3manager.NewUploader(session)
	if err != nil {
		return "", err
	}

	filename := filter.JobID + ".csv"

	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: &store.bucket,
		Key:    &filename,
	})
	if err != nil {
		return "", err
	}

	return result.Location, nil
}