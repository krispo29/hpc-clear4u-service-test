package main

import (
	"context"
	dlog "log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"cloud.google.com/go/storage"
	"github.com/go-kit/log"

	"hpc-express-service/config"
	"hpc-express-service/database"
	"hpc-express-service/factory"
	"hpc-express-service/gcs"
	"hpc-express-service/server"
)

func main() {

	// sets the maximum number of CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// Set Logging
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Config
	config := config.LoadConfig()

	// GCS
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		dlog.Fatalf("Failed to create GCS client: %v", err)
	}
	_gcs := gcs.InitialGCSClient(config.GCSProjectID, config.GCSBucketName, gcsClient)

	// PostgreSQL
	postgreSQLConn, err := database.NewPostgreSQLConnection(
		config.PostgreSQLUser,
		config.PostgreSQLPassword,
		config.PostgreSQLName,
		config.PostgreSQLHost,
		config.PostgreSQLPort,
		config.PostgreSQLSSLMode,
	)

	if err != nil {
		logger.Log("DBconnection", "postgreSQLConn", err)
	}
	defer postgreSQLConn.Close()

	/*
		Repositories Factory
	*/
	repoFactory := factory.NewRepositoryFactory()

	/*
		Services Factory
	*/
	svcFactory := factory.NewServiceFactory(repoFactory, _gcs, config)

	/*
		Logging Factory
	*/
	factory.InitialLoggingFactory(logger, svcFactory)

	// Initial Server
	srv := server.New(
		svcFactory,
		postgreSQLConn,
		config.Mode,
	)

	// Gracefully Shutdown
	server := &http.Server{Addr: "0.0.0.0:" + config.Port, Handler: srv}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Log("graceful shutdown", "timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Log("graceful shutdown", err)
		}
		serverStopCtx()
	}()

	// Run the server
	logger.Log("transport", "http", "address", ":"+config.Port, "msg", "listening", "mode", config.Mode)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Log("ListenAndServe", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
