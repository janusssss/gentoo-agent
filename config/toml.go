package config

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
)

var GetConfig = sync.OnceValue(func() *Config {
	confPath, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	conf := &Config{
		path: filepath.Join(confPath, "honi", "config.toml"),
	}
	conf.readTomlConfig()
	return conf
})

type Config struct {
	OpenAI LLM `toml:"openAI"`
	path   string
}

type LLM struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	BaseUrl  string `toml:"base_url"`
	ApiKey   string `toml:"api_key"`
}

func (c *Config) readTomlConfig() {
	if _, err := toml.DecodeFile(c.path, c); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return
		}
		panic("读取配置文件错误:" + err.Error())
	}
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
