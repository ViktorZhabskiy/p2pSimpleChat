package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

type Storage struct {
	Peers    map[enode.ID]*Peer
	PeerC    chan *Peer
	DeletePC chan *Peer
	sync.Mutex
}

func (s *Storage) GetPeer(key enode.ID) (*Peer, error) {
	s.Lock()
	defer s.Unlock()

	if peer, ok := s.Peers[key]; ok {
		return peer, nil
	}

	return nil, errors.New("Storage don`t have this peer")
}

func (s *Storage) SetPeer(key enode.ID, peer *Peer) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Peers[key]; ok {
		return errors.New(fmt.Sprintf("Peer %s already exist!", key))
	}

	s.Peers[key] = peer
	return nil
}

func (s *Storage) DelPeer(key enode.ID) error {
	s.Lock()
	defer s.Unlock()

	delete(s.Peers, key)

	return nil
}

func (s *Storage) CheckPeers() {
	go func() {
		for {
			p := <-s.PeerC
			if _, ok := s.GetPeer(p.Peer.ID()); ok != nil {
				fmt.Printf("New peer %s connect \n", p.Name)
				s.SetPeer(p.Peer.ID(), p)
			}
		}
	}()
}

func (s *Storage) DisconnectPeers() {
	go func() {
		for {
			p := <-s.DeletePC
			s.DelPeer(p.Peer.ID())
			fmt.Printf("%s: peer disconnected \n", p.Name)
		}
	}()
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
