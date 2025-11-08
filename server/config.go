// Copyright 2025 Joshua Rich <joshua.rich@gmail.com>.
// SPDX-License-Identifier: MIT

package server

import "time"

// Config contains server-related config options.
type Config struct {
	// Host is the hostname to listen on.
	Host string `toml:"host" validate:"hostname_rfc1123"`
	// Port is the port to listen on.
	Port string `toml:"port" validate:"numeric,gt=0,lt=65535"`
	// CertFile points to the file containing a server certificate.
	CertFile string `toml:"cert" validate:"omitempty,required_with=KeyFile,file"`
	// KeyFile points to the file containing a server key.
	KeyFile      string        `toml:"key" validate:"omitempty,required_with=CertFile,file"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
	IdleTimeout  time.Duration `toml:"idle_timeout"`
}

// NewConfig creates a new default server config with sane values.
func NewConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         "8223",
		CertFile:     "",
		KeyFile:      "",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

type configOption func(*Config)

// WithHost sets the hostname to listen on.
func WithHost(host string) configOption {
	return func(c *Config) {
		if host != "" {
			c.Host = host
		}
	}
}

// WithPort sets the port to listen on.
func WithPort(port string) configOption {
	return func(c *Config) {
		if port != "" {
			c.Port = port
		}
	}
}

// WithCertFile sets the path to a certificate to use for serving over HTTPS.
func WithCertFile(file string) configOption {
	return func(c *Config) {
		if file != "" {
			c.CertFile = file
		}
	}
}

// WithKeyFile sets the path to a key to use for serving over HTTPS.
func WithKeyFile(file string) configOption {
	return func(c *Config) {
		if file != "" {
			c.KeyFile = file
		}
	}
}
