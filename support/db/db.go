package db

import (
	"log"

	"github.com/widirahman62/pkg-go-elgo/database/connectors"
	"github.com/widirahman62/pkg-go-elgo/kernel"
)

type config struct {
	Connection map[string]string
}

type D []struct {
	Key   string
	Value interface{}
}
type E struct {
	Key   string
	Value interface{}
}
type M map[string]interface{}
type A []interface{}

func (d D) Get() *[]struct {
	Key   string
	Value interface{}
} {
	return (*[]struct{Key string; Value interface{}})(&d)
}

func (e E) Get() *struct {
	Key   string
	Value interface{}
} {
	return (*struct{Key string; Value interface{}})(&e)
}

func (m M) Get() *map[string]interface{}{
	return (*map[string]interface{})(&m)
}

func (a A) Get() *[]interface{} {
	return (*[]interface{})(&a)
}

func checkDefaults() string {
	if kernel.Config.Database.Defaults == "" {
		log.Fatal("Database defaults is not set")
	}
	return kernel.Config.Database.Defaults
}

func UseConnection(name string) config {
	if name == "" {
		name = checkDefaults()
	}
	return config{Connection: kernel.Config.Database.Connections[name]}
}

func (c *config) MongoConnection() *connectors.MongoDB {
	if c.Connection["driver"] != "mongodb" {
		log.Fatalf("MongoConnection() function not support for '%s' driver", c.Connection["driver"])
	}
	return connectors.NewMongoDB(&c.Connection)
}

func MongoConnection() *connectors.MongoDB {
	db := UseConnection(kernel.Config.Database.Defaults)
	return db.MongoConnection()
}
