package config

import (
	"errors"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	LogLevel       string `yaml:"log_level"`
	TelegramApiKey string `yaml:"telegram_api_key"`
	DataDir        string `yaml:"data_dir"`
	UserAgent      string `yaml:"user_agent"`
	Metrics        struct {
		Bind string `yaml:"bind"`
	} `yaml:"metrics"`
	Parsing struct {
		Interval  time.Duration `yaml:"interval"`
		UserAgent string        `yaml:"user_agent"`
	} `yaml:"parsing"`
	DiscordConfig *DiscordConfig `yaml:"DiscordConfig"`
}

type DiscordConfig struct {
	WebHook 	string `yaml:"webhook"`
	PriceFrom 	int `yaml:"price_from"`
	PriceTo 	int `yaml:"price_to"`
	RoomsFrom 	int `yaml:"rooms_from"`
	RoomsTo 	int `yaml:"rooms_to"`
}

func New(path string) (*Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(contents, &c)
	if err != nil {
		return nil, err
	}

	// Set defaults if not provided
	if len(c.DataDir) == 0 {
		c.DataDir = "/data"
	}
	if len(c.Metrics.Bind) == 0 {
		c.Metrics.Bind = "127.0.0.1:3949"
	}
	if c.Parsing.Interval == 0 {
		c.Parsing.Interval = 3 * time.Minute
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	c.LogLevel = strings.ToLower(c.LogLevel)

	// Validation
	if len(c.TelegramApiKey) == 0 && c.DiscordConfig == nil {
		return nil, errors.New("missing telegram_api_key or DiscordConfig")
	}
	if _, _, err := net.SplitHostPort(c.Metrics.Bind); err != nil {
		return nil, errors.New("missing or invalid metrics_bind")
	}
	if c.Parsing.Interval < 3*time.Second {
		return nil, errors.New("parsing interval cannot be lower than 3 seconds")
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return nil, errors.New("invalid log_level")
	}

	return &c, err
}
