package wallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wemeetagain/go-hdwallet"
	"io/ioutil"
	"math"
	"net/http"
)

const NumberAddress = 20

type WalletData struct {
	Items []Wallet
}

type Wallet struct {
	Address    string
	Balance    string
	PrivateKey string
}

// GroupRandomAddress to get list Wallet
func GroupRandomAddress(wallets *WalletData, channelWallets *chan WalletData) {
	data, err := randomAddress()
	if err == nil {
		balance := checkBalanceOfAddress(data.Address)
		data.Balance = fmt.Sprintf("%g", balance)
	}
	wallets.Items = append(wallets.Items, data)
	if len(wallets.Items) == NumberAddress {
		*channelWallets <- *wallets
	}
}

// randomAddress use pakage hdwallet to random gen privatekey, address of Bitcoin.
func randomAddress() (Wallet, error) {
	seed, err := hdwallet.GenSeed(128)
	if err != nil {
		return Wallet{}, err
	}
	masterKey := hdwallet.MasterKey(seed)
	childKey, _ := masterKey.Child(0)
	if err != nil {
		return Wallet{}, err
	}
	childPub := childKey.Pub()
	address := childPub.Address()
	keyString := hex.EncodeToString(childKey.Key)[2:]
	return Wallet{Address: address, PrivateKey: keyString}, nil
}

// checkBalanceOfAddress call get Api to blockstream Bitcoin to get balance of Address
func checkBalanceOfAddress(address string) float64 {
	response, err := http.Get("https://blockstream.info/api/address/" + address)
	if err == nil && response.StatusCode == 200 {
		result := response.Body
		defer result.Close()
		data, err := ioutil.ReadAll(result)
		if err != nil {
			return 0.0
		}
		mapData := make(map[string]interface{})
		json.Unmarshal(data, &mapData)
		chainStatData := mapData["chain_stats"].(map[string]interface{})
		mempoolStats := mapData["mempool_stats"].(map[string]interface{})
		fundedSum := chainStatData["funded_txo_sum"].(float64) + mempoolStats["funded_txo_sum"].(float64)
		spentSum := chainStatData["spent_txo_sum"].(float64) + mempoolStats["spent_txo_sum"].(float64)
		return (fundedSum - spentSum) / math.Pow10(8)
	}
	return 0.0
}
