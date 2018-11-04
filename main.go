package main

import (
	"bufio"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/crypto"
)

const ConfigFilePath = "./chat.conf"

func main() {
	conf := parseConf(ConfigFilePath)

	privKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Generate private key failed", err)
	}

	chat := InitChat(conf, privKey)
	chat.Start()

	go handleInput(chat)

	select {}
}

type FileConf struct {
	Name           string
	ListenAddr     string
	BootstrapNodes []string `toml:"bootstrap_nodes"`
	MaxPeers       int      `toml:"max_peers"`
}

func parseConf(filePath string) FileConf {
	conf := FileConf{}
	if _, err := toml.DecodeFile(filePath, &conf); err != nil {
		log.Fatal("Fail read or decode conf file: ", err)
	}
	return conf
}

func handleInput(c *Chat) {
	go func() {
		r := bufio.NewReader(os.Stdin)
		message, err := r.ReadString('\n')
		if err != nil {
			return
		}
		c.SendMessage(message)

		os.Stdin.Close()
	}()

	return
}
