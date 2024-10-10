package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Debug             bool   `yaml:"debug"`
	TGBotToken        string `yaml:"tg_bot_token"`
	PathToIDRSA       string `yaml:"path_to_id_rsa" validate:"required,file"`
	Port              int    `yaml:"port" validate:"required"`
	MaxLocalLogSizeMB int    `yaml:"max_local_log_size_mb" validate:"required"`
	NumberRowsToLoad  int    `yaml:"number_rows_to_load" validate:"required"`
}

type ServerConfig struct {
	ID            int      `yaml:"id" validate:"required"`
	Enabled       bool     `yaml:"enabled"`
	Name          string   `yaml:"name" validate:"required"`
	Host          string   `yaml:"host" validate:"-"`
	User          string   `yaml:"user" validate:"-"`
	Port          int      `yaml:"port" validate:"-"`
	LogDir        string   `yaml:"log_dir" validate:"required"`
	LogLayout     string   `yaml:"log_layout" validate:"required"`
	LogLevel      string   `yaml:"log_levels" validate:"required"`
	LogTimeFormat string   `yaml:"log_time_format" validate:"required"`
	StartLine     string   `yaml:"start_line" validate:"required"`
	ChatIDs       []string `yaml:"chat_ids" validate:"required"`
	Notify        bool     `yaml:"notify" validate:"required"`
	IsLocal       bool     `yaml:"is_local" validate:"checkHUP"`
}

type DefaultServerConfig struct {
	StartLine     string   `yaml:"start_line"`
	LogDir        string   `yaml:"log_dir"`
	LogLayout     string   `yaml:"log_layout"`
	LogLevel      string   `yaml:"log_levels"`
	LogTimeFormat string   `yaml:"log_time_format"`
	ChatIDs       []string `yaml:"chat_ids"`
	Notify        bool     `yaml:"notify"`
}

type Config struct {
	App      AppConfig           `yaml:"app"`
	Defaults DefaultServerConfig `yaml:"default_servers_setting"`
	Servers  []ServerConfig      `yaml:"servers"`
}

var DefaultSettings = make(map[string]interface{})

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

func setDefaultSettings(s interface{}) {
	val := reflect.ValueOf(s).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := val.Type().Field(i).Name

		if field.IsZero() {
			if defaultValue, exists := DefaultSettings[fieldName]; exists {
				switch defaultValue.(type) {
				case []interface{}:
					var result []string
					for _, value := range defaultValue.([]interface{}) {
						if v, ok := value.(string); ok {
							result = append(result, v)
						}
					}
					field.Set(reflect.ValueOf(result))
				default:
					field.Set(reflect.ValueOf(defaultValue))
				}
			}
		}
	}
}

// translateError function for custom error messages
func translateError(err validator.ValidationErrors) map[string]string {
	customErrMsg := make(map[string]string)
	getErrorMsg := func(fieldError validator.FieldError) string {
		switch fieldError.Tag() {
		case "required":
			return fmt.Sprintf("%s is required field [%s(%s)]", fieldError.StructField(), fieldError.Namespace(), fieldError.Kind())
		case "checkHUP":
			return fmt.Sprintf("If server is not local, his host, user and port required [%s]", fieldError.Namespace())
		case "file":
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
			log.Printf("[WARN] config validation error: %s", message)
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

	inrec, _ := json.Marshal(config.Defaults)
	_ = json.Unmarshal(inrec, &DefaultSettings)

	for i := range config.Servers {
		setDefaultSettings(&config.Servers[i])
	}

	validate := validator.New()
	_ = validate.RegisterValidation("checkHUP", checkHUPValidator)

	if err = validateConfigPart(validate, config.App); err != nil {
		log.Fatal(err)
	}
	for i, server := range config.Servers {
		if err = validateConfigPart(validate, server); err != nil {
			log.Printf(
				"[WARN] Server number %d ==> %v; This server will be forcibly disabled from the general list of servers", i, err,
			)
			config.Servers[i].Enabled = false
			log.Printf("[INFO] Server number %d is disabled ==> %v", i, config.Servers[i].Enabled)
		}
	}

	Cfg = config
	log.Println("[INFO] config successfully loaded")
	return nil
}
