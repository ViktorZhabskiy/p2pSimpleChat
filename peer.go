package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func CheckPeer(validetPeer chan *Peer, peers map[enode.ID]*Peer) {
	for {
		p := <-validetPeer
		if _, ok := peers[p.Peer.ID()]; !ok {
			fmt.Printf("New peer %s(%s) connect \n", p.Peer.Name(), p.Peer.RemoteAddr().String())
			peers[p.Peer.ID()] = &Peer{
				Peer: p.Peer,
				RW:   p.RW,
				OutC: p.OutC,
			}
		} else {
			delete(peers, p.Peer.ID())
		}
	}

}

type Peer struct {
	Peer *p2p.Peer
	RW   p2p.MsgReadWriter
	Name string
	OutC chan string
}

func (p *Peer) Send(mess string) {
	p.OutC <- mess
}
