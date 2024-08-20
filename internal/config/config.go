package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	TGBotToken  string `yaml:"tg_bot_token"`
	PathToIDRSA string `yaml:"path_to_id_rsa"`
	Port        int    `yaml:"port"`
}

type ServerConfig struct {
	Host      string   `yaml:"host"`
	User      string   `yaml:"user"`
	Port      int      `yaml:"port"`
	LogDir    string   `yaml:"log_dir"`
	StartLine string   `yaml:"start_line"`
	ChatIDs   []string `yaml:"chat_ids"`
}

type Keywords struct {
	Time string `yaml:"time"`
	Info string `yaml:"info"`
	Lvl  string `yaml:"lvl"`
	Msg  string `yaml:"msg"`
	Err  string `yaml:"err"`
}

type Config struct {
	App      AppConfig      `yaml:"app"`
	Servers  []ServerConfig `yaml:"servers"`
	Keywords Keywords       `yaml:"keywords"`
}

var Cfg Config

func LoadConfig(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return errors.New("failed to parse config.yaml" + err.Error())
	}

	Cfg = config
	return nil
}
