package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

const minSessionExpiration = 15 * time.Second

type SM3AuthConfig struct {
	UserName string `json:"username"`
	Data     string `json:"data"`
}

type UserServiceConfig struct {
	SessionExpirationStr string        `json:"session_expiration"`
	SessionExpiration    time.Duration `json:"-"`
	KeepAliveDurationStr string        `json:"keepalive_duration"`
	KeepAliveDuration    time.Duration `json:"-"`
	PdAddress            string        `json:"pd_addr"`
}

type Config struct {
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	UserService UserServiceConfig `json:"user_service"`
}

func (c *UserServiceConfig) Validate() error {
	sessionExpiration, err := time.ParseDuration(c.SessionExpirationStr)
	if err != nil {
		return err
	}
	if sessionExpiration < minSessionExpiration {
		sessionExpiration = minSessionExpiration
	}
	c.SessionExpiration = sessionExpiration

	keepAliveDuration, err := time.ParseDuration(c.KeepAliveDurationStr)
	if err != nil {
		return err
	}
	if keepAliveDuration <= 0 || keepAliveDuration >= c.SessionExpiration {
		keepAliveDuration = c.SessionExpiration / 2
	}
	c.KeepAliveDuration = keepAliveDuration
	return nil
}

func (c *Config) Validate() error {
	if net.ParseIP(c.Host) == nil {
		return errors.New("invalid host")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port")
	}

	return c.UserService.Validate()
}

func ParseConfigFile(file string) (*Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := Config{}
	dec := json.NewDecoder(f)
	err = dec.Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Config) Print() {
	b, err := json.MarshalIndent(c, "", "\t")
	if err == nil {
		fmt.Printf("config:\n%s\n", string(b))
	}
}
