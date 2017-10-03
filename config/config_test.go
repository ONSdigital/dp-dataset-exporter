package config_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/config"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestSpec(t *testing.T) {

	cfg, err := config.Get()

	Convey("Given an environment with no environment variables set", t, func() {

		Convey("When the config values are retrieved", func() {

			Convey("There should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("The values should be set to the expected defaults", func() {
				So(cfg.BindAddr, ShouldEqual, ":22500")
				So(cfg.KafkaAddr, ShouldResemble, []string{"localhost:9092"})
				So(cfg.FilterJobConsumerTopic, ShouldEqual, "filter-job-submitted")
				So(cfg.FilterJobConsumerGroup, ShouldEqual, "dp-dataset-exporter")
				So(cfg.DatabaseAddress, ShouldEqual, "bolt://localhost:7687")
				So(cfg.FilterAPIURL, ShouldEqual, "http://localhost:22100")
				So(cfg.FilterAPIAuthToken, ShouldEqual, "FD0108EA-825D-411C-9B1D-41EF7727F465")
				So(cfg.CSVExportedProducerTopic, ShouldEqual, "common-output-created")
				So(cfg.AWSRegion, ShouldEqual, "eu-west-1")
				So(cfg.S3BucketName, ShouldEqual, "csv-exported")
				So(cfg.GracefulShutdownTimeout, ShouldEqual, time.Second*10)
			})
		})
	})
}
