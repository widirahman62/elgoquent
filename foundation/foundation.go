package foundation

import (
	"log"

	"github.com/widirahman62/elgoquent/kernel"
	"github.com/widirahman62/elgoquent/register"
	"github.com/widirahman62/elgoquent/support/env"
)

type Base struct {
	Config register.Config
}

func RegisterBase(base *Base) {
	if env.Err != nil {
		log.Fatal(env.Err)
	}
	kernel.Config.App = base.Config.App
	kernel.Config.Database = base.Config.Database
}
