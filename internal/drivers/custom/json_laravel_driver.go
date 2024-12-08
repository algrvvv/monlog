package drivers

import (
	"encoding/json"
	"fmt"
	"time"

	drivers "github.com/algrvvv/monlog/internal/drivers/registry"
	"github.com/algrvvv/monlog/internal/logger/log"
	"github.com/algrvvv/monlog/internal/types"
)

type LaravelJSONDriver struct {
	Name string
}

func init() {
	// NOTE: для его использования в конфигурации нужного сервера укажите "log_driver": "json:laravel"
	log.PrintInfo("load laravel json driver")
	drivers.RegisterDriver("json:laravel", func() types.LineHandleDriver {
		return NewLaravelJSONDriver()
	})
	log.PrintInfo("laravel json driver loaded")
}

// создаем струтуру json, которую мы ожидаем получить и заанмаршлить.
type laravelJSON struct {
	Message string `json:"message"`
	Context struct {
		IP        string `json:"ip,omitempty"`
		Exception struct {
			Class   string `json:"class"`
			Message string `json:"message"`
			Code    int    `json:"code"`
			File    string `json:"file"`
		} `json:"exception,omitempty"`
	} `json:"context"`
	Level     int       `json:"level"`
	LevelName string    `json:"level_name"`
	Channel   string    `json:"channel"`
	Datetime  time.Time `json:"datetime"`
	Extra     any       `json:"extra"`
}

func NewLaravelJSONDriver() LaravelJSONDriver {
	return LaravelJSONDriver{Name: "json:laravel"}
}

func (l LaravelJSONDriver) GetName() string {
	return l.Name
}

func (l LaravelJSONDriver) Handle(j string) string {
	var out laravelJSON
	if err := json.Unmarshal([]byte(j), &out); err != nil {
		log.PrintErrorf("driver:%s: failed to unmarshal string: %s: %v\n", l.GetName(), j, err)
		return j
	}

	var ex string
	if out.Context.Exception.File != "" {
		ex = fmt.Sprintf(
			"%s: %s - %s",
			out.Context.Exception.Class,
			out.Context.Exception.Message,
			out.Context.Exception.File,
		)
	}

	return fmt.Sprintf(
		"[%v] %s %s; %s",
		out.Datetime.Format(time.DateTime),
		out.LevelName,
		out.Message,
		ex,
	)
}
