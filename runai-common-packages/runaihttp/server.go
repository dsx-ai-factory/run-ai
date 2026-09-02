/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaihttp

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

// ServerConfig holds configuration for the HTTP serve function
type ServerConfig struct {
	ServicePort     int           `mapstructure:"service_port"`            // ServicePort is the port of the service
	WriteTimeout    time.Duration `mapstructure:"server_write_timeout"`    // WriteTimeout is the maximum duration before timing out writes of the response
	ReadTimeout     time.Duration `mapstructure:"server_read_timeout"`     // ReadTimeout is the maximum duration for reading the entire request, including the body.
	IdleTimeout     time.Duration `mapstructure:"server_idle_timeout"`     // IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	ShutdownTimeout time.Duration `mapstructure:"server_shutdown_timeout"` // ShutdownTimeout is the maximum duration for graceful server shutdown
	IPAddress       string        `mapstructure:"server_ip_address"`       // IPAddress to bind the server to
	HealthHost      string        `mapstructure:"server_health_host"`      // HealthHost to use for health probes
	HealthEndpoint  string        `mapstructure:"server_health_endpoint"`  // HealthEndpoint to use for health probes
}

// SetDefaults sets default values for the ServerConfig
func (c *ServerConfig) SetDefaults() {
	c.ServicePort = 8080
	c.WriteTimeout = time.Second * 15
	c.ReadTimeout = time.Second * 15
	c.IdleTimeout = time.Second * 60
	c.ShutdownTimeout = time.Second * 15
	c.IPAddress = "0.0.0.0"
	c.HealthHost = "localhost"
	c.HealthEndpoint = "/"
}

func readConfig() *ServerConfig {
	cfg, err := config.GetR[*ServerConfig]()
	if err != nil {
		logging.Fatalf("failed to initialize server configuration, error: %v", err)
	}
	return cfg
}

func init() {
	logging.Init(nil)
	executeFlags()
}

// executeFlags - Parser supported extended flags of the service and executes accordingly
func executeFlags() {
	flagset := flag.NewFlagSet("server", flag.ContinueOnError)
	healthFlag := flagset.Bool("health", false, "perform health check against configured host and port")
	_ = flagset.Parse(os.Args[1:])
	if !*healthFlag {
		return
	}

	cfg := readConfig()

	resp, err := http.Get("http://" + cfg.HealthHost + ":" + strconv.Itoa(cfg.ServicePort) + cfg.HealthEndpoint) //nolint:noctx // CLI health-check helper; context propagation adds no value here
	if err != nil {
		logging.Errorf("health check failed for host %s port %d", cfg.HealthHost, cfg.ServicePort)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logging.Errorf("health check failed for host %s port %d", cfg.HealthHost, cfg.ServicePort)
		os.Exit(1)
	}

	logging.Infof("health check passed for host %s port %d", cfg.HealthHost, cfg.ServicePort)

	os.Exit(0)
}

// ServeHTTP start an HTTP server that listing and serving request
func ServeHTTP(router chi.Router, configuration *ServerConfig) {
	if configuration == nil {
		configuration = readConfig()
	}
	addr := configuration.IPAddress + ":" + strconv.Itoa(configuration.ServicePort)
	logging.Infof("setting up router on address %s", addr)

	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: configuration.WriteTimeout,
		ReadTimeout:  configuration.ReadTimeout,
		IdleTimeout:  configuration.IdleTimeout,
		Handler:      router, // Pass our instance of chi.Router in.
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logging.Fatalf("Error while starting server, %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), configuration.ShutdownTimeout)
	//ctx, cancel := context.WithTimeout(context.Background(), configuration.ShutdownTimeout)
	defer cancel()
	logging.Infof("shutting down")
	_ = srv.Shutdown(ctx)
}
