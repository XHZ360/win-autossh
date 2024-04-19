package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type AddressPair struct {
	remote *net.TCPAddr
	local  *net.TCPAddr
	mode   ForwardMode
}

func strToAddressPair(s string, mode ForwardMode) (AddressPair, error) {
	var p AddressPair
	p.mode = mode

	addresses := strings.Split(s, ",")
	if len(addresses) != 2 {
		return p, fmt.Errorf("invalid address num in pair, expect 2, got: %d", len(addresses))
	}
	lAddr, err := net.ResolveTCPAddr("tcp", addresses[0])
	if err != nil {
		return p, fmt.Errorf("invalid local address: %s", addresses[0])
	}
	p.local = lAddr
	rAddr, err := net.ResolveTCPAddr("tcp", addresses[1])
	if err != nil {
		return p, fmt.Errorf("invalid remote address: %s", addresses[1])
	}
	p.remote = rAddr
	return p, nil
}
func toAddressPair(listStr []string, mode ForwardMode) []AddressPair {
	var res []AddressPair
	for _, s := range listStr {
		pair, err := strToAddressPair(s, mode)
		if err != nil {
			log.Println(err)
		} else {
			res = append(res, pair)
		}

	}
	return res
}

type ServerConfig struct {
	addr     string
	user     string
	password *string
	keyfile  *string
}
type Config struct {
	test     string
	mappings []AddressPair
	server   ServerConfig
}

var configPaths = [...]string{"./config.private.yaml",
	"./config.yaml"}

func readConfig(workDir string) (*Config, error) {
	_configPath := ""
	for _, path := range configPaths {
		if filepath.IsAbs(path) {
			_configPath = path
		} else {
			_configPath = filepath.Join(workDir, path)
		}
		// check if file exists
		if _, err := os.Stat(_configPath); err == nil {
			break
		}
	}
	if _configPath == "" {
		log.Fatal("cannot find config file, tried:", configPaths)
	}
	f, err := os.Open(_configPath)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	viper.SetConfigType("yaml")
	viperErr := viper.ReadConfig(f)
	if viperErr != nil {
		return nil, err
	}
	c := Config{}
	c.test = viper.GetString("test")
	rtl := toAddressPair(viper.GetStringSlice("mappings.rtl"), ForwardModeRtl)
	ltr := toAddressPair(viper.GetStringSlice("mappings.ltr"), ForwardModeLtr)
	c.mappings = append(rtl, ltr...)
	c.server.addr = viper.GetString("server.addr")
	c.server.user = viper.GetString("server.user")
	if viper.InConfig("server.password") {
		str := viper.GetString("server.password")
		c.server.password = &str
	} else {
		c.server.password = nil
	}
	if viper.InConfig("server.keyfile") {
		str := viper.GetString("server.keyfile")
		c.server.keyfile = &str
	} else {
		c.server.keyfile = nil
	}
	return &c, nil
}
