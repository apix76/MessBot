package main

import (
	"MessBot/Conf"
	"MessBot/Framework"
	"encoding/json"
	"log"
	"os"
)

func main() {
	conf := GetConf()
	Framework.LoopFramework(conf)
}

func GetConf() Conf.Conf {
	file, err := os.Open("MessConfig.cfg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var conf Conf.Conf
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
