package config

import (
	"io/ioutil"
	"log"

	"github.com/gzuidhof/tygo/tygo"
	"gopkg.in/yaml.v2"
)

func ReadFromFilepath(cfgFilepath string) tygo.Config {
	b, err := ioutil.ReadFile(cfgFilepath)
	if err != nil {
		log.Fatalf("Could not read config file from %s: %v", cfgFilepath, err)
	}
	conf := tygo.Config{}
	err = yaml.Unmarshal(b, &conf)
	if err != nil {
		log.Fatalf("Could not parse config file from: %v", err)
	}

	return conf
}
