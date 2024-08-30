package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"

	"github.com/algrvvv/monlog/internal/logger"
)

type AppConfig struct {
	TGBotToken        string `yaml:"tg_bot_token"`
	PathToIDRSA       string `yaml:"path_to_id_rsa" validate:"required,file"`
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

// checkHUPValidator function to check Host, User, Port if the IsLocal field is false.
// hence the name - CheckHostUserPort
func checkHUPValidator(fl validator.FieldLevel) bool {
	config := fl.Parent().Interface().(ServerConfig)
	if !config.IsLocal {
		if config.Host == "" || config.Port == 0 || config.User == "" {
			return false
		}
	}
	return true
}

// translateError function for custom error messages
func translateError(err validator.ValidationErrors) map[string]string {
	customErrMsg := make(map[string]string)
	getErrorMsg := func(fieldError validator.FieldError) string {
		if fieldError.Tag() == "required" {
			return fmt.Sprintf("%s is required field [%s(%s)]", fieldError.StructField(), fieldError.Namespace(), fieldError.Kind())
		} else if fieldError.Tag() == "checkHUP" {
			return fmt.Sprintf("If server is not local, his host, user and port required [%s]", fieldError.Namespace())
		} else if fieldError.Tag() == "file" {
			return fmt.Sprintf("%s must be an existing file [%s]", fieldError.StructField(), fieldError.Namespace())
		}
		return fieldError.Error()
	}

	for _, fieldError := range err {
		customErrMsg[fieldError.StructField()] = getErrorMsg(fieldError)
	}
	return customErrMsg
}

// validateConfigPart function for custom error messages
func validateConfigPart(validate *validator.Validate, s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var validationErrors validator.ValidationErrors
		errors.As(err, &validationErrors)
		customErrMsg := translateError(validationErrors)
		for _, message := range customErrMsg {
			logger.Warn("Config validation error: " + message)
		}
		return errors.New("failed to parse config.yml")
	}
	return nil
}

var Cfg Config

// LoadConfig the main function for load config
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
		logger.Fatal(err.Error(), err)
	}
	for i, server := range config.Servers {
		if err = validateConfigPart(validate, server); err != nil {
			logger.Warn(
				fmt.Sprintf("Server number %d ==> %v; This server will be forcibly disabled from the general list of servers", i, err),
				slog.Any("error", err))
			config.Servers[i].Enabled = false
			logger.Info(fmt.Sprintf("Server number %d is disabled ==> %v", i, server.Enabled))
		}
	}

	Cfg = config
	return nil
}
