package main

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func createProtocol(validetPeer chan Peer) p2p.Protocol {
	return p2p.Protocol{
		Name:    "chat",
		Version: 1,
		Length:  1,
		Run:     decoreteProto(validetPeer),
	}
}

func decoreteProto(validetPeer chan Peer) func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	return func(p *p2p.Peer, rw p2p.MsgReadWriter) error {

		validetPeer <- Peer{p, rw}

		inmsg, err := rw.ReadMsg()
		if err != nil {
			if err == io.EOF {
				validetPeer <- Peer{p, rw}
				fmt.Printf("%s(%s): peer disconnected \n", p.Name(), p.RemoteAddr().String())
			} else {
				fmt.Printf("%s(%s): message read failed: %s \n", p.Name(), p.RemoteAddr().String(), err)
			}
			return err
		}

		var mess string
		err = inmsg.Decode(&mess)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%s", mess)

		return nil
	}
}

func CheckPeer(validetPeer chan Peer, peers map[enode.ID]*Peer) {
	for {
		p := <-validetPeer
		if _, ok := peers[p.Peer.ID()]; !ok {
			fmt.Printf("New peer %s(%s) connect \n", p.Peer.Name(), p.Peer.RemoteAddr().String())
			peers[p.Peer.ID()] = &Peer{
				Peer: p.Peer,
				RW:   p.RW,
			}
		} else {
			delete(peers, p.Peer.ID())
		}
	}

}

type Peer struct {
	Peer *p2p.Peer
	RW   p2p.MsgReadWriter
}
