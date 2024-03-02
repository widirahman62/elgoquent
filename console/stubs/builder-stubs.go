package stubs

type BuilderStubs struct {}

func (*BuilderStubs) Register(param ...string) string {
	return `
package `+param[0]+`

import (
    "`+param[1]+`/foundation"
    config "`+param[2]+`"
)

var base foundation.Base

func init() {
	base.Config.Database.Defaults = config.Database.Defaults
	base.Config.Database.Connections = config.Database.Connections
}

func Register() {
	foundation.RegisterBase(&base)
}
`
}

func (*BuilderStubs) ConfigDatabase(param ...string) string {
	return `
package `+param[0]+`

import (
	"`+param[1]+`/support/env"
)

type database struct {
	Defaults    string
	Connections map[string]map[string]string
}

var Database = &database{
	Defaults: env.IfEmptySet("mysql").Get("DB_CONNECTION"),
	Connections: map[string]map[string]string{
		"mongodb": {
			"driver":   "mongodb",
			"host":     env.IfEmptySet("127.0.0.1").Get("MONGO_DB_HOST"),
			"port":     env.IfEmptySet("27017").Get("MONGO_DB_PORT"),
			"database": env.Get("MONGO_DB_DATABASE"),
			"username": env.Get("MONGO_DB_USERNAME"),
			"password": env.Get("MONGO_DB_PASSWORD"),
		},
		"mysql": {
			"driver":   "mysql",
			"host":     env.IfEmptySet("127.0.0.1").Get("MYSQL_DB_DB_HOST"),
			"port":     env.IfEmptySet("3306").Get("MYSQL_DB_PORT"),
			"database": env.Get("MYSQL_DB_DATABASE"),
			"username": env.IfEmptySet("root").Get("MYSQL_DB_USERNAME"),
			"password": env.Get("MYSQL_DB_PASSWORD"),
		},
	},
}
`
}
