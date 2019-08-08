package file

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/s3crypto"
)

//go:generate moq -out uploader_moq_test.go . Uploader
//go:generate moq -out cryptouploader_moq_test.go . CryptoUploader
//go:generate moq -out vault_moq_test.go . VaultClient

// Uploader represents the methods required to upload to s3 without encryption
type Uploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// CryptoUploader represents the methods required to upload to s3 with encryption
type CryptoUploader interface {
	UploadWithPSK(input *s3manager.UploadInput, psk []byte) (*s3manager.UploadOutput, error)
}

// VaultClient is an interface to represent methods called to action upon vault
type VaultClient interface {
	WriteKey(path, key, value string) error
}

// Store provides file storage via S3.
type Store struct {
	uploader       Uploader
	cryptoUploader CryptoUploader
	publicBucket   string
	privateBucket  string
	vaultPath      string
	vaultClient    VaultClient
}

// NewStore returns a new store instance for the given AWS region and S3 bucket name.
func NewStore(
	region,
	publicBucket,
	privateBucket,
	vaultPath string,
	vaultClient VaultClient) (*Store, error) {

	config := aws.NewConfig().WithRegion(region)

	session, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	return &Store{
		uploader:       s3manager.NewUploader(session),
		cryptoUploader: s3crypto.NewUploader(session, &s3crypto.Config{HasUserDefinedPSK: true}),
		publicBucket:   publicBucket,
		privateBucket:  privateBucket,
		vaultPath:      vaultPath,
		vaultClient:    vaultClient,
	}, nil
}

// PutFile stores the contents of the given reader to a csv file of given the supplied name.
func (store *Store) PutFile(reader io.Reader, filename string, isPublished bool) (uploadedFileURL string, err error) {
	var result *s3manager.UploadOutput

	if isPublished {
		log.Info("uploading public file to S3", log.Data{
			"bucket": store.publicBucket,
			"name":   filename,
		})

		result, err = store.uploader.Upload(&s3manager.UploadInput{
			Body:   reader,
			Bucket: &store.publicBucket,
			Key:    &filename,
		})
		if err != nil {
			return "", err
		}
	} else {
		log.Info("uploading private file to S3", log.Data{
			"bucket": store.privateBucket,
			"name":   filename,
		})

		psk := createPSK()
		vaultPath := store.vaultPath + "/" + path.Base(filename)
		vaultKey := "key"

		log.Info("writing key to vault", log.Data{
			"vault_path": vaultPath,
		})
		if err := store.vaultClient.WriteKey(vaultPath, vaultKey, hex.EncodeToString(psk)); err != nil {
			return "", err
		}

		result, err = store.cryptoUploader.UploadWithPSK(&s3manager.UploadInput{
			Body:   reader,
			Bucket: &store.privateBucket,
			Key:    &filename,
		}, psk)
		if err != nil {
			return "", err
		}

		log.Info("writing key to vault", log.Data{
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
