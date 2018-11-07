package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Chat struct {
	srv     *p2p.Server
	storage *Storage
	ctx     context.Context
}

func InitChat(ctx context.Context, conf FileConf, privKey *ecdsa.PrivateKey) *Chat {
	fmt.Println("Init Chat")

	storage := &Storage{
		Peers:    map[enode.ID]*Peer{},
		PeerC:    make(chan *Peer),
		DeletePC: make(chan *Peer),
	}

	storage.CheckPeers()
	storage.DisconnectPeers()

	bootNodes := readBootNodes(conf.BootstrapNodes)

	return &Chat{
		srv: &p2p.Server{
			Config: p2p.Config{
				Name:           conf.Name,
				PrivateKey:     privKey,
				Protocols:      []p2p.Protocol{createProtocol(storage)},
				ListenAddr:     conf.ListenAddr,
				MaxPeers:       conf.MaxPeers,
				BootstrapNodes: bootNodes,
			},
		},
		storage: storage,
		ctx:     ctx,
	}
}

func (c *Chat) Start() {
	fmt.Println("Start p2p server")

	err := c.srv.Start()
	if err != nil {
		log.Fatal("Start p2p.Server failed", err)
	}

	fmt.Println("Node url:", c.srv.Self())
	fmt.Println("-Stated!-")
	fmt.Printf("Your name:[%s] \n", c.srv.Config.Name)

	c.wait()
}

func (c *Chat) SendMessage(mess string) {
	for _, peer := range c.storage.Peers {
		peer.Send(mess)
	}
}

func (c *Chat) wait() {
	go func() {
		<-c.ctx.Done()
		c.srv.Stop()
		c.storage = nil
		fmt.Println("P2P server stop!")
		os.Exit(0)
	}()

}

func readBootNodes(nodes []string) []*enode.Node {
	bootNodes := []*enode.Node{}
	for _, p := range nodes {
		peer, err := enode.ParseV4(p)
		if err != nil {
			fmt.Println(fmt.Errorf("invalid enode: %v", err))
			continue
		}
		bootNodes = append(bootNodes, peer)
	}
	return bootNodes
}
