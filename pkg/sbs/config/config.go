package config

import (
	"log"
	"os"
	"reflect"

	"github.com/spf13/viper"
)

const (
	defaultBaseDir = "/var/local/vpc/sbs/"
)

func init() {
	viper.SetConfigFile("sbs")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("/etc/vpc")
	viper.AddConfigPath("/var/local/etc/vpc")
	viper.AddConfigPath("$HOME/.vpc")
}

//Read reads the config files
func Read() error {
	err := viper.ReadInConfig()
	if err != nil {
		_, isNotFound := err.(viper.ConfigFileNotFoundError)
		_, isPathNotFound := err.(*os.PathError)
		if isNotFound || isPathNotFound {
			return nil
		}
		log.Printf("type %s", reflect.TypeOf(err))
		return err
	}

	return nil
}

//BlockStoreDir storage location for persistent data
func BlockStoreDir() string {
	dir := viper.GetString("block-store-dir")
	if dir == "" {
		return defaultBaseDir
	}

	return dir
}
