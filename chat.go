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
	srv            *p2p.Server
	peers          map[enode.ID]*Peer
	bootstrapNodes []string
	ctx            context.Context
}

func InitChat(ctx context.Context, conf FileConf, privKey *ecdsa.PrivateKey) *Chat {
	fmt.Println("Init Chat")

	peers := make(map[enode.ID]*Peer)
	var validatePeer = make(chan *Peer)

	go CheckPeer(validatePeer, peers)

	return &Chat{
		srv: &p2p.Server{
			Config: p2p.Config{
				Name:       conf.Name,
				PrivateKey: privKey,
				Protocols:  []p2p.Protocol{createProtocol(validatePeer)},
				ListenAddr: conf.ListenAddr,
				MaxPeers:   conf.MaxPeers,
			},
		},
		peers:          peers,
		bootstrapNodes: conf.BootstrapNodes,
		ctx:            ctx,
	}
}

func (c *Chat) Start() {
	fmt.Println("Start p2p server")

	err := c.srv.Start()
	if err != nil {
		log.Fatal("Start p2p.Server failed", err)
	}

	fmt.Println("Node info:", c.srv.Self())

	c.connectNodes()
	fmt.Println("-Stated!-")
	fmt.Printf("Your name:[%s] \n", c.srv.Config.Name)

	c.wait()
}

func (c *Chat) SendMessage(mess string) {
	for _, peer := range c.peers {
		peer.Send(mess)
	}
}

func (c *Chat) wait() {
	go func() {
		<-c.ctx.Done()
		c.srv.Stop()
		fmt.Println("P2P server stop!")
		c.peers = nil
		os.Exit(0)
	}()

}

func (c *Chat) connectNodes() {
	for _, p := range c.bootstrapNodes {
		peer, err := enode.ParseV4(p)
		if err != nil {
			fmt.Println(fmt.Errorf("invalid enode: %v", err))
		}

		c.srv.AddPeer(peer)
	}
}
