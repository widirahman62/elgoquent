package register

type AppConf struct {
	Name string
}

type DatabaseConf struct {
	Defaults    string
	Connections map[string]map[string]string
}

type Config struct {
	Database DatabaseConf
	App      AppConf
}
