package config_test

import (
	"testing"
	"time"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpec(t *testing.T) {

	cfg, err := config.Get()

	Convey("Given an environment with no environment variables set", t, func() {

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func(c C) {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func(c C) {
				So(cfg.BindAddr, ShouldEqual, ":22500")
				So(cfg.KafkaAddr, ShouldResemble, []string{"localhost:9092"})
				So(cfg.FilterConsumerTopic, ShouldEqual, "filter-job-submitted")
				So(cfg.FilterConsumerGroup, ShouldEqual, "dp-dataset-exporter")
				So(cfg.FilterAPIURL, ShouldEqual, "http://localhost:22100")
				So(cfg.CSVExportedProducerTopic, ShouldEqual, "common-output-created")
				So(cfg.DatasetAPIURL, ShouldEqual, "http://localhost:22000")
				So(cfg.AWSRegion, ShouldEqual, "eu-west-1")
				So(cfg.S3BucketName, ShouldEqual, "csv-exported")
				So(cfg.S3BucketURL, ShouldEqual, "")
				So(cfg.S3PrivateBucketName, ShouldEqual, "csv-exported")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, time.Second*10)
				So(cfg.HealthCheckInterval, ShouldEqual, 10*time.Second)
				So(cfg.HealthCheckRecoveryInterval, ShouldEqual, time.Minute)
				So(cfg.VaultAddress, ShouldEqual, "http://localhost:8200")
				So(cfg.VaultPath, ShouldEqual, "secret/shared/psk")
				So(cfg.VaultToken, ShouldEqual, "")
				So(cfg.DownloadServiceURL, ShouldEqual, "http://localhost:23600")
				So(cfg.DownloadServiceInternalURL, ShouldEqual, "http://localhost:23600")
				So(cfg.ServiceAuthToken, ShouldEqual, "0f49d57b-c551-4d33-af1e-a442801dd851")
				So(cfg.StartupTimeout, ShouldEqual, 125*time.Second)
			})
		})
	})
}
