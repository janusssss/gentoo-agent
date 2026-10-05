package config

import (
	"sync"

	"github.com/BurntSushi/toml"
)

var GetConfig = sync.OnceValue(func() *Config {
	conf := &Config{}
	ReadTomlConfig(conf)
	return conf
})

type Config struct {
	OpenAI LLM `toml:"openAI"`
}

type LLM struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	BaseUrl  string `toml:"base_url"`
}

func ReadTomlConfig(conf *Config) {
	if _, err := toml.DecodeFile("config.toml", conf); err != nil {
		panic(err)
	}
}
