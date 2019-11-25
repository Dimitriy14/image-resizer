package config

import (
	"encoding/json"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

var (
	// FilePath is a basic config file (could be changed by flag -config=yourConfigFile.json)
	FilePath = "config.json"
	//Conf is a configuration instance
	Conf Configuration
)

type Configuration struct {
	ListenURL string `json:"ListenURL"   default:":8181" envconfig:"ListenURL"`
	BasePath  string `json:"BasePath"    default:"/resizer"`

	Postgres struct {
		Host     string `json:"Host"`
		Port     string `json:"Port"`
		DBName   string `json:"DBName"   default:"resizer"`
		User     string `json:"User"     default:"admin"`
		Password string `json:"Password" default:"Pass@1377"`
	} `json:"Postgres"`

	AWS struct {
		Region               string `json:"Region"               default:"eu-central-1"`
		Bucket               string `json:"Bucket"               default:"resized-images-yal"`
		ACL                  string `json:"ACL"                  default:"public-read"`
		ServerSideEncryption string `json:"ServerSideEncryption" default:"AES256"`
	}

	UseLogFile bool   `json:"UseLogFile" default:"false"`
	LogFile    string `json:"LogFile"    default:"resizer.log"`
	LogLevel   string `json:"LogLevel"   default:"debug"`
}

func Load() error {
	if err := readFile(&Conf); err != nil {
		return err
	}

	if err := readEnv(&Conf); err != nil {
		return err
	}

	log.Printf("Configuration: %+v", Conf)
	return nil
}

func readFile(cfg *Configuration) error {
	fileContent, err := os.Open(FilePath)
	if err != nil {
		return err
	}

	if err = json.NewDecoder(fileContent).Decode(&Conf); err != nil {
		return err
	}
	return nil
}

func readEnv(cfg *Configuration) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return err
	}
	return nil
}
