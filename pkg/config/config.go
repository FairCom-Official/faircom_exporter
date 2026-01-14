package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the exporter configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	FairCom    FairComConfig    `yaml:"faircom"`
	Collectors CollectorsConfig `yaml:"collectors"`
	Log        LogConfig        `yaml:"log"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port      int              `yaml:"port"`
	Path      string           `yaml:"path"`
	TLS       *TLSConfig       `yaml:"tls,omitempty"`
	BasicAuth *BasicAuthConfig `yaml:"basic_auth,omitempty"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// BasicAuthConfig holds basic authentication configuration
type BasicAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// FairComConfig holds FairCom connection settings
type FairComConfig struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	CtstatPath string `yaml:"ctstat_path"`
	Timeout    int    `yaml:"timeout"` // in seconds
}

// CollectorsConfig holds collector enable/disable flags
type CollectorsConfig struct {
	Cache        bool `yaml:"cache"`
	Transactions bool `yaml:"transactions"`
	Locks        bool `yaml:"locks"`
	Files        bool `yaml:"files"`
	ISAM         bool `yaml:"isam"`
	SQL          bool `yaml:"sql"`
	Users        bool `yaml:"users"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `yaml:"level"`       // debug, info, warn, error
	Format     string `yaml:"format"`      // json or text
	File       string `yaml:"file"`        // log file path
	MaxSize    int    `yaml:"max_size"`    // max size in MB before rotation
	MaxBackups int    `yaml:"max_backups"` // max number of old log files
	MaxAge     int    `yaml:"max_age"`     // max days to retain old log files
	Compress   bool   `yaml:"compress"`    // compress rotated files
}

// GetTimeout returns the timeout as a time.Duration
func (c *FairComConfig) GetTimeout() time.Duration {
	if c.Timeout <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.Timeout) * time.Second
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 9100
	}
	if cfg.Server.Path == "" {
		cfg.Server.Path = "/metrics"
	}
	if cfg.FairCom.CtstatPath == "" {
		cfg.FairCom.CtstatPath = "/opt/faircom/ctstat"
	}
	if cfg.FairCom.Timeout == 0 {
		cfg.FairCom.Timeout = 10
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
	if cfg.Log.File == "" {
		cfg.Log.File = "/var/log/faircom_exporter/faircom_exporter.log"
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 100
	}
	if cfg.Log.MaxBackups == 0 {
		cfg.Log.MaxBackups = 3
	}
	if cfg.Log.MaxAge == 0 {
		cfg.Log.MaxAge = 28
	}

	// Default all collectors to enabled if not specified
	if !cfg.Collectors.Cache && !cfg.Collectors.Transactions && !cfg.Collectors.Locks &&
		!cfg.Collectors.Files && !cfg.Collectors.ISAM && !cfg.Collectors.SQL && !cfg.Collectors.Users {
		cfg.Collectors.Cache = true
		cfg.Collectors.Transactions = true
		cfg.Collectors.Locks = true
		cfg.Collectors.Files = true
		cfg.Collectors.ISAM = true
		cfg.Collectors.SQL = true
		cfg.Collectors.Users = true
	}

	return &cfg, nil
}
