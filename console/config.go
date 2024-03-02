package console

const (
    builderInit = iota
    builderChangeFilePath
    builderChangeDirPath
    builderMoveWorkdir
)

const (
    makeModel = iota
)    

type (
	teaErr error
)

var (
	asciiArt = `
    ________                                   __ 
   / ____/ /____ ____  ____ ___  _____  ____  / /_
  / __/ / / __  / __ \/ __  / / / / _ \/ __ \/ __/
 / /___/ / /_/ / /_/ / /_/ / /_/ /  __/ / / / /_  
/_____/_/\__, /\____/\__, /\__,_/\___/_/ /_/\__/  
        /____/         /_/                        

`
	appname 	  = "elgoquent"
	packageName	  = "github.com/widirahman62/elgoquent"
	version       = "0.1.1"
    defaultEnv = map[string]string{
    "DB_CONNECTION" : "mongodb",
    "MONGO_DB_HOST" : "127.0.0.1",
    "MONGO_DB_PORT": "27017",
    "MONGO_DB_DATABASE": "mongodb",
    "MONGO_DB_USERNAME": "",
    "MONGO_DB_PASSWORD": "",
}
)
