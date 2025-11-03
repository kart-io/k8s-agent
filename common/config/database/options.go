// Package database provides database configuration options.
package database

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// Options defines database configuration.
type Options struct {
	// Connection
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	User     string `json:"user" mapstructure:"user"`
	Password string `json:"password" mapstructure:"password"`
	Database string `json:"database" mapstructure:"database"`
	SSLMode  string `json:"sslMode" mapstructure:"ssl_mode"`

	// Connection pool
	MaxOpenConns    int           `json:"maxOpenConns" mapstructure:"max_open_conns"`
	MaxIdleConns    int           `json:"maxIdleConns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"connMaxLifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime" mapstructure:"conn_max_idle_time"`

	// Logging
	LogLevel string `json:"logLevel" mapstructure:"log_level"` // silent, error, warn, info

	// Migration
	AutoMigrate bool `json:"autoMigrate" mapstructure:"auto_migrate"`
}

// DefaultOptions returns default database options.
func DefaultOptions() *Options {
	return &Options{
		Host:            "localhost",
		Port:            3306,
		User:            "root",
		Password:        "",
		Database:        "",
		SSLMode:         "disable",
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        "error",
		AutoMigrate:     false,
	}
}

// Validate validates the database options.
func (o *Options) Validate() error {
	if o.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if o.Port <= 0 || o.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", o.Port)
	}
	if o.User == "" {
		return fmt.Errorf("database user is required")
	}
	if o.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if o.MaxOpenConns < 0 {
		return fmt.Errorf("max open connections must be >= 0")
	}
	if o.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections must be >= 0")
	}
	if o.MaxIdleConns > o.MaxOpenConns && o.MaxOpenConns > 0 {
		return fmt.Errorf("max idle connections must be <= max open connections")
	}
	return nil
}

// Complete completes the database options with defaults.
func (o *Options) Complete() error {
	if o.Host == "" {
		o.Host = "localhost"
	}
	if o.Port == 0 {
		o.Port = 3306
	}
	if o.SSLMode == "" {
		o.SSLMode = "disable"
	}
	if o.MaxOpenConns == 0 {
		o.MaxOpenConns = 100
	}
	if o.MaxIdleConns == 0 {
		o.MaxIdleConns = 10
	}
	if o.ConnMaxLifetime == 0 {
		o.ConnMaxLifetime = time.Hour
	}
	if o.ConnMaxIdleTime == 0 {
		o.ConnMaxIdleTime = 10 * time.Minute
	}
	if o.LogLevel == "" {
		o.LogLevel = "error"
	}
	return nil
}

// AddFlags adds database configuration flags.
func (o *Options) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	prefix := ""
	if len(prefixes) > 0 {
		prefix = prefixes[0] + "-"
	}

	fs.StringVar(&o.Host, prefix+"db-host", o.Host, "Database host")
	fs.IntVar(&o.Port, prefix+"db-port", o.Port, "Database port")
	fs.StringVar(&o.User, prefix+"db-user", o.User, "Database user")
	fs.StringVar(&o.Password, prefix+"db-password", o.Password, "Database password")
	fs.StringVar(&o.Database, prefix+"db-name", o.Database, "Database name")
	fs.StringVar(&o.SSLMode, prefix+"db-ssl-mode", o.SSLMode, "Database SSL mode")

	fs.IntVar(&o.MaxOpenConns, prefix+"db-max-open-conns", o.MaxOpenConns, "Maximum open connections")
	fs.IntVar(&o.MaxIdleConns, prefix+"db-max-idle-conns", o.MaxIdleConns, "Maximum idle connections")
	fs.DurationVar(&o.ConnMaxLifetime, prefix+"db-conn-max-lifetime", o.ConnMaxLifetime, "Connection maximum lifetime")
	fs.DurationVar(&o.ConnMaxIdleTime, prefix+"db-conn-max-idle-time", o.ConnMaxIdleTime, "Connection maximum idle time")

	fs.StringVar(&o.LogLevel, prefix+"db-log-level", o.LogLevel, "Database log level")
	fs.BoolVar(&o.AutoMigrate, prefix+"db-auto-migrate", o.AutoMigrate, "Enable auto migration")
}

// Functional options

// Option is a functional option for database configuration.
type Option func(*Options)

// WithHost sets the database host.
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithPort sets the database port.
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithUser sets the database user.
func WithUser(user string) Option {
	return func(o *Options) {
		o.User = user
	}
}

// WithPassword sets the database password.
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithDatabase sets the database name.
func WithDatabase(database string) Option {
	return func(o *Options) {
		o.Database = database
	}
}

// WithMaxOpenConns sets the maximum open connections.
func WithMaxOpenConns(n int) Option {
	return func(o *Options) {
		o.MaxOpenConns = n
	}
}

// WithMaxIdleConns sets the maximum idle connections.
func WithMaxIdleConns(n int) Option {
	return func(o *Options) {
		o.MaxIdleConns = n
	}
}

// WithAutoMigrate enables auto migration.
func WithAutoMigrate(enable bool) Option {
	return func(o *Options) {
		o.AutoMigrate = enable
	}
}