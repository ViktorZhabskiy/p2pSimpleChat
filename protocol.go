package main

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ethereum/go-ethereum/p2p"
)

const MessageCode = 0

func createProtocol(validetPeer chan *Peer) p2p.Protocol {
	return p2p.Protocol{
		Name:    "chat",
		Version: 1,
		Length:  1,
		Run:     decoreteProto(validetPeer),
	}
}

func decoreteProto(validetPeer chan *Peer) func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	return func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		outC := make(chan string)
		errorC := make(chan error)

		peer := &Peer{
			Peer: p,
			RW:   rw,
			OutC: outC,
			Name: fmt.Sprintf("%s(%s)", p.Name(), p.RemoteAddr().String()),
		}

		validetPeer <- peer

		wg := sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		wg.Add(1)
		TxMessage(ctx, cancel, peer, &wg, errorC)

		wg.Add(1)
		RxMessage(ctx, cancel, validetPeer, peer, &wg, errorC)

		wg.Wait()

		return <-errorC
	}
}

func TxMessage(ctx context.Context, cancel context.CancelFunc, peer *Peer, wg *sync.WaitGroup, errorC chan error) {
	go func() {
	breakLoop:
		for {
			select {
			case <-ctx.Done():
				return
			case message := <-peer.OutC:
				if err := p2p.Send(peer.RW, MessageCode, message); err != nil {
					fmt.Printf("Fail send message err %s", err)
					errorC <- err
					cancel()
					break breakLoop
				}
			}
		}

		wg.Done()
	}()
}

func RxMessage(ctx context.Context, cancel context.CancelFunc, validetPeer chan *Peer, peer *Peer, wg *sync.WaitGroup, errorC chan error) {
	go func() {
		for {
			inmsg, err := peer.RW.ReadMsg()

			if inmsg.Code != MessageCode {
				fmt.Println("Receive wrong message status code")
				continue
			}

			if err != nil {
				if err == io.EOF {
					validetPeer <- peer
					fmt.Printf("%s: peer disconnected \n", peer.Name)

					errorC <- err
					cancel()
					break
				} else {
					fmt.Printf("%s: message read failed: %s \n", peer.Name, err)
					continue
				}
			}

			var mess string
			err = inmsg.Decode(&mess)

			if err != nil {
				fmt.Println("Fail decode p2p message", err)

				errorC <- err
				cancel()
				break
			}

			fmt.Printf("[%s]: %s", peer.Name, mess)
		}

		wg.Done()
	}()
}
