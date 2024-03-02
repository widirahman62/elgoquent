package kernel

import (
	"github.com/widirahman62/elgoquent/register"
)

var (
	Config = register.Config{
		App:      register.AppConf{},
		Database: register.DatabaseConf{},
	}
)
