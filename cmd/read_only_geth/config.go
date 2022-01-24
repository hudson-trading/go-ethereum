package main

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/node"
)

func makeFullNode(datadir string) (*node.ReadOnlyNode, ethapi.Backend) {
	stack, cfg := makeConfigNode(datadir)
	backend, _ := utils.RegisterReadOnlyEthService(stack, cfg)
	return stack, backend
}

func makeConfigNode(datadir string) (*node.ReadOnlyNode, *ethconfig.Config) {
	node_cfg := defaultNodeConfig()
	eth_cfg := ethconfig.Defaults
	utils.SetReadOnlyNodeConfig(&node_cfg, datadir)
	stack, err := node.NewReadOnly(&node_cfg)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}
	utils.SetReadOnlyEthConfig(&eth_cfg)
	return stack, &eth_cfg
}

func defaultNodeConfig() node.Config {
	cfg := node.DefaultConfig
	cfg.HTTPModules = append(cfg.HTTPModules, "eth")
	cfg.WSModules = append(cfg.WSModules, "eth")
	cfg.IPCPath = "geth.ipc"
	return cfg
}
