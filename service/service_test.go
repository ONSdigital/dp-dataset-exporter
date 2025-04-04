package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-dataset-exporter/config"
	"github.com/ONSdigital/dp-dataset-exporter/file"
	"github.com/ONSdigital/dp-dataset-exporter/service"
	serviceMock "github.com/ONSdigital/dp-dataset-exporter/service/mock"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-graph/v2/mock"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v4"
	"github.com/ONSdigital/dp-kafka/v4/kafkatest"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errKafkaConsumer = errors.New("Kafka consumer error")
	errHealthcheck   = errors.New("healthCheck error")
)

var funcDoGetKafkaConsumerErr = func(ctx context.Context, kafkaCfg *config.KafkaConfig) (kafka.IConsumerGroup, error) {
	return nil, errKafkaConsumer
}

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

func TestRun(t *testing.T) {
	Convey("Having a set of mocked dependencies", t, func() {
		consumerMock := &kafkatest.IConsumerGroupMock{
			StartFunc:     func() error { return nil },
			LogErrorsFunc: func(ctx context.Context) {},
			CheckerFunc:   func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			ChannelsFunc:  func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
		}

		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		funcDoGetKafkaConsumerOk := func(ctx context.Context, kafkaCfg *config.KafkaConfig) (kafka.IConsumerGroup, error) {
			return consumerMock, nil
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetDatasetAPIClient := func(cfg *config.Config) service.DatasetAPI {
			return &serviceMock.DatasetAPIMock{}
		}

		funcDoGetFilterStoreClient := func(cfg *config.Config, serviceAuthToken string) service.FilterStore {
			return &serviceMock.FilterStoreMock{}
		}

		funcDoGetGraphDB := func(ctx context.Context) (*graph.DB, error) {
			return &graph.DB{
				Driver:      &mock.Mock{},
				CodeList:    nil,
				Hierarchy:   nil,
				Instance:    nil,
				Observation: nil,
				Dimension:   nil,
				Errors:      nil,
			}, nil

		}

		funcDoGetFileStore := func(ctx context.Context, cfg *config.Config) (fileStore *file.Store, err error) {
			f, err := file.NewStore(ctx, "", "", "", "", "")
			if err != nil {
				return nil, err
			}
			return f, nil
		}

		funcDoGetKafkaProducerOk := func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
			return &kafkatest.IProducerMock{
				ChannelsFunc: func() *kafka.ProducerChannels {
					return &kafka.ProducerChannels{}
				},
				LogErrorsFunc: func(ctx context.Context) {
					// Do nothing
				},
			}, nil
		}

		Convey("Given that initialising Kafka consumer returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHealthCheckFunc:      funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:       funcDoGetHTTPServerNil,
				DoGetKafkaConsumerFunc:    funcDoGetKafkaConsumerErr,
				DoGetKafkaProducerFunc:    funcDoGetKafkaProducerOk,
				DoGetDatasetAPIClientFunc: funcDoGetDatasetAPIClient,
				DoGetFilterStoreFunc:      funcDoGetFilterStoreClient,
				DoGetObservationStoreFunc: funcDoGetGraphDB,
				DoGetFileStoreFunc:        funcDoGetFileStore,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errKafkaConsumer)
				So(svcList.KafkaConsumer, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.KafkaProducer, ShouldBeFalse)
				So(svcList.FilterStore, ShouldBeFalse)
				So(svcList.DatasetAPI, ShouldBeFalse)
				So(svcList.ObservationStore, ShouldBeFalse)
				So(svcList.FileStore, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:       funcDoGetHTTPServerNil,
				DoGetHealthCheckFunc:      funcDoGetHealthcheckErr,
				DoGetKafkaConsumerFunc:    funcDoGetKafkaConsumerOk,
				DoGetKafkaProducerFunc:    funcDoGetKafkaProducerOk,
				DoGetDatasetAPIClientFunc: funcDoGetDatasetAPIClient,
				DoGetFilterStoreFunc:      funcDoGetFilterStoreClient,
				DoGetObservationStoreFunc: funcDoGetGraphDB,
				DoGetFileStoreFunc:        funcDoGetFileStore,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.KafkaConsumer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
				So(svcList.KafkaProducer, ShouldBeTrue)
				So(svcList.FilterStore, ShouldBeTrue)
				So(svcList.DatasetAPI, ShouldBeTrue)
				So(svcList.ObservationStore, ShouldBeTrue)
				So(svcList.FileStore, ShouldBeTrue)
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc:       funcDoGetHTTPServer,
				DoGetHealthCheckFunc:      funcDoGetHealthcheckOk,
				DoGetKafkaConsumerFunc:    funcDoGetKafkaConsumerOk,
				DoGetDatasetAPIClientFunc: funcDoGetDatasetAPIClient,
				DoGetFilterStoreFunc:      funcDoGetFilterStoreClient,
				DoGetObservationStoreFunc: funcDoGetGraphDB,
				DoGetFileStoreFunc:        funcDoGetFileStore,
				DoGetKafkaProducerFunc:    funcDoGetKafkaProducerOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.KafkaConsumer, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeTrue)
				So(svcList.FilterStore, ShouldBeTrue)
				So(svcList.DatasetAPI, ShouldBeTrue)
				So(svcList.ObservationStore, ShouldBeTrue)
				So(svcList.FileStore, ShouldBeTrue)
			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 8)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Kafka consumer")
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":22500")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &serviceMock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetKafkaConsumerFunc:    funcDoGetKafkaConsumerOk,
				DoGetKafkaProducerFunc:    funcDoGetKafkaProducerOk,
				DoGetFilterStoreFunc:      funcDoGetFilterStoreClient,
				DoGetDatasetAPIClientFunc: funcDoGetDatasetAPIClient,
				DoGetObservationStoreFunc: funcDoGetGraphDB,
				DoGetFileStoreFunc:        funcDoGetFileStore,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
				So(svcList.HealthCheck, ShouldBeTrue)
				So(svcList.KafkaConsumer, ShouldBeTrue)
				So(svcList.KafkaProducer, ShouldBeTrue)
				So(svcList.FilterStore, ShouldBeTrue)
				So(svcList.DatasetAPI, ShouldBeTrue)
				So(svcList.ObservationStore, ShouldBeTrue)
				So(svcList.FileStore, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 8)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Kafka consumer")
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Having a correctly initialised service", t, func() {
		hcStopped := false

		consumerMock := &kafkatest.IConsumerGroupMock{
			StartFunc:     func() error { return nil },
			StopFunc:      func() error { return nil },
			LogErrorsFunc: func(ctx context.Context) {},
			CloseFunc:     func(ctx context.Context, optFuncs ...kafka.OptFunc) error { return nil },
			CheckerFunc:   func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			ChannelsFunc:  func() *kafka.ConsumerGroupChannels { return &kafka.ConsumerGroupChannels{} },
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &serviceMock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &serviceMock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				return nil
			},
		}

		funcDoGetDatasetAPIClient := func(cfg *config.Config) service.DatasetAPI {
			return &serviceMock.DatasetAPIMock{}
		}

		funcDoGetFilterStoreClient := func(cfg *config.Config, serviceAuthToken string) service.FilterStore {
			return &serviceMock.FilterStoreMock{}
		}

		funcDoGetGraphDB := func(ctx context.Context) (*graph.DB, error) {
			return &graph.DB{
				Driver:      &mock.Mock{},
				CodeList:    nil,
				Hierarchy:   nil,
				Instance:    nil,
				Observation: nil,
				Dimension:   nil,
				Errors:      nil,
			}, nil

		}

		funcDoGetFileStore := func(ctx context.Context, cfg *config.Config) (fileStore *file.Store, err error) {
			f, err := file.NewStore(ctx, "", "", "", "", "")
			if err != nil {
				return nil, err
			}
			return f, nil
		}

		funcDoGetKafkaProducerOk := func(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
			return &kafkatest.IProducerMock{
				ChannelsFunc: func() *kafka.ProducerChannels {
					return &kafka.ProducerChannels{}
				},
				LogErrorsFunc: func(ctx context.Context) {
					// Do nothing
				},
			}, nil
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetKafkaConsumerFunc: func(ctx context.Context, kafkaCfg *config.KafkaConfig) (kafka.IConsumerGroup, error) {
					return consumerMock, nil
				},
				DoGetKafkaProducerFunc:    funcDoGetKafkaProducerOk,
				DoGetFilterStoreFunc:      funcDoGetFilterStoreClient,
				DoGetDatasetAPIClientFunc: funcDoGetDatasetAPIClient,
				DoGetObservationStoreFunc: funcDoGetGraphDB,
				DoGetFileStoreFunc:        funcDoGetFileStore,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StartCalls()), ShouldEqual, 1)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(consumerMock.StartCalls()), ShouldEqual, 1)
			So(len(consumerMock.CheckerCalls()), ShouldEqual, 0)
			So(len(consumerMock.ChannelsCalls()), ShouldEqual, 2)
			So(len(consumerMock.LogErrorsCalls()), ShouldEqual, 1)
			So(len(consumerMock.CloseCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			failingserverMock := &serviceMock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &serviceMock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return failingserverMock },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetKafkaConsumerFunc: func(ctx context.Context, kafkaCfg *config.KafkaConfig) (kafka.IConsumerGroup, error) {
					return consumerMock, nil
				},
				DoGetKafkaProducerFunc:    funcDoGetKafkaProducerOk,
				DoGetFilterStoreFunc:      funcDoGetFilterStoreClient,
				DoGetDatasetAPIClientFunc: funcDoGetDatasetAPIClient,
				DoGetObservationStoreFunc: funcDoGetGraphDB,
				DoGetFileStoreFunc:        funcDoGetFileStore,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(len(hcMock.StartCalls()), ShouldEqual, 1)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(consumerMock.CloseCalls()), ShouldEqual, 1)
			So(len(consumerMock.CloseCalls()), ShouldEqual, 1)
			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
		})
	})
}
