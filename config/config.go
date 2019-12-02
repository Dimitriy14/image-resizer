package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"syscall"
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

	PostgresHost     string `json:"PostgresHost"     default:"localhost"`
	PostgresPort     string `json:"PostgresPort"     default:"5431"`
	PostgresDBName   string `json:"PostgresDBName"   default:"resizer"`
	PostgresUser     string `json:"PostgresUser"     default:"app"`
	PostgresPassword string `json:"PostgresPassword" default:"1337"`

	AWSID     string `json:"-"     envconfig:"AWS_ACCESS_KEY_ID"`
	AWSSecret string `json:"-"     envconfig:"AWS_SECRET_ACCESS_KEY"`

	AWSRegion               string `json:"AWSRegion"               default:"eu-central-1"`
	AWSBucket               string `json:"AWSBucket"               default:"resized-images-yal"`
	AWSACL                  string `json:"AWSACL"                  default:"public-read"`
	AWSServerSideEncryption string `json:"AWSServerSideEncryption" default:"AES256"`
	AWSImageStorageURL      string `json:"AWSImageStorageURL"      default:"https://resized-images-yal.s3.eu-central-1.amazonaws.com"`

	LogFile  string `json:"LogFile"`
	LogLevel string `json:"LogLevel"                 default:"debug"`
}

func Load() error {
	if err := readFile(&Conf); err != nil {
		return err
	}
	log.Printf("Configuration: %+v", Conf)
	if err := mergeEnvconfig(&Conf); err != nil {
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

	return json.NewDecoder(fileContent).Decode(&Conf)
}

func mergeEnvconfig(config *Configuration) (err error) {
	configElements := reflect.TypeOf(config).Elem()
	for i := 0; i < configElements.NumField(); i++ {
		envKey, hasEnvconfigTag := configElements.Field(i).Tag.Lookup("envconfig")
		if !hasEnvconfigTag {
			continue
		}
		envValue, found := syscall.Getenv(envKey)
		if !found {

			continue
		}
		fmt.Println(envKey, envValue)
		structFieldName := configElements.Field(i).Name
		envField := reflect.ValueOf(config).Elem().FieldByName(structFieldName)
		switch envField.Kind() {
		case reflect.String:
			envField.SetString(envValue)
		case reflect.Int, reflect.Int64:
			intEnvValue, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return fmt.Errorf("can not parse field %s value %s as Int64 type", envField.Type().Name(), envValue)
			}
			envField.SetInt(intEnvValue)
		}
	}
	return
}
