package config

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

//go:embed default/*
var defaultFs embed.FS

var GetConfig = sync.OnceValue(func() *Config {
	conf, err := newConfig()
	if err != nil {
		panic("获取配置错误:" + err.Error())
	}
	return conf
})

type Config struct {
	LLM          LLM    `toml:"LLM"`
	SystemPrompt string `toml:"-"`
	path         string
	Debug        bool `toml:"debug"`
}

type LLM struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	BaseUrl  string `toml:"base_url"`
	ApiKey   string `toml:"api_key"`
}

func newConfig() (*Config, error) {
	confPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	confPath = filepath.Join(confPath, "honi")
	conf := &Config{
		path: filepath.Join(confPath, "config.toml"),
	}
	if err := conf.readTomlConfig(); err != nil {
		return nil, err
	}
	honiPath := filepath.Join(confPath, "honi.md")
	buff, err := os.ReadFile(honiPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		buff, err = defaultFs.ReadFile("default/honi.md")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(honiPath, buff, 0644); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Dir(honiPath), 0755); err != nil {
				return nil, err
			}
			return newConfig()
		}
	}
	conf.SystemPrompt = string(buff)

	if conf.Debug {
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(logger)
	}
	return conf, nil
}

func (c *Config) readTomlConfig() error {
	if _, err := toml.DecodeFile(c.path, c); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		buff, err := defaultFs.ReadFile("default/config.toml")
		if err != nil {
			return err
		}
		if _, err := toml.Decode(string(buff), c); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) WriteConfig() error {
	buff, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	if err = os.WriteFile(c.path, buff, 0644); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(c.path), 0755); err != nil {
			return err
		}
		return c.WriteConfig()
	}
	return nil
}

func (c *Config) Compare() bool {
	old, err := os.ReadFile(c.path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		slog.Error("real file failed", "err", err)
		return false
	}
	curr, err := toml.Marshal(c)
	if err != nil {
		slog.Error("marshal tonml failed", "err", err)
		return false
	}

	return bytes.Equal(old, curr)
}
