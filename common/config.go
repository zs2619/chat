package common

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type WebConfigType struct {
	LogLevel      string `json:"logLevel"`
	NatsURI       string `json:"natsURI"`
	RTMServerPort int    `json:"rtmServerPort"`
	RTMServerType string `json:"rtmServerType"`
}

func getConfigRaw(config string) ([]byte, error) {
	configPath := os.Getenv("KISSRTM_CONFIGPATH")
	filePath := filepath.Join(configPath, config)
	raw, err := ioutil.ReadFile(filePath)
	return raw, err
}

var WebConfig WebConfigType

func LoadWebConfig(config string) (*WebConfigType, error) {
	rawFile, err := getConfigRaw(config)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rawFile, &WebConfig)
	if err != nil {
		return nil, err
	}
	return &WebConfig, nil
}
