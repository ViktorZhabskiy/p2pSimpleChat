package main

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ethereum/go-ethereum/p2p"
)

const MessageCode = 0

func createProtocol(storage *Storage) p2p.Protocol {
	return p2p.Protocol{
		Name:    "chat",
		Version: 1,
		Length:  1,
		Run:     decorateProto(storage),
	}
}

func decorateProto(storage *Storage) func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	return func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
		outC := make(chan string)
		errorC := make(chan error)

		peer := &Peer{
			Peer: p,
			RW:   rw,
			OutC: outC,
			Name: fmt.Sprintf("%s(%s)", p.Name(), p.RemoteAddr().String()),
		}

		storage.PeerC <- peer

		wg := sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())

		TxMessage(ctx, cancel, peer, &wg, errorC)

		RxMessage(ctx, cancel, storage.DeletePC, peer, &wg, errorC)

		err := <-errorC

		wg.Wait()

		return err
	}
}

func TxMessage(ctx context.Context, cancel context.CancelFunc, peer *Peer, wg *sync.WaitGroup, errorC chan error) {
	wg.Add(1)

	go func() {
	breakLoop:
		for {
			select {
			case <-ctx.Done():
				break breakLoop
			case message := <-peer.OutC:
				if err := p2p.Send(peer.RW, MessageCode, message); err != nil {
					fmt.Printf("Fail send message err %s", err)
					cancel()
					errorC <- err
					break breakLoop
				}
			}
		}

		wg.Done()
	}()
}

func RxMessage(ctx context.Context, cancel context.CancelFunc, deletePC chan *Peer, peer *Peer, wg *sync.WaitGroup, errorC chan error) {
	wg.Add(1)

	go func() {
	breakLoop:
		for {
			select {
			case <-ctx.Done():
				break breakLoop
			default:
				inmsg, err := peer.RW.ReadMsg()

				if inmsg.Code != MessageCode {
					fmt.Println("Receive wrong message status code")
					continue
				}

				if err != nil {
					if err == io.EOF {
						deletePC <- peer

						cancel()
						errorC <- err
						break breakLoop

					} else {
						fmt.Printf("%s: message read failed: %s \n", peer.Name, err)
						continue
					}
				}

				var mess string
				err = inmsg.Decode(&mess)

				if err != nil {
					fmt.Println("Fail decode p2p message", err)
					cancel()
					errorC <- err
					break breakLoop
				}

				fmt.Printf("[%s]: %s", peer.Name, mess)
			}

		}

		wg.Done()
	}()
}
