package entity

import (
	"context"
	"encoding/json"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	corehost "github.com/libp2p/go-libp2p/core/host"
	peer "github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type Node struct {
	Host    corehost.Host
	PubSub  *pubsub.PubSub
	Topic   *pubsub.Topic
	Sub     *pubsub.Subscription
	Mempool *Mempool
	Ctx     context.Context
}

// NewNode creates a libp2p host, connects to bootstrap peers, and subscribes to pubsub topic
func NewNode(ctx context.Context, mempool *Mempool, peers []string) (*Node, error) {
	// Create a new libp2p Host
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}

	// Dial bootstrap peers
	for _, addr := range peers {
		if addr == "" {
			continue
		}
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			fmt.Println("Invalid multiaddr:", err)
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			fmt.Println("Failed to parse peer addr:", err)
			continue
		}
		if err := h.Connect(ctx, *info); err != nil {
			fmt.Println("Connection failed:", err)
		} else {
			fmt.Println("Connected to peer:", info.ID)
		}
	}

	// Set up PubSub
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}
	topic, err := ps.Join("tx-topic")
	if err != nil {
		return nil, err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	node := &Node{Host: h, PubSub: ps, Topic: topic, Sub: sub, Mempool: mempool, Ctx: ctx}
	go node.listen()

	fmt.Println("Node ID:", h.ID())
	for _, a := range h.Addrs() {
		fmt.Printf(" - %s/p2p/%s\n", a, h.ID().String())
	}
	return node, nil
}

func (n *Node) listen() {
	for {
		msg, err := n.Sub.Next(n.Ctx)
		if err != nil {
			fmt.Println("Sub error:", err)
			continue
		}
		var tx Transaction
		if err := json.Unmarshal(msg.Data, &tx); err != nil {
			continue
		}
		fmt.Println("RX from peer:", tx.ID)
		n.Mempool.Add(tx)
	}
}

func (n *Node) Broadcast(tx Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return n.Topic.Publish(n.Ctx, data)
}
