package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

var (
	// App's
	ErrAppInvalidName    = errors.New("invalid name of service")
	ErrAppInvalidVersion = errors.New("invalid version of service")

	// Server's
	ErrServerInvalidServerCert = errors.New("invalid path to server cert")
	ErrServerInvalidServerKey  = errors.New("invalid path to server key")
	ErrServerInvalidCaCert     = errors.New("invalid path to ca cert")
	ErrServerTooLittleTimeout  = errors.New("too little timeout")

	// Postgres's
	ErrPostgresInvalidUser       = errors.New("invalid user")
	ErrPostgresInvalidPassword   = errors.New("invalid password")
	ErrPostgresInvalidDBName     = errors.New("invalid dbname")
	ErrPostgresTooLittleLifetime = errors.New("too little lifetime of conn")
	ErrPostgresTooLittleIdleTime = errors.New("too little idle time of conn")

	// Cache's
	ErrCacheTooLittleChatsTTL = errors.New("too little chat's ttl")

	// Jwt's
	ErrJWTInvalidPublicKeyPath = errors.New("invalid path to public key")

	// General's
	ErrInvalidPath = errors.New("invalid path to config file")
)

// naming of current service
type app struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (a *app) validate() error {
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return ErrAppInvalidName
	}

	a.Version = strings.TrimSpace(a.Version)
	if a.Version == "" {
		return ErrAppInvalidVersion
	}

	return nil
}

// settings of server's runner
type server struct {
	GRPC struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Protocol string `yaml:"protocol"`
		TLS      struct {
			Enable     bool   `yaml:"enable"`
			ServerCert string `yaml:"server-cert"`
			ServerKey  string `yaml:"server-key"`
			CaCert     string `yaml:"ca-cert"`
		} `yaml:"tls"`
	} `yaml:"grpc"`
	Timeout time.Duration `yaml:"timeout"`
}

func (s *server) applyDefaults() {
	s.GRPC.Host = strings.TrimSpace(s.GRPC.Host)
	if s.GRPC.Host == "" {
		s.GRPC.Host = "localhost"
	}

	if s.GRPC.Port <= 0 || s.GRPC.Port > 65535 {
		s.GRPC.Port = 50051
	}

	s.GRPC.Protocol = strings.TrimSpace(s.GRPC.Protocol)
	if s.GRPC.Protocol == "" {
		s.GRPC.Protocol = "tcp"
	}
}

func (s *server) validate() error {
	if s.GRPC.TLS.Enable {
		s.GRPC.TLS.ServerCert = strings.TrimSpace(s.GRPC.TLS.ServerCert)
		if s.GRPC.TLS.ServerCert == "" {
			return ErrServerInvalidServerCert
		}

		s.GRPC.TLS.ServerKey = strings.TrimSpace(s.GRPC.TLS.ServerKey)
		if s.GRPC.TLS.ServerKey == "" {
			return ErrServerInvalidServerKey
		}

		s.GRPC.TLS.CaCert = strings.TrimSpace(s.GRPC.TLS.CaCert)
		if s.GRPC.TLS.CaCert == "" {
			return ErrServerInvalidCaCert
		}
	}

	if s.Timeout < 100*time.Millisecond {
		return ErrServerTooLittleTimeout
	}

	return nil
}

type postgres struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	SSLMode string `yaml:"sslmode"`
	Auth    struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"auth"`
	Conn struct {
		MaxIdles    int           `yaml:"max-idles"`
		MaxOpens    int           `yaml:"max-opens"`
		MaxLifetime time.Duration `yaml:"max-lifetime"`
		MaxIdleTime time.Duration `yaml:"max-idle-time"`
	} `yaml:"conn"`
}

func (pg *postgres) applyDefaults() {
	pg.Host = strings.TrimSpace(pg.Host)
	if pg.Host == "" {
		pg.Host = "localhost"
	}

	if pg.Port <= 0 || pg.Port > 65535 {
		pg.Port = 5432
	}

	pg.SSLMode = strings.TrimSpace(pg.SSLMode)
	if pg.SSLMode == "" {
		pg.SSLMode = "disable"
	}

	if pg.Conn.MaxIdles <= 0 {
		pg.Conn.MaxIdles = 1
	}
	if pg.Conn.MaxOpens <= 0 {
		pg.Conn.MaxOpens = 1
	}
}

func (pg *postgres) validate() error {
	pg.Auth.User = strings.TrimSpace(pg.Auth.User)
	if pg.Auth.User == "" {
		return ErrPostgresInvalidUser
	}

	pg.Auth.Password = strings.TrimSpace(pg.Auth.Password)
	if pg.Auth.Password == "" {
		return ErrPostgresInvalidPassword
	}

	pg.Auth.DBName = strings.TrimSpace(pg.Auth.DBName)
	if pg.Auth.DBName == "" {
		return ErrPostgresInvalidDBName
	}

	if pg.Conn.MaxLifetime < time.Minute {
		return ErrPostgresTooLittleLifetime
	}

	if pg.Conn.MaxIdleTime < time.Minute {
		return ErrPostgresTooLittleIdleTime
	}

	return nil
}

type redis struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Auth struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"auth"`
}

func (r *redis) applyDefaults() {
	r.Host = strings.TrimSpace(r.Host)
	if r.Host == "" {
		r.Host = "localhost"
	}

	if r.Port <= 0 || r.Port > 65535 {
		r.Port = 6379
	}
}

type cache struct {
	ChatsTTL time.Duration `yaml:"chats-ttl"`
}

func (c *cache) validate() error {
	if c.ChatsTTL < 100*time.Millisecond {
		return ErrCacheTooLittleChatsTTL
	}

	return nil
}

type jwt struct {
	PublicKeyPath string `yaml:"public-key-path"`
}

func (j *jwt) validate() error {
	j.PublicKeyPath = strings.TrimSpace(j.PublicKeyPath)
	if j.PublicKeyPath == "" {
		return ErrJWTInvalidPublicKeyPath
	}

	return nil
}

type Config struct {
	Env     string `yaml:"env"`
	App     app    `yaml:"app"`
	Server  server `yaml:"server"`
	Service struct {
		Cache cache `yaml:"cache"`
	} `yaml:"service"`
	Secrets struct {
		Postgres postgres `yaml:"postgres"`
		Redis    redis    `yaml:"redis"`
		JWT      jwt      `yaml:"jwt"`
	} `yaml:"secrets"`
}

// parse config file
// .env must be loaded
func New(filepath string) (*Config, error) {
	filepath = strings.TrimSpace(filepath)
	if filepath == "" {
		return nil, ErrInvalidPath
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	// Apply all fields by default
	cfg.Server.applyDefaults()
	cfg.Secrets.Postgres.applyDefaults()
	cfg.Secrets.Redis.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Secrets.Postgres.validate(); err != nil {
		return fmt.Errorf("invalid postgres: %w", err)
	}
	if err := c.Service.Cache.validate(); err != nil {
		return fmt.Errorf("invalid cache: %w", err)
	}

	return nil
}
