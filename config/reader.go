package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"fmt"
)

func ReadFile(filePath string) (Config, error) {
	config := Config{}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		return config,
			fmt.Errorf("Error unmarshalling config file '%s': %s",
				filePath, err)
	}
	return config, nil
}
