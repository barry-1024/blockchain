package clients

import (
	"fmt"
	"sync"

	"git.bipal.space/shared-lib/blockchain/client"
	"git.bipal.space/shared-lib/blockchain/eth"
	"git.bipal.space/shared-lib/blockchain/tron"
)

var (
	clientMap map[string]client.BlockChainClient = make(map[string]client.BlockChainClient)
	lock      sync.Mutex
)

func GetClientByConfig(config *client.ChainConfiguration) (client.BlockChainClient, error) {
	lock.Lock()
	defer lock.Unlock()

	key := fmt.Sprintf("%s_%s", config.ChainName, config.Endpoints[0])
	if cli, ok := clientMap[key]; ok {
		return cli, nil
	}
	cli, err := NewClient(config)
	if err == nil {
		clientMap[key] = cli
	}
	return cli, err
}

func NewClient(config *client.ChainConfiguration) (client.BlockChainClient, error) {
	if config.ChainName == "Tron" {
		return tron.NewTronClient(config)
	} else {
		cli, err := eth.NewEthClient(config)
		if err != nil {
			return nil, err
		}
		cli.SupportEIP1559 = config.SupportEIP1559
		return cli, nil
	}
	return nil, fmt.Errorf("chain=%s is not supported", config.ChainName)
}

type ClientManager struct {
	chainClients map[uint64]client.BlockChainClient
}

func NewClientManager(configs map[uint64]*client.ChainConfiguration) (*ClientManager, error) {
	cm := &ClientManager{
		chainClients: make(map[uint64]client.BlockChainClient),
	}
	for cid, config := range configs {
		cli, err := NewClient(config)
		if err != nil {
			return nil, err
		}
		cm.chainClients[cid] = cli
	}
	return cm, nil
}

func (cm *ClientManager) GetClient(chainID uint64) (client.BlockChainClient, error) {
	if cli, ok := cm.chainClients[chainID]; ok {
		return cli, nil
	}
	return nil, fmt.Errorf("chain=%d not found", chainID)
}
