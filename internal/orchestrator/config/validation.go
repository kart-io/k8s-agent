package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.validateServer(); err != nil {
		return fmt.Errorf("server config invalid: %w", err)
	}

	if err := c.validateDatabase(); err != nil {
		return fmt.Errorf("database config invalid: %w", err)
	}

	if err := c.validateRedis(); err != nil {
		return fmt.Errorf("redis config invalid: %w", err)
	}

	if err := c.validateNATS(); err != nil {
		return fmt.Errorf("nats config invalid: %w", err)
	}

	if err := c.validateAI(); err != nil {
		return fmt.Errorf("AI config invalid: %w", err)
	}

	return nil
}

func (c *Config) validateServer() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	return nil
}

func (c *Config) validateDatabase() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}

	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}

	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("max idle connections (%d) cannot exceed max open connections (%d)",
			c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}

	if c.Database.ConnMaxLifetime < 0 {
		return fmt.Errorf("connection max lifetime cannot be negative")
	}

	return nil
}

func (c *Config) validateRedis() error {
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis address is required")
	}

	// Validate addr format (host:port)
	parts := strings.Split(c.Redis.Addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("redis addr must be in format host:port")
	}

	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		return fmt.Errorf("invalid redis database: %d (must be 0-15)", c.Redis.DB)
	}

	if c.Redis.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be positive")
	}

	if c.Redis.MinIdleConns < 0 {
		return fmt.Errorf("redis min idle connections cannot be negative")
	}

	if c.Redis.DialTimeout <= 0 {
		return fmt.Errorf("redis dial timeout must be positive")
	}

	return nil
}

func (c *Config) validateNATS() error {
	if c.NATS.URL == "" {
		return fmt.Errorf("nats URL is required")
	}

	// Basic URL validation
	if _, err := url.Parse(c.NATS.URL); err != nil {
		return fmt.Errorf("invalid nats URL: %w", err)
	}

	if c.NATS.MaxReconnect < -1 {
		return fmt.Errorf("invalid max reconnect: %d (use -1 for unlimited)", c.NATS.MaxReconnect)
	}

	if c.NATS.ReconnectWait <= 0 {
		return fmt.Errorf("reconnect wait must be positive")
	}

	return nil
}

func (c *Config) validateAI() error {
	if c.AI.ReasoningServiceURL == "" {
		return fmt.Errorf("reasoning service URL is required")
	}

	if c.AI.AgentManagerURL == "" {
		return fmt.Errorf("agent manager URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(c.AI.ReasoningServiceURL); err != nil {
		return fmt.Errorf("invalid reasoning service URL: %w", err)
	}

	if _, err := url.Parse(c.AI.AgentManagerURL); err != nil {
		return fmt.Errorf("invalid agent manager URL: %w", err)
	}

	if c.AI.Timeout <= 0 {
		return fmt.Errorf("AI timeout must be positive")
	}

	if c.AI.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	return nil
}
