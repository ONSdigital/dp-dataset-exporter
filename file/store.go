package file

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	s3client "github.com/ONSdigital/dp-s3"
	"github.com/ONSdigital/log.go/log"
)

//go:generate moq -out filetest/uploader.go -pkg filetest . Uploader
//go:generate moq -out filetest/vault.go -pkg filetest . VaultClient

// Uploader represents the methods required to upload to s3 with and without encryption
type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	UploadWithPSK(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error)
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
	Uploader       Uploader
	CryptoUploader Uploader
	PublicURL      string
	PublicBucket   string
	privateBucket  string
	VaultPath      string
	VaultClient    VaultClient
}

// NewStore returns a new store instance for the given AWS region and S3 bucket name.
func NewStore(
	region,
	publicURL,
	publicBucket,
	privateBucket,
	vaultPath string,
	vaultClient VaultClient,
) (*Store, error) {

	uploader, err := s3client.NewUploader(region, publicBucket, false)
	if err != nil {
		return nil, err
	}

	cryptoUploader := s3client.NewUploaderWithSession(region, privateBucket, true, uploader.Session())

	return &Store{
		Uploader:       uploader,
		CryptoUploader: cryptoUploader,
		PublicURL:      publicURL,
		PublicBucket:   publicBucket,
		privateBucket:  privateBucket,
		VaultPath:      vaultPath,
		VaultClient:    vaultClient,
	}, nil
}

// PutFile stores the contents of the given reader to a csv file of given the supplied name.
func (store *Store) PutFile(reader io.Reader, filename string, isPublished bool) (uploadedFileURL string, err error) {
	ctx := context.Background()
	var result *s3manager.UploadOutput

	if isPublished {
		log.Event(ctx, "uploading public file to S3", log.Data{
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
		log.Event(ctx, "uploading private file to S3", log.Data{
			"bucket": store.privateBucket,
			"name":   filename,
		})

		psk := createPSK()
		vaultPath := store.VaultPath + "/" + path.Base(filename)
		vaultKey := "key"

		log.Event(ctx, "writing key to vault", log.Data{
			"vault_path": vaultPath,
		})
		if err := store.VaultClient.WriteKey(vaultPath, vaultKey, hex.EncodeToString(psk)); err != nil {
			return "", err
		}

		result, err = store.CryptoUploader.UploadWithPSK(&s3manager.UploadInput{
			Body:   reader,
			Bucket: &store.privateBucket,
			Key:    &filename,
		}, psk)
		if err != nil {
			return "", err
		}

		log.Event(ctx, "writing key to vault", log.Data{
			"result.Location": result.Location,
			"vault_path":      vaultPath,
		})
	}

	return url.PathUnescape(result.Location)
}

func createPSK() []byte {
	key := make([]byte, 16)
	rand.Read(key)

	return key
}
