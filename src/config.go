package main

import (
	"fmt"
	"gopkg.in/ini.v1"
)

// Server :: [Server]-Section of .ini
type Server struct {
	ListenAddress string `ini:"listen_address"`
	ListenPort    int    `ini:"listen_port"`
	Domain        string `ini:"domain"`
}

type Config struct {
	Server
}

var (
	parsed = false
	config = &Config{
		Server: Server{
			ListenAddress: "127.0.0.1",
			ListenPort:    17397,
			Domain:        "localhost",
		},
	}
)

const (
	iniPath = "config.ini"
)

func parse() {
	if parsed {
		return
	}

	file, err := ini.Load(iniPath)

	if err != nil {
		panic(fmt.Errorf("fatal error while reading configuration from 'config.ini': %w", err))
	}

	err = file.MapTo(config)

	if err != nil {
		panic(fmt.Errorf("fatal error while pasing configuration to types: %w", err))
	}

	parsed = true
}

func GetListenAddress() string {
	parse()
	return config.ListenAddress
}

func GetListenPort() int {
	parse()
	return config.ListenPort
}

func GetDomain() string {
	parse()
	return config.Domain
}
