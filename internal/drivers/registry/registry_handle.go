package drivers

import (
	"fmt"
	"strings"

	"github.com/algrvvv/monlog/internal/types"
)

type DriverFactory func() types.LineHandleDriver

var drivers = make(map[string]types.LineHandleDriver)

func RegisterDriver(name string, factory DriverFactory) {
	if _, ok := drivers[name]; ok {
		panic("Driver already registered: " + name)
	}
	drivers[name] = factory()
}

func Handle(driverName, line string) string {
	if driverName == "" {
		return line
	}

	// driverName = strings.TrimPrefix(driverName, "use drivers;")
	driverName = strings.TrimSpace(driverName)

	if driver, ok := drivers[driverName]; !ok {
		fmt.Printf("WARNING! driver with name %s not found\n", driverName)
		return line
	} else {
		return driver.Handle(line)
	}
}
