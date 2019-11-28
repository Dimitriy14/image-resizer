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
		Host     string `json:"Host"     default:"localhost"`
		Port     string `json:"Port"     default:"5431"`
		DBName   string `json:"DBName"   default:"resizer"`
		User     string `json:"User"     default:"app"`
		Password string `json:"Password" default:"1337"`
	} `json:"Postgres"`

	AWS struct {
		ID     string `json:""    envconfig:"AWS_ACCESS_KEY_ID"`
		Secret string `json:"-"   envconfig:"AWS_SECRET_ACCESS_KEY"`

		Region               string `json:"Region"               default:"eu-central-1"`
		Bucket               string `json:"Bucket"               default:"resized-images-yal"`
		ACL                  string `json:"ACL"                  default:"public-read"`
		ServerSideEncryption string `json:"ServerSideEncryption" default:"AES256"`
		ImageStorageURL      string `json:"ImageStorageURL"      default:"https://resized-images-yal.s3.eu-central-1.amazonaws.com"`
	} `json:"AWS"`

	LogFile  string `json:"LogFile"`
	LogLevel string `json:"LogLevel"   default:"debug"`
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
	err := envconfig.Process("envconfig", cfg)
	if err != nil {
		return err
	}
	return nil
}
