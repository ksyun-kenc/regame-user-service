package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

const minExpiredDuration = 15 * time.Second

const (
	AuthTypeCode = iota
	AuthTypeSM3
	AuthTypeToken
)

type AuthConfig struct {
	Version  int    `json:"version"`
	Type     int    `json:"type"`
	UserName string `json:"username"`
	Data     string `json:"data"`
}

type Config struct {
	Host                 string        `json:"host"`
	Port                 int           `json:"port"`
	ExpiredDurationStr   string        `json:"expired_duration"`
	ExpiredDuration      time.Duration `json:"-"`
	KeepAliveDurationStr string        `json:"keepalive_duration"`
	KeepAliveDuration    time.Duration `json:"-"`
	AuthCfg              []AuthConfig  `json:"auth_cfg"`
}

func (c *Config) Validate() error {
	if net.ParseIP(c.Host) == nil {
		return errors.New("host invalid")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("port invalid")
	}

	expiredDuration, err := time.ParseDuration(c.ExpiredDurationStr)
	if err != nil {
		return err
	}
	if expiredDuration < minExpiredDuration {
		expiredDuration = minExpiredDuration
	}
	c.ExpiredDuration = expiredDuration

	keepAliveDuration, err := time.ParseDuration(c.KeepAliveDurationStr)
	if err != nil {
		return err
	}
	if keepAliveDuration <= 0 || keepAliveDuration >= c.ExpiredDuration {
		keepAliveDuration = c.ExpiredDuration / 2
	}
	c.KeepAliveDuration = keepAliveDuration
	return nil
}

func ParseConfigFromJsonFile(file string) (*Config, error) {
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
