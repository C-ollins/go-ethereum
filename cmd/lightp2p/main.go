package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/forkid"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	slog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/params"
)

func main() {
	dev, err := NewDevP2P()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	err = dev.start()
	if err != nil {
		fmt.Println("error strarting:", err)
		return
	}
}

var protocolLengths = map[uint]uint64{eth.ETH67: 17, eth.ETH66: 17}
var TxPool eth.TxPool
var td *big.Int = big.NewInt(44266186)

// BSC
var BCSMainnetHash common.Hash = common.HexToHash("0x0d21840abff46b96c84b2ac9e10e4f5cdaeb5693cb665db62a2f3b02d2d57b5b")
var blockchainHeadHash common.Hash = common.HexToHash("0x7dfeec2e037fdf2f3b90bab657b3d8d5a27edc73f8d2adb437dddb36d2aa2093")
var blockchainHead uint64 = 22260849

type devp2p struct {
	key         *ecdsa.PrivateKey
	forkID      forkid.ID
	chainParams *params.ChainConfig
	knownTxs    *knownCache
}

func NewDevP2P() (*devp2p, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	config := getConfig()
	forkID := forkid.NewID(config, BCSMainnetHash, blockchainHead)

	return &devp2p{
		chainParams: config,
		key:         key,
		forkID:      forkID,
		knownTxs:    newKnownCache(32768),
	}, nil
}

func (dev *devp2p) start() error {
	nodes, err := encodeEnodeUrl(Bootnodes)
	if err != nil {
		return err
	}
	config := p2p.Config{
		PrivateKey:      dev.key,
		MaxPeers:        10,
		BootstrapNodes:  nodes,
		Protocols:       dev.getProtocol(),
		EnableMsgEvents: true,
		Logger:          slog.Root(),
	}

	server := &p2p.Server{
		Config: config,
	}

	if err := server.Start(); err != nil {
		log.Println("Error in Starting Server")
	}

	peerEvent := make(chan *p2p.PeerEvent)
	eventSub := server.SubscribeEvents(peerEvent)
	defer eventSub.Unsubscribe()

	go func() {
		for {
			peerLen := server.PeerCount()
			log.Println("***********-PEERS LENGTH-************:", peerLen)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case a := <-peerEvent:

			if a.Error != "" {
				// fmt.Println(a.Error, a.RemoteAddress)
			}

			if a.Type != p2p.PeerEventTypeMsgRecv {

				continue
			}

			// fmt.Println(a.Type, a.MsgCode)

		case err := <-eventSub.Err():
			return err
		}
	}
}

func (dev *devp2p) getProtocol() []p2p.Protocol {

	var networkID uint64 = 56 // BSC
	protocol := []p2p.Protocol{
		{Name: eth.ProtocolName,
			Version: eth.ProtocolVersions[1],
			Length:  protocolLengths[eth.ETH66],
			Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
				_peer := eth.NewPeer(eth.ProtocolVersions[1], peer, rw, TxPool)

				err := _peer.Handshake(networkID, td, blockchainHeadHash, BCSMainnetHash, dev.forkID, func(id forkid.ID) error {
					
					return nil
				})

				if err != nil {
					return err
				}

				fmt.Println("Handshake new peer", _peer.Info().Protocols, peer.Info().Caps)

				for {
					msg, err := rw.ReadMsg()
					if err != nil {
						fmt.Println("error reading from peer:", err)
						return fmt.Errorf("error reading from peer: %v", err)
					}

					switch msg.Code {
					case eth.NewPooledTransactionHashesMsg:
						// fmt.Println("NewPooledTransactionHashesMsg")
						hashes := new(eth.NewPooledTransactionHashesPacket)
						if err := msg.Decode(hashes); err != nil {
							return fmt.Errorf("message %v: %v", msg, err)
						}

						err = _peer.RequestTxs(*hashes)
						if err != nil {
							fmt.Println("error requesting transaction:", err)
						}
						fmt.Println(*hashes)
						fmt.Println("Requested pooled txs:", len(*hashes))
					case eth.PooledTransactionsMsg:
						fmt.Println("Pooled transaction")

						var txs eth.PooledTransactionsPacket
						if err := msg.Decode(&txs); err != nil {
							fmt.Printf("message error %v: %v\n", msg, err)
							continue
						}
						for _, tx := range txs {
							var data interface{}
							err := json.Unmarshal(tx.Data(), &data)
							if err != nil {
								fmt.Println("error unmarshaling data:", err)
								continue
							}

							fmt.Println(data)
						}
					case eth.NewBlockMsg:
						fmt.Println("NewBlockMsg")
					case eth.TransactionsMsg:
						fmt.Println("TransactionsMsg")
					default:
						fmt.Println("Unknown:", msg.Code)
					}
				}
			},
			DialCandidates: nil,
			Attributes:     []enr.Entry{},
		},
	}
	return protocol
}
