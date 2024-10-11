package config

import (
	"fmt"

	"github.com/sasbury/mini"
)

var Config *mini.Config

func init() {
	Config = GetConfig()
	shortToken := GetValueString("general", "token", "WRONG_CONF")[0:10]

	fmt.Println("---------------------------------")
	fmt.Println(" > TOKEN: ", fmt.Sprintf("%s...", shortToken))
	fmt.Println(" > PREFIX: ", GetValueString("general", "prefix", "."))
	fmt.Println(" > SOUNDS_DIR: ", GetValueString("general", "sounds_dir", "."))
	fmt.Println("---------------------------------")
}

func GetConfig() *mini.Config {
	fmt.Println("Loading config file", "file", "../config/config.ini")
	conf, err := mini.LoadConfiguration("../config/config.ini")
	if err != nil {
		fmt.Println("error reading config file config.ini")
		panic(err)
	}
	return conf
}

func GetValueString(section, key, def string) string {
	value := Config.StringFromSection(section, key, def)
	return value
}

func GetValueInt(section, key string, def int64) int64 {
	value := Config.IntegerFromSection(section, key, def)
	return value
}

func GetValueBool(section, key string, def bool) bool {
	value := Config.BooleanFromSection(section, key, def)

	return value
}
