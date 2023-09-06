package foundation

import (
	"github.com/widirahman62/pkg-go-elgo/kernel"
	"github.com/widirahman62/pkg-go-elgo/register"
)

type Base struct {
	Config register.Config
}

func RegisterBase(base *Base) {
	kernel.Config.App = base.Config.App
	kernel.Config.Database = base.Config.Database
}
