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
	ErrServerInvalidCert      = errors.New("invalid path to server cert")
	ErrServerInvalidKey       = errors.New("invalid path to server key")
	ErrServerInvalidCaCert    = errors.New("invalid ca cert")
	ErrServerTooLittleRequest = errors.New("too little timeout of request")

	// Cipher's
	ErrCipherTooLittleNonce = errors.New("too little size of nonce")

	// Postgres
	ErrPostgresInvalidUser       = errors.New("invalid user")
	ErrPostgresInvalidPassword   = errors.New("invalid password")
	ErrPostgresInvalidDBName     = errors.New("invalid name of db")
	ErrPostgresTooLittleLifetime = errors.New("too little lifetime")
	ErrPostgresTooLittleIdleTime = errors.New("too little idle time")

	// JWT's
	ErrJWTInvalidPublicKeyPath = errors.New("invalid path to public key")

	// Services
	ErrServicesInvalidHost       = errors.New("invalid host of service")
	ErrServicesInvalidPort       = errors.New("invalid port of service")
	ErrServicesInvalidClientCert = errors.New("invalid path to client cert")
	ErrServicesInvalidClientKey  = errors.New("invalid path to client key")
	ErrServicesInvalidCaCert     = errors.New("invalid path to ca cert")

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

type server struct {
	GRPC struct {
		Host     string `yaml:"host"`     // by def: localhost
		Port     int    `yaml:"port"`     // by def: 50052
		Protocol string `yaml:"protocol"` // by def: tcp
		TLS      struct {
			Enable     bool   `yaml:"enable"`
			ServerCert string `yaml:"server-cert"`
			ServerKey  string `yaml:"server-key"`
			CaCert     string `yaml:"ca-cert"`
		} `yaml:"tls"`
	} `yaml:"grpc"`
	Timeout time.Duration `yaml:"timeout"`
}

func (s *server) applyDef() {
	s.GRPC.Host = strings.TrimSpace(s.GRPC.Host)
	if s.GRPC.Host == "" {
		s.GRPC.Host = "localhost"
	}

	if s.GRPC.Port <= 0 || s.GRPC.Port > 65535 {
		s.GRPC.Port = 50052
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
	}

	if s.Timeout < 100*time.Millisecond {
		return ErrServerTooLittleRequest
	}

	return nil
}

type cipher struct {
	NonceSize int `yaml:"nonce-size"`
}

func (c *cipher) validate() error {
	if c.NonceSize <= 0 {
		return ErrCipherTooLittleNonce
	}

	return nil
}

type postgres struct {
	Host    string `yaml:"host"`    // by def: localhost
	Port    int    `yaml:"port"`    // by def: 5432
	SSLMode string `yaml:"sslmode"` // by def: disable
	Auth    struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"auth"`
	Conn struct {
		MaxIdles    int           `yaml:"max-idles"` // by def: 1
		MaxOpens    int           `yaml:"max-opens"` // by def: 1
		MaxLifetime time.Duration `yaml:"max-lifetime"`
		MaxIdleTime time.Duration `yaml:"max-idle-time"`
	} `yaml:"conn"`
}

func (pg *postgres) applyDef() {
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

	pg.Conn.MaxIdles = max(pg.Conn.MaxIdles, 1)
	pg.Conn.MaxOpens = max(pg.Conn.MaxOpens, 1)
}

// The validation of password, db and user
// will be while creating connection with postgres
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

	if pg.Conn.MaxLifetime < time.Second {
		return ErrPostgresTooLittleLifetime
	}

	if pg.Conn.MaxIdleTime < time.Second {
		return ErrPostgresTooLittleIdleTime
	}

	return nil
}

type redis struct {
	Host string `yaml:"host"` // by def: localhost
	Port int    `yaml:"port"` // by def: 6379
	Auth struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"auth"`
}

func (r *redis) applyDef() {
	r.Host = strings.TrimSpace(r.Host)
	if r.Host == "" {
		r.Host = "localhost"
	}

	if r.Port <= 0 || r.Port > 65535 {
		r.Port = 6379
	}
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

type xvibeAuth struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  struct {
		Enable     bool   `yaml:"enable"`
		ClientCert string `yaml:"client-cert"`
		ClientKey  string `yaml:"client-key"`
		CaCert     string `yaml:"ca-cert"`
	} `yaml:"tls"`
}

func (xa *xvibeAuth) validate() error {
	xa.Host = strings.TrimSpace(xa.Host)
	if xa.Host == "" {
		return ErrServicesInvalidHost
	}

	if xa.Port <= 0 || xa.Port > 65535 {
		return ErrServicesInvalidPort
	}

	if xa.TLS.Enable {
		xa.TLS.ClientCert = strings.TrimSpace(xa.TLS.ClientCert)
		if xa.TLS.ClientCert == "" {
			return ErrServicesInvalidClientCert
		}

		xa.TLS.ClientKey = strings.TrimSpace(xa.TLS.ClientKey)
		if xa.TLS.ClientKey == "" {
			return ErrServicesInvalidClientKey
		}

		xa.TLS.CaCert = strings.TrimSpace(xa.TLS.CaCert)
		if xa.TLS.CaCert == "" {
			return ErrServicesInvalidCaCert
		}
	}

	return nil
}

type Config struct {
	Env      string `yaml:"env"`
	App      app    `yaml:"app"`
	Server   server `yaml:"server"`
	Services struct {
		XvibeAuth xvibeAuth `yaml:"xvibe-auth"`
	} `yaml:"services"`
	Secrets struct {
		Cipher   cipher   `yaml:"cipher"`
		Postgres postgres `yaml:"postgres"`
		Redis    redis    `yaml:"redis"`
		JWT      jwt      `yaml:"jwt"`
	} `yaml:"secrets"`
}

// load config
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
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	cfg.Server.applyDef()
	cfg.Secrets.Postgres.applyDef()
	cfg.Secrets.Redis.applyDef()

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
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Secrets.Cipher.validate(); err != nil {
		return fmt.Errorf("invalid cipher: %w", err)
	}
	if err := c.Secrets.Postgres.validate(); err != nil {
		return fmt.Errorf("invalid postgres: %w", err)
	}
	if err := c.Secrets.JWT.validate(); err != nil {
		return fmt.Errorf("invalid jwt: %w", err)
	}
	if err := c.Services.XvibeAuth.validate(); err != nil {
		return fmt.Errorf("invalid xvibe-auth: %w", err)
	}

	return nil
}
