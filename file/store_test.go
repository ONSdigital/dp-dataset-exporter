package file

import (
	"errors"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/file/filetest"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	publicTestLocation  = "https://csv-exported/myfile.csv"
	privateTestLocation = "s3://csv-exported/myfile.csv"
)

func TestPutFileErrorScenarios(t *testing.T) {
	Convey("Given a store with an uploader that returns an error", t, func() {
		uploaderMock := &filetest.UploaderMock{}
		uploaderMock.UploadFunc = func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return nil, errors.New("uploader error")
		}

		store := &Store{uploader: uploaderMock}

		Convey("When PutFile is called for a published version", func() {
			url, err := store.PutFile(strings.NewReader(""), "", true)

			Convey("Then the correct error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "uploader error")
				So(url, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a store with a vaultClient that returns an error", t, func() {
		vaultClientMock := &filetest.VaultClientMock{}
		vaultClientMock.WriteKeyFunc = func(string, string, string) error {
			return errors.New("vault client error")
		}

		store := &Store{vaultClient: vaultClientMock}

		Convey("When PutFile is called for an unpublished version", func() {
			url, err := store.PutFile(strings.NewReader(""), "", false)

			Convey("Then the correct error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "vault client error")
				So(url, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a store with a cryptoUploader that returns an error", t, func() {
		vaultClientMock := &filetest.VaultClientMock{}
		vaultClientMock.WriteKeyFunc = func(string, string, string) error {
			return nil
		}

		cryptoUploaderMock := &filetest.CryptoUploaderMock{}
		cryptoUploaderMock.UploadWithPSKFunc = func(*s3manager.UploadInput, []byte) (*s3manager.UploadOutput, error) {
			return nil, errors.New("crypto uploader error")
		}

		store := &Store{vaultClient: vaultClientMock, cryptoUploader: cryptoUploaderMock}

		Convey("When PutFile is called for an unpublished version", func() {
			url, err := store.PutFile(strings.NewReader(""), "", false)

			Convey("Then the correct error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "crypto uploader error")
				So(url, ShouldBeEmpty)
			})
		})
	})
}

func TestPutFileSuccessSceanarios(t *testing.T) {
	Convey("Given a store exists with a valid uploader", t, func() {
		uploaderMock := &filetest.UploaderMock{}
		uploaderMock.UploadFunc = func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
			return &s3manager.UploadOutput{Location: publicTestLocation}, nil
		}

		store := &Store{uploader: uploaderMock}

		Convey("When PutFile is called for a published version", func() {
			url, err := store.PutFile(strings.NewReader(""), "", true)

			Convey("Then the file location should be returned", func() {
				So(err, ShouldBeNil)
				So(url, ShouldEqual, publicTestLocation)
			})
		})
	})

	Convey("Given a store with a valid vault client and cryptoUploader", t, func() {
		vaultClientMock := &filetest.VaultClientMock{}
		vaultClientMock.WriteKeyFunc = func(string, string, string) error {
			return nil
		}

		cryptoUploaderMock := &filetest.CryptoUploaderMock{}
		cryptoUploaderMock.UploadWithPSKFunc = func(*s3manager.UploadInput, []byte) (*s3manager.UploadOutput, error) {
			return &s3manager.UploadOutput{Location: privateTestLocation}, nil
		}

		store := &Store{vaultClient: vaultClientMock, cryptoUploader: cryptoUploaderMock}

		Convey("When PutFile is called for an unpublished version", func() {
			url, err := store.PutFile(strings.NewReader(""), "", false)

			Convey("Then the correct location is returned", func() {
				So(err, ShouldBeNil)
				So(url, ShouldEqual, privateTestLocation)
			})
		})
	})
}
