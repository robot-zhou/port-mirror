package main

import (
	"encoding/json"
	"os"
)

const (
	AppName           = "port-mirror"
	DefautlConfigFile = "/etc/port-mirror.json"

	DefaultAliveTimeout = 60
	DefaultReadTimeout  = 5
	DefaultWriteTimeout = 30
)

type MirrorConfig struct {
	Local        string `json:"local,omitempty"`
	Target       string `json:"target,omitempty"`
	Proxy        string `json:"proxy,omitempty"`
	PacUrl       string `json:"pac_url,omitempty"`
	AliveTimeout int    `json:"alive_timeout,omitempty"`
	ReadTimeout  int    `json:"read_timeout,omitempty"`
	WriteTimeout int    `json:"write_timeout,omitempty"`
}

func (sc *MirrorConfig) ResetDefautl() {
	if sc.AliveTimeout <= 0 {
		sc.AliveTimeout = DefaultAliveTimeout
	}
	if sc.ReadTimeout <= 0 {
		sc.ReadTimeout = DefaultReadTimeout
	}
	if sc.WriteTimeout <= 0 {
		sc.WriteTimeout = DefaultWriteTimeout
	}
}

type Config struct {
	LogLevel string         `json:"log_level:omitempty"`
	LogFile  string         `json:"log_file:omitempty"`
	Mirrors  []MirrorConfig `json:"mirrors,omitempty"`
}

var (
	config Config = Config{
		LogLevel: "info",
		LogFile:  "stdout",
	}
)

func GetConfig() *Config {
	return &config
}

func GetConfigJson() string {
	str, _ := json.MarshalIndent(&config, "", "    ")
	return string(str)
}

func LoadConfig(a ...string) error {
	var (
		err      error
		content  []byte
		filepath string = append(a, DefautlConfigFile)[0]
	)

	if content, err = os.ReadFile(filepath); err == nil {
		err = json.Unmarshal(content, &config)
	}

	for idx := range config.Mirrors {
		config.Mirrors[idx].ResetDefautl()
	}

	return err
}
