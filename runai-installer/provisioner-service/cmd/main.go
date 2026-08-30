package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/router"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 90 * time.Second
	shutdownTimeout   = 15 * time.Second
)

func init() {
	logging.Init(nil)
	secretsPath := os.Getenv("SECRETS_MOUNT_PATH")
	if secretsPath == "" {
		secretsPath = "/etc/secrets"
	}
	config.Init(secretsPath)
}

func main() {
	serverConfig := runai_config.Get[*config.Server]()
	tlsConfig := runai_config.Get[*config.TLS]()
	jfrogConfig := runai_config.Get[*config.JFrog]()

	logging.Infof("Configuration loaded - Port: %d", serverConfig.Port)
	logging.Debugf("TLS paths configured: cert=%s, key=%s", tlsConfig.CertPath, tlsConfig.KeyPath)

	logging.Debugf("JFrog token loaded: %t", jfrogConfig.Token != "")

	r := router.Create()

	port := serverConfig.Port
	addr := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	_ = srv.Shutdown(ctx)
}
