package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/faircom/prometheus-exporter/pkg/collector"
	"github.com/faircom/prometheus-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Version is set at build time
	Version   = "dev"
	BuildTime = "unknown"

	// Logger instance
	logger = logrus.New()
)

func setupLogging(cfg *config.Config) error {
	// Set log level
	level, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.Log.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Setup file output with rotation
	logDir := filepath.Dir(cfg.Log.File)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Log.File,
		MaxSize:    cfg.Log.MaxSize, // megabytes
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge, // days
		Compress:   cfg.Log.Compress,
	}

	// Write to both file and stderr
	multiWriter := io.MultiWriter(fileWriter, os.Stderr)
	logger.SetOutput(multiWriter)

	// Redirect standard log to logrus
	log.SetOutput(logger.Writer())
	log.SetFlags(0)

	return nil
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	version := flag.Bool("version", false, "Print version information")
	flag.Parse()

	if *version {
		fmt.Printf("FairCom Prometheus Exporter %s (built %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	if err := setupLogging(cfg); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	logger.Infof("FairCom Prometheus Exporter %s (built %s)", Version, BuildTime)
	logger.Infof("Loading configuration from %s", *configPath)
	logger.Infof("Log output: %s (max_size: %dMB, max_backups: %d, max_age: %dd)",
		cfg.Log.File, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge)

	// Create single FairCom collector
	logger.Info("Initializing FairCom metrics collector...")
	faircomCollector := collector.NewFairComCollector(cfg.FairCom)

	// Register collector with Prometheus
	prometheus.MustRegister(faircomCollector)

	// Setup HTTP server
	http.Handle(cfg.Server.Path, promhttp.Handler())

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Starting Prometheus exporter on %s%s", addr, cfg.Server.Path)

	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
