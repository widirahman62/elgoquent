package kernel

import (
	"github.com/widirahman62/pkg-go-elgo/register"
)

var (
	Config = register.Config{
		App:      register.AppConf{},
		Database: register.DatabaseConf{},
	}
)
