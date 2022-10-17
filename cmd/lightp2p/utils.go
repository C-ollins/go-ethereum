package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
)

// the main Ethereum network.
var Bootnodes = []string{
	"enode://1cc4534b14cfe351ab740a1418ab944a234ca2f702915eadb7e558a02010cb7c5a8c295a3b56bcefa7701c07752acd5539cb13df2aab8ae2d98934d712611443@52.71.43.172:30311",
	"enode://28b1d16562dac280dacaaf45d54516b85bc6c994252a9825c5cc4e080d3e53446d05f63ba495ea7d44d6c316b54cd92b245c5c328c37da24605c4a93a0d099c4@34.246.65.14:30311",
	"enode://5a7b996048d1b0a07683a949662c87c09b55247ce774aeee10bb886892e586e3c604564393292e38ef43c023ee9981e1f8b335766ec4f0f256e57f8640b079d5@35.73.137.11:30311",
}

func encodeEnodeUrl(enodeUrl []string) ([]*enode.Node, error) {
	bootNodes := []*enode.Node{}
	for i := 0; i < len(enodeUrl); i++ {
		fm, err := enode.ParseV4(enodeUrl[i])
		if err != nil {
			return nil, err
		}
		bootNodes = append(bootNodes, fm)
	}

	return bootNodes, nil
}

func getConfig() *params.ChainConfig {
	var config *params.ChainConfig
	bytes, err := ioutil.ReadFile("genesis.json")

	if err != nil {
		fmt.Println("Unable to load config")
		os.Exit(1)
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		fmt.Println("Json decode error")
		os.Exit(1)
	}

	fmt.Println("Succes in parsing JSON")

	return config

}
