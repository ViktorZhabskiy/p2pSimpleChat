package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Chat struct {
	srv            *p2p.Server
	peers          map[enode.ID]*Peer
	bootstrapNodes []string
}

func InitChat(conf FileConf, privKey *ecdsa.PrivateKey) *Chat {
	fmt.Println("Init Chat")

	peers := make(map[enode.ID]*Peer)
	var validatePeer = make(chan Peer)

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
	}
}

func (c *Chat) Start() {
	fmt.Println("Start p2p server")

	err := c.srv.Start()
	if err != nil {
		log.Fatal("Start p2p.Server failed", err)
	}

	fmt.Println("Node info: ", c.srv.Self())
	// c.subscribe()
	c.connectNodes()
	fmt.Println("-Stated!-")

}

func (c *Chat) SendMessage(mess string) {
	for _, peer := range c.peers {
		peer.Send(mess)
	}
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

func (c *Chat) subscribe() {
	eventOneC := make(chan *p2p.PeerEvent)
	sub_one := c.srv.SubscribeEvents(eventOneC)
	go func(sub_one event.Subscription) {
		for {
			peerevent := <-eventOneC
			switch peerevent.Type {
			case p2p.PeerEventTypeAdd:
				fmt.Printf("New peer %s connect \n", peerevent.Peer)
			case p2p.PeerEventTypeDrop:
				fmt.Printf("peer %s disconnect \n", peerevent.Peer)
				delete(c.peers, peerevent.Peer)
			case p2p.PeerEventTypeMsgRecv:
				fmt.Println(peerevent)
			default:
				fmt.Println("Unknown peer event type")
			}
		}
	}(sub_one)
}
