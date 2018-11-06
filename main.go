package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	ctx, cancel := context.WithCancel(context.Background())
	handleStopProgramm(ctx, cancel)

	chat := InitChat(ctx, conf, privKey)
	chat.Start()

	go handleInput(ctx, chat)

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

func handleStopProgramm(ctx context.Context, cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigC
		log.Printf("%s signal received. shuting down!", sig)
		cancel()
		os.Exit(0)
	}()
}

func handleInput(ctx context.Context, c *Chat) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				os.Stdin.Close()
				return
			default:
				r := bufio.NewReader(os.Stdin)
				message, err := r.ReadString('\n')
				if err != nil {
					fmt.Println("Fail read from stdin: ", err)
					return
				}
				c.SendMessage(message)
			}

		}
	}()

	return
}
