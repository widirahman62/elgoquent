package env

import "github.com/joho/godotenv"

var envFile, Err = godotenv.Read()

type env struct {
	defaults string
}

func IfEmptySet(s string) *env {
	return &env{defaults: s}
}

func (e *env) Get(s string) string {
	if envFile[s] == "" {
        return e.defaults
    }
    return envFile[s]
}

func Get(s string) string {
	return envFile[s]
}

