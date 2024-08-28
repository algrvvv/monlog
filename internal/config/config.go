package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	TGBotToken        string `yaml:"tg_bot_token"`
	PathToIDRSA       string `yaml:"path_to_id_rsa" validate:"required"`
	Port              int    `yaml:"port" validate:"required"`
	MaxLocalLogSizeMB int    `yaml:"max_local_log_size_mb" validate:"required"`
	NumberRowsToLoad  int    `yaml:"number_rows_to_load" validate:"required"`
}

type ServerConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Name      string   `yaml:"name" validate:"required"`
	Host      string   `yaml:"host" validate:"-"`
	User      string   `yaml:"user" validate:"-"`
	Port      int      `yaml:"port" validate:"-"`
	LogDir    string   `yaml:"log_dir" validate:"required"`
	StartLine string   `yaml:"start_line" validate:"required"`
	ChatIDs   []string `yaml:"chat_ids"`
	Notify    bool     `yaml:"notify"`
	IsLocal   bool     `yaml:"is_local" validate:"checkHUP"`
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

func checkHUPValidator(fl validator.FieldLevel) bool {
	config := fl.Parent().Interface().(ServerConfig)
	if !config.IsLocal {
		if config.Host == "" || config.Port == 0 || config.User == "" {
			return false
		}
	}
	return true
}

func translateError(err validator.ValidationErrors) map[string]string {
	customErrMsg := make(map[string]string)
	getErrorMsg := func(fieldError validator.FieldError) string {
		if fieldError.Tag() == "required" {
			return fmt.Sprintf("%s is required field [%s(%s)]", fieldError.StructField(), fieldError.Namespace(), fieldError.Kind())
		} else if fieldError.Tag() == "checkHUP" {
			return fmt.Sprintf("If server is not local, his host, user and port required [%s]", fieldError.Namespace())
		}
		return fieldError.Error()
	}

	for _, fieldError := range err {
		customErrMsg[fieldError.StructField()] = getErrorMsg(fieldError)
	}
	return customErrMsg
}

func validateConfigPart(validate *validator.Validate, s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		customErrMsg := translateError(validationErrors)
		for field, message := range customErrMsg {
			log.Printf("Config validation error: %s: %s\n", field, message)
		}
		return errors.New("failed to parse config.yml")
	}
	return nil
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

	validate := validator.New()
	_ = validate.RegisterValidation("checkHUP", checkHUPValidator)

	if err = validateConfigPart(validate, config.App); err != nil {
		log.Fatal(err)
	}
	for i, server := range config.Servers {
		if err = validateConfigPart(validate, server); err != nil {
			log.Fatalf("[server-%d] %v", i, err)
		}
	}

	Cfg = config
	return nil
}
