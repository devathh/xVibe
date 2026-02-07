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
	ErrServerInvalidCert      = errors.New("invalid server cert")
	ErrServerInvalidKey       = errors.New("invalid server key")
	ErrServerInvalidCaCert    = errors.New("invalid ca cert")
	ErrServerTooLittleTimeout = errors.New("too little request's timeout")

	// Postgres
	ErrPostgresInvalidUser       = errors.New("invalid user")
	ErrPostgresInvalidPassword   = errors.New("invalid password")
	ErrPostgresInvalidDBName     = errors.New("invalid dbname")
	ErrPostgresInvalidMaxOpens   = errors.New("invalid max open conns")
	ErrPostgresInvalidMaxIdles   = errors.New("invalid max idle conns")
	ErrPostgresTooLittleLifetime = errors.New("too little lifetime")
	ErrPostgresTooLittleIdletime = errors.New("too little idletime")

	// Session's
	ErrSessionRefreshTTL = errors.New("invalid refresh ttl")
	ErrSessionAccessTTL  = errors.New("invalid access ttl")

	// Cache's
	ErrCacheTooLittleUserTTL = errors.New("too little user cache's ttl")

	// Jwt's
	ErrJWTPrivateKeyPath = errors.New("invalid path to private key")
	ErrJWTPublicKeyPath  = errors.New("invalid path to public key")

	// General's
	ErrInvalidPath = errors.New("invalid path to config file")
)

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
	// Apply default host
	s.GRPC.Host = strings.TrimSpace(s.GRPC.Host)
	if s.GRPC.Host == "" {
		s.GRPC.Host = "localhost"
	}

	// Apply default port
	if s.GRPC.Port <= 0 || s.GRPC.Port > 65535 {
		s.GRPC.Port = 50050
	}

	// Apply default protocol
	s.GRPC.Protocol = strings.TrimSpace(s.GRPC.Protocol)
	if s.GRPC.Protocol == "" {
		s.GRPC.Protocol = "tcp"
	}
}

// The validity of the certificate
// is checked at enable: true,
// at the time of creating the tls config
func (s *server) validate() error {
	if s.Timeout < 100*time.Millisecond {
		return ErrServerTooLittleTimeout
	}

	if !s.GRPC.TLS.Enable {
		return nil
	}

	s.GRPC.TLS.ServerCert = strings.TrimSpace(s.GRPC.TLS.ServerCert)
	if s.GRPC.TLS.ServerCert == "" {
		return ErrServerInvalidCert
	}

	s.GRPC.TLS.ServerKey = strings.TrimSpace(s.GRPC.TLS.ServerKey)
	if s.GRPC.TLS.ServerKey == "" {
		return ErrServerInvalidKey
	}

	s.GRPC.TLS.CaCert = strings.TrimSpace(s.GRPC.TLS.CaCert)
	if s.GRPC.TLS.CaCert == "" {
		return ErrServerInvalidCaCert
	}

	return nil
}

type postgres struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	SSLMode string `yaml:"sslmode"`

	Auth struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"auth"`

	Conn struct {
		MaxOpens    int           `yaml:"max-opens"`
		MaxIdles    int           `yaml:"max-idles"`
		MaxLifetime time.Duration `yaml:"max-lifetime"`
		MaxIdletime time.Duration `yaml:"max-idletime"`
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
}

// The validity of dbname, user, password
// will be checked at the time with connecting
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

	if pg.Conn.MaxIdletime < time.Second {
		return ErrPostgresTooLittleIdletime
	}

	if pg.Conn.MaxLifetime < time.Second {
		return ErrPostgresTooLittleLifetime
	}

	if pg.Conn.MaxIdles <= 0 {
		return ErrPostgresInvalidMaxIdles
	}

	if pg.Conn.MaxOpens <= 0 {
		return ErrPostgresInvalidMaxOpens
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

type session struct {
	RefreshTTL time.Duration `yaml:"refresh-ttl"`
	AccessTTL  time.Duration `yaml:"access-ttl"`
}

func (s *session) validate() error {
	if s.RefreshTTL <= 100*time.Millisecond {
		return ErrSessionRefreshTTL
	}

	if s.AccessTTL <= 100*time.Millisecond {
		return ErrSessionAccessTTL
	}

	return nil
}

type cache struct {
	UserTTL time.Duration `yaml:"user-ttl"`
}

func (c *cache) validate() error {
	if c.UserTTL < time.Second {
		return ErrCacheTooLittleUserTTL
	}

	return nil
}

type jwt struct {
	PrivateKeyPath string `yaml:"private-key-path"`
	PublicKeyPath  string `yaml:"public-key-path"`
}

func (j *jwt) validate() error {
	j.PrivateKeyPath = strings.TrimSpace(j.PrivateKeyPath)
	if j.PrivateKeyPath == "" {
		return ErrJWTPrivateKeyPath
	}

	j.PublicKeyPath = strings.TrimSpace(j.PublicKeyPath)
	if j.PublicKeyPath == "" {
		return ErrJWTPublicKeyPath
	}

	return nil
}

// Global main struct of service config
type Config struct {
	Env     string `yaml:"env"`
	App     app    `yaml:"app"`
	Server  server `yaml:"server"`
	Service struct {
		Session session `yaml:"session"`
		Cache   cache   `yaml:"cache"`
	} `yaml:"service"`
	Secrets struct {
		Postgres postgres `yaml:"postgres"`
		Redis    redis    `yaml:"redis"`
		JWT      jwt      `yaml:"jwt"`
	} `yaml:"secrets"`
}

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
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default value
	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

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
	if err := c.Service.Session.validate(); err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}
	if err := c.Secrets.JWT.validate(); err != nil {
		return fmt.Errorf("invalid jwt: %w", err)
	}
	if err := c.Service.Cache.validate(); err != nil {
		return fmt.Errorf("invalid cache: %w", err)
	}

	return nil
}
