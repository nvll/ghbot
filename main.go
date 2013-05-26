package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type Config struct {
	Server       string
	Port         int
	Nick         []string
	Channels     []string
	Users        []string
	ClientId     string
	ClientSecret string
}

func main() {
	var err error
	var cfg Config
	var cfg_blob []byte
	var child = make(chan bool)
	var msg = make(chan string)

	cfg_blob, err = ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Couldn't open config")
	}

	err = json.Unmarshal(cfg_blob, &cfg)
	if err != nil {
		log.Printf("%v\n", string(cfg_blob))
		log.Fatal("Couldn't parse config json: ", err.Error())
	}

	fmt.Printf("%v\n", cfg)
	go startServer(child, cfg, msg)
	go irc(cfg, child, msg)

	<-child
}
