package config

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConfig implements Config interface for testing
type mockConfig struct {
	validateCalled bool
	completeCalled bool
	validateError  error
	completeError  error
}

func (m *mockConfig) Validate() error {
	m.validateCalled = true
	return m.validateError
}

func (m *mockConfig) Complete() error {
	m.completeCalled = true
	return m.completeError
}

// mockFlaggableConfig implements FlaggableConfig interface for testing
type mockFlaggableConfig struct {
	mockConfig
	addFlagsCalled bool
	receivedPrefix string
}

func (m *mockFlaggableConfig) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	m.addFlagsCalled = true
	if len(prefixes) > 0 {
		m.receivedPrefix = prefixes[0]
	}
}

func TestConfigInterface(t *testing.T) {
	t.Run("Config interface implementation", func(t *testing.T) {
		var _ Config = (*mockConfig)(nil)

		cfg := &mockConfig{}

		err := cfg.Validate()
		assert.NoError(t, err)
		assert.True(t, cfg.validateCalled)

		err = cfg.Complete()
		assert.NoError(t, err)
		assert.True(t, cfg.completeCalled)
	})

	t.Run("Config with errors", func(t *testing.T) {
		validateErr := assert.AnError
		completeErr := assert.AnError

		cfg := &mockConfig{
			validateError: validateErr,
			completeError: completeErr,
		}

		err := cfg.Validate()
		assert.Equal(t, validateErr, err)

		err = cfg.Complete()
		assert.Equal(t, completeErr, err)
	})
}

func TestFlaggableConfigInterface(t *testing.T) {
	t.Run("FlaggableConfig interface implementation", func(t *testing.T) {
		var _ FlaggableConfig = (*mockFlaggableConfig)(nil)

		cfg := &mockFlaggableConfig{}
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

		cfg.AddFlags(fs)
		assert.True(t, cfg.addFlagsCalled)
		assert.Equal(t, "", cfg.receivedPrefix)
	})

	t.Run("FlaggableConfig with prefix", func(t *testing.T) {
		cfg := &mockFlaggableConfig{}
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)

		cfg.AddFlags(fs, "app")
		assert.True(t, cfg.addFlagsCalled)
		assert.Equal(t, "app", cfg.receivedPrefix)
	})

	t.Run("FlaggableConfig includes Config methods", func(t *testing.T) {
		cfg := &mockFlaggableConfig{}

		err := cfg.Validate()
		assert.NoError(t, err)
		assert.True(t, cfg.validateCalled)

		err = cfg.Complete()
		assert.NoError(t, err)
		assert.True(t, cfg.completeCalled)
	})
}

func TestOptionType(t *testing.T) {
	type testConfig struct {
		Name  string
		Value int
	}

	t.Run("Option type definition", func(t *testing.T) {
		// Verify Option type can be defined and assigned
		var opt Option[testConfig]
		opt = func(c *testConfig) {
			c.Name = "test"
		}
		assert.NotNil(t, opt)

		// Verify it can be called
		cfg := &testConfig{}
		opt(cfg)
		assert.Equal(t, "test", cfg.Name)
	})

	t.Run("Option function creation", func(t *testing.T) {
		withName := func(name string) Option[testConfig] {
			return func(c *testConfig) {
				c.Name = name
			}
		}

		withValue := func(value int) Option[testConfig] {
			return func(c *testConfig) {
				c.Value = value
			}
		}

		cfg := &testConfig{}
		withName("test")(cfg)
		assert.Equal(t, "test", cfg.Name)

		withValue(42)(cfg)
		assert.Equal(t, 42, cfg.Value)
	})
}

func TestApply(t *testing.T) {
	type testConfig struct {
		Name   string
		Value  int
		Active bool
	}

	withName := func(name string) Option[testConfig] {
		return func(c *testConfig) {
			c.Name = name
		}
	}

	withValue := func(value int) Option[testConfig] {
		return func(c *testConfig) {
			c.Value = value
		}
	}

	withActive := func(active bool) Option[testConfig] {
		return func(c *testConfig) {
			c.Active = active
		}
	}

	t.Run("Apply single option", func(t *testing.T) {
		cfg := &testConfig{}
		Apply(cfg, withName("test"))

		assert.Equal(t, "test", cfg.Name)
		assert.Equal(t, 0, cfg.Value)
		assert.False(t, cfg.Active)
	})

	t.Run("Apply multiple options", func(t *testing.T) {
		cfg := &testConfig{}
		Apply(cfg,
			withName("test"),
			withValue(42),
			withActive(true),
		)

		assert.Equal(t, "test", cfg.Name)
		assert.Equal(t, 42, cfg.Value)
		assert.True(t, cfg.Active)
	})

	t.Run("Apply no options", func(t *testing.T) {
		cfg := &testConfig{
			Name:   "default",
			Value:  100,
			Active: false,
		}

		Apply(cfg)

		// Should remain unchanged
		assert.Equal(t, "default", cfg.Name)
		assert.Equal(t, 100, cfg.Value)
		assert.False(t, cfg.Active)
	})

	t.Run("Apply options can override", func(t *testing.T) {
		cfg := &testConfig{Name: "old"}

		Apply(cfg,
			withName("new1"),
			withName("new2"), // Last one wins
		)

		assert.Equal(t, "new2", cfg.Name)
	})

	t.Run("Apply with nil options slice", func(t *testing.T) {
		cfg := &testConfig{Name: "test"}
		var opts []Option[testConfig]

		require.NotPanics(t, func() {
			Apply(cfg, opts...)
		})

		assert.Equal(t, "test", cfg.Name)
	})
}

func TestApplyWithComplexTypes(t *testing.T) {
	type nestedConfig struct {
		Host string
		Port int
	}

	type complexConfig struct {
		Name   string
		Nested *nestedConfig
		Tags   []string
	}

	withNested := func(host string, port int) Option[complexConfig] {
		return func(c *complexConfig) {
			c.Nested = &nestedConfig{Host: host, Port: port}
		}
	}

	withTags := func(tags ...string) Option[complexConfig] {
		return func(c *complexConfig) {
			c.Tags = tags
		}
	}

	t.Run("Apply with nested structs", func(t *testing.T) {
		cfg := &complexConfig{}
		Apply(cfg,
			withNested("localhost", 8080),
			withTags("tag1", "tag2"),
		)

		require.NotNil(t, cfg.Nested)
		assert.Equal(t, "localhost", cfg.Nested.Host)
		assert.Equal(t, 8080, cfg.Nested.Port)
		assert.Equal(t, []string{"tag1", "tag2"}, cfg.Tags)
	})
}
