package config_test

import (
	"github.com/ONSdigital/dp-dataset-exporter/config"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
				So(cfg.KafkaAddr, ShouldEqual, "localhost:9092")
				So(cfg.FilterJobConsumerTopic, ShouldEqual, "filter-job-submitted")
				So(cfg.FilterJobConsumerGroup, ShouldEqual, "dp-dataset-exporter")
				So(cfg.DatabaseAddress, ShouldEqual, "bolt://localhost:7687")
				So(cfg.FilterAPIURL, ShouldEqual, "http://localhost:22100")
				So(cfg.ResultProducerTopic, ShouldEqual, "csv-filtered")
			})
		})
	})
}
