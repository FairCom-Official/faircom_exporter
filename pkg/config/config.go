package config

import (
	"gopkg.in/yaml.v3"
	"os"
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
	Port      int             `yaml:"port"`
	Path      string          `yaml:"path"`
	TLS       TLSConfig       `yaml:"tls,omitempty"`
	BasicAuth BasicAuthConfig `yaml:"basic_auth,omitempty"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CAFile   string `yaml:"ca_file"`
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
	Servername string          `yaml:"servername"`
	TLS        TLSConfig       `yaml:"tls"`
	BasicAuth  BasicAuthConfig `yaml:"basic_auth"`
}

// CollectorsConfig holds collector enable/disable flags
type CollectorsConfig struct {
	Cache        bool `yaml:"cache"`
	Transactions bool `yaml:"transactions"`
	CallTime     bool `yaml:"call_timing"`
	Locks        bool `yaml:"locks"`
	Files        bool `yaml:"files"`
	IO           bool `yaml:"io"`
	ISAM         bool `yaml:"isam"`
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
	if cfg.FairCom.Servername == "" {
		cfg.FairCom.Servername = "FAIRCOMS"
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

	// Default collectors to enabled if not specified
	// CallTimes defaults false
	if !cfg.Collectors.Cache && !cfg.Collectors.Transactions && !cfg.Collectors.Locks && !cfg.Collectors.IO && !cfg.Collectors.Files && !cfg.Collectors.ISAM && !cfg.Collectors.Users {
		cfg.Collectors.Cache = true
		cfg.Collectors.IO = true
		cfg.Collectors.Transactions = true
		cfg.Collectors.Locks = true
		cfg.Collectors.Files = true
		cfg.Collectors.ISAM = true
		cfg.Collectors.Users = true
	}

	return &cfg, nil
}
