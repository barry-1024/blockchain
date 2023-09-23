package tron

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"git.bipal.space/shared-lib/blockchain/utils"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"testing"

	"git.bipal.space/shared-lib/blockchain/client"
	"github.com/ethereum/go-ethereum/common"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/stretchr/testify/assert"
)

type treConfig struct {
	HDPath      string   `json:"hdPath"`
	Mnemonic    string   `json:"mnemonic"`
	PrivateKeys []string `json:"privateKeys"`
}

var (
	config = client.ChainConfiguration{
		ChainName: "Tron",
		Endpoints: []string{"https://api.trongrid.io/jsonrpc", "https://api.trongrid.io", "https://api.trongrid.io"},
		APIKey:    "bb198433-9f3d-4a66-b058-e230d7c2f8e2",
	}
	tConfig = client.ChainConfiguration{
		Endpoints: []string{"https://api.shasta.trongrid.io/jsonrpc",
			"https://api.shasta.trongrid.io",
			"https://api.shasta.trongrid.io"},
	}
	dockerInfo = treConfig{}
)

const (
	strABI = "[]"
	strBIN = "6080604052348015600f57600080fd5b50d38015601b57600080fd5b50d28015602757600080fd5b50603f8060356000396000f3fe6080604052600080fdfea26474726f6e58221220712baa34db109019874f477a57c6466ec66f286ede31713f912cb0bf9b0e4fb164736f6c634300080b0033"

	treDockerDomain = "http://host.docker.internal:9090"
)

func TestMain(m *testing.M) {
	//setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	response, err := http.Get(treDockerDomain + "/admin/accounts-json")
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(content, &dockerInfo)
}

func TestPubkeyToAddress(t *testing.T) {
	pubKey := "0404B604296010A55D40000B798EE8454ECCC1F8900E70B1ADF47C9887625D8BAE3866351A6FA0B5370623268410D33D345F63344121455849C9C28F9389ED9731"
	pubValue, err := hex.DecodeString(pubKey)
	assert.Nil(t, err, "parse pubkey should success")
	ecdsaPubKey, err := ecrypto.UnmarshalPubkey(pubValue)
	assert.Nil(t, err, "public key invalid")
	addrByte := ecrypto.PubkeyToAddress(*ecdsaPubKey)
	address := make([]byte, 0, 32)
	address = append(address, addressPrefix)
	address = append(address, addrByte.Bytes()...)
	target, err := hex.DecodeString("412A2B9F7641D0750C1E822D0E49EF765C8106524B")
	assert.Nil(t, err, "decode address failed")
	assert.Equal(t, bytes.Compare(address, target), 0, "should equal")
}

func TestDecimalsOf(t *testing.T) {
	client, err := NewTronClient(&tConfig)
	assert.Nil(t, err, "create client failed")
	// USDT
	decimal, err := client.DecimalsOf("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t")
	assert.Nil(t, err, "get decimals failed")
	assert.Equal(t, decimal, uint8(6), "decimals should be same")
}

func TestBalanceAt(t *testing.T) {
	client, err := NewTronClient(&tConfig)
	assert.Nil(t, err, "create client failed")
	// address is magic's address
	//address := "TR2giB1C897abNrR1bJsuhREcz4oUoGhEG"
	address := "TPTgNUHoJWE5SWUqNb75emXW2yYZvqU5hb"
	balance, err := client.BalanceAt(address)
	assert.Nil(t, err, "balance at failed")
	fmt.Println("balanceAt: ", balance)
	assert.Greater(t, balance.Uint64(), uint64(10), "account balance should > 10")
}

func TestBalanceOf(t *testing.T) {
	client, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	// address is magic's address
	address := "TR2giB1C897abNrR1bJsuhREcz4oUoGhEG"
	contract := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	balance, err := client.BalanceOf(contract, address)
	assert.Nil(t, err, "balance of should success")
	assert.Greater(t, balance.Uint64(), uint64(0), "should > 0")
}

func TestSymbolOf(t *testing.T) {
	client, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	// address is magic's address
	contract := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	symbol, err := client.SymbolOf(contract)
	assert.Nil(t, err, "symbol failed")
	assert.Equal(t, "USDT", symbol, "symbol incorrect")
}

func TestTotalSupply(t *testing.T) {
	client, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	// address is magic's address
	contract := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	total, err := client.TotalSupplyOf(contract)
	assert.Nil(t, err, "get total supply failed")
	assert.Greater(t, total.Uint64(), uint64(1000000), "total should big")
}

func TestSuggestFee(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	to := "TY3ba19r36hcvQVotETd3tH4HDC14Ucha6"
	td := client.Transaction{
		//Contract: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		ABI:    trc20ABIName,
		Method: "approve",
		From:   "TExz1XRq54uQftRkAs3rXKGFKLXJ9JR4rJ",
		To:     "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Amount: big.NewInt(10),
		Fee:    &client.FeeLimit{},
	}
	data, err := tclient.GetTransactionDataByABI(td.Method, trc20ABIName, to, td.Amount)
	assert.Nil(t, err, "get transaction data failed")
	td.Data = data
	fee, err := tclient.GetSuggestFee(&td)
	assert.Nil(t, err, "get suggest fee failed")
	assert.Greater(t, fee.Gas.Uint64(), uint64(0), "fee should > 0")
}

func TestSignature(t *testing.T) {
	privateKey := utils.TronPrivateKeyFrom
	pk, _ := ecrypto.HexToECDSA(privateKey)
	data := "e53d19d58f686073c60f4779fe9fb48301c511f604b3dd2e45388a7b135c5f90"
	buf, _ := hex.DecodeString(data)
	sig, err := ecrypto.Sign(buf, pk)
	assert.Nil(t, err, "signature wrong")
	assert.Equal(t, hex.EncodeToString(sig), "82ad3501c5ce19c1e2f95441902fac255de6953f66ad22557ee533428e80ee4c6c3cedc6055950848f165997bc9a9278900dc740b8768dfa03d9330e6d964fde00", "signature wrong")

	privateKey = utils.TronPrivateKeySign
	pk, _ = ecrypto.HexToECDSA(privateKey)
	data = "1b59e60f3ec70116ae065f59f3277e19bfbc0757619524d7d48545eab1b7a42c"
	buf, _ = hex.DecodeString(data)
	sig, err = ecrypto.Sign(buf, pk)
	fmt.Println(hex.EncodeToString(sig), err)
}

func TestCallContract(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	data, err := tclient.GetTransactionDataByABI("totalSupply", trc20ABIName)
	assert.Nil(t, err, "generate data failed")
	// call usdt totalSupplyOf
	td := client.Transaction{
		//Contract: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		From: "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ",
		To:   "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Data: data,
	}
	res, err := tclient.CallContract(&td)
	assert.Nil(t, err, "call contract failed")
	assert.NotEqual(t, bytes.Compare(res, []byte("00000000000000000000000000000000")), 0, "should not be empty")
}

func TestCallContractWithData(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	ownerAddr := "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ"
	data, err := tclient.GetTransactionDataByABI("balanceOf", trc20ABIName, ownerAddr)
	assert.Nil(t, err, "generate data failed")
	td := client.Transaction{
		//Contract: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		From: "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ",
		To:   "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Data: data,
	}
	res, err := tclient.CallContract(&td)
	assert.Nil(t, err, "call contract failed")
	assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", hex.EncodeToString(res), "should equal")
}

func TestGetGasPrice(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	gasPrice, tipPrice, err := tclient.GetGasPrice()
	assert.Nil(t, err, "get gas price failed")
	assert.Greater(t, gasPrice.Uint64(), uint64(0), "gas price should > 0")
	assert.Equal(t, tipPrice.Uint64(), uint64(0), "tip price should be nil")
	fmt.Println("gasPrice: ", gasPrice, "tipPrice: ", tipPrice)
}

func TestABiCall(t *testing.T) {
	tclient, err := NewTronClient(&tConfig)
	assert.Nil(t, err, "create client failed")
	swapABI, err := ioutil.ReadFile("./swap.abi")
	assert.Nil(t, err, "read abi failed")
	err = tclient.RegisterABI("swapContract", string(swapABI))
	assert.Nil(t, err, "register abi failed")
	tokenIn, err := tclient.AddressFromString("TVnFbxVHgu5EgCocuSB4AwKVWyscPgAodE")
	assert.Nil(t, err, "address_failed")
	tokenOut, err := tclient.AddressFromString("TEYPs9bE4z4ZuU99rk9ekbbBbaR3bYooGu")
	assert.Nil(t, err, "address_failed")
	path := []common.Address{tokenIn, tokenOut}
	provider, _ := hex.DecodeString("58eb47b3")
	providerAddr := [4]byte{}
	copy(providerAddr[:], provider)
	data, err := tclient.GetTransactionDataByABI("getAmountOut", "swapContract", providerAddr, path, big.NewInt(100000000))
	assert.Nil(t, err, "generate data failed")
	result, err := tclient.CallContract(&client.Transaction{
		//From: "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ",
		ABI:     "swapContract",
		From:    "TJzstqwcSEeiRQYD7hMPDpaSssc9Tw6TTS",
		To:      "TJzstqwcSEeiRQYD7hMPDpaSssc9Tw6TTS",
		ChainID: big.NewInt(2),
		Data:    data,
	})
	assert.Nil(t, err, "call contract failed")
	fields, err := tclient.UnpackByABI("getAmountOut", "swapContract", result)
	assert.Nil(t, err, "should success")
	fmt.Println(len(fields), fields)
}

func TestDeployContract(t *testing.T) {
	tConfig := client.ChainConfiguration{
		Endpoints: []string{"http://host.docker.internal:9090/jsonrpc",
			"http://host.docker.internal:9090",
			"http://host.docker.internal:9090"},
	}
	tclient, err := NewTronClient(&tConfig)
	privateKey := dockerInfo.PrivateKeys[0]
	pk, _ := ecrypto.HexToECDSA(privateKey)
	publicKey := pk.PublicKey
	from := address.PubkeyToAddress(publicKey)
	assert.Nil(t, err, "create client failed")
	td := client.Transaction{}
	td.ChainID = new(big.Int).SetUint64(1)
	td.From = from.String()
	td.To = ""
	td.Method = ""
	td.Amount = big.NewInt(0)
	td.Nonce = 0
	td.Fee = &client.FeeLimit{}
	td.Fee.GasFeeCap = new(big.Int).SetUint64(10000)
	td.Fee.Gas = new(big.Int).SetUint64(200000090)
	td.Fee.GasTipCap = new(big.Int).SetUint64(0)
	message, hash, addr, err := tclient.DeployContract(strABI, strBIN, &td)

	assert.Nil(t, err, "pre-deploy failed")
	assert.NotEqual(t, addr, "", "addr is empty")
	sig, err := ecrypto.Sign(hash, pk)
	assert.Nil(t, err, "sign message failed")
	txID, err := tclient.BroadcastTransaction(message, sig)
	assert.Nil(t, err, "broadcast failed")
	assert.True(t, bytes.Equal(txID, hash), "id incorrect")
}

func TestContractTransaction(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	data, err := tclient.GetTransactionDataByABI("transfer", trc20ABIName, "TLi7bUTJyvGddcdyMvUmjYpyU3JV5u381U", big.NewInt(100))
	fmt.Println(hex.EncodeToString(data))
	assert.Nil(t, err, "generate data failed")
	td := client.Transaction{
		From:   "TDkA3HphEwk8FutZXLnoSF1zeSVNWz1ov5",
		To:     "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		Method: "transfer",
		ABI:    trc20ABIName,
		Data:   data,
		Amount: big.NewInt(0),
	}
	gas, err := tclient.EstimateGas(&td)
	assert.Nil(t, err, "estimate gas failed")
	gasPrice, _, err := tclient.GetGasPrice()
	assert.Nil(t, err, "get gas price failed")
	feeLimit := client.FeeLimit{
		Gas:       big.NewInt(int64(gas)),
		GasFeeCap: gasPrice,
	}
	td.Fee = &feeLimit
	ret, err := tclient.CallContract(&td)
	fmt.Println(string(ret), err)
	trans, hash, err := tclient.GetTransaction(&td)
	assert.Nil(t, err, "get transaction failed")
	fmt.Println(string(trans), string(hash))
}

func TestTransaction(t *testing.T) {
	tConfig := client.ChainConfiguration{
		Endpoints: []string{"http://host.docker.internal:9090/jsonrpc",
			"http://host.docker.internal:9090",
			"http://host.docker.internal:9090"},
	}
	pkFrom, err := ecrypto.HexToECDSA(utils.TronPkFrom)
	assert.Nil(t, err, "generate private key failed")
	pkTo, err := ecrypto.HexToECDSA(utils.TronPkTo)
	assert.Nil(t, err, "generate private key failed")
	tclient, err := NewTronClient(&tConfig)
	assert.Nil(t, err, "create client failed")
	fmt.Println("from: ", address.PubkeyToAddress(pkFrom.PublicKey).String())
	td := client.Transaction{
		Method: "",
		From:   address.PubkeyToAddress(pkFrom.PublicKey).String(),
		To:     address.PubkeyToAddress(pkTo.PublicKey).String(),
		Amount: big.NewInt(1000),
		Fee:    &client.FeeLimit{},
	}
	td.Fee.Gas = big.NewInt(1000)
	message, hash, err := tclient.GetTransaction(&td)
	assert.Nil(t, err, "get transaction failed")
	assert.NotNil(t, message, "message nil")
	assert.NotNil(t, hash, "message nil")
	sig, err := ecrypto.Sign(hash, pkFrom)
	assert.Nil(t, err, "sign message failed")

	txID, err := tclient.BroadcastTransaction(message, sig)
	assert.Nil(t, err, "broadcast failed")
	assert.True(t, bytes.Equal(txID, hash), "id incorrect")
}

func TestGetOnChainTransaction(t *testing.T) {
	tclient, err := NewTronClient(&tConfig)
	//tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	info, err := tclient.GetTransactionByHash("22f671e62356915fdb9df097d8338ae854f9334bf02edc61147f5c991086aba6")
	assert.Nil(t, err, "get transaction failed")
	assert.Greater(t, len(info.Logs), 0, "logs should be >1")
	for _, log := range info.Logs {
		fmt.Println(string(log.Topics[0]), hex.EncodeToString(log.Topics[0]))
	}
	assert.Equal(t, info.Tx.To, "TEEoHaP7SS5WNAkDwf81uBXqpR6Hjb849N", "wrong address")

	fmt.Println(info, info.Tx.To, hex.EncodeToString(info.Tx.Data), *info.Tx.Fee)
}

func TestTransferLog(t *testing.T) {
	cli, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	info, err := cli.GetTransactionByHash("e3f747b265c39125b91c319526a640e49310b19b072515c39e1616364393a0b1")
	assert.Equal(t, 1, int(len(info.Logs)), "logs should be 1")
	assert.Greater(t, len(info.Logs[0].Topics), 0, "logs should be > 0")
	data := map[string]string{}
	err = json.Unmarshal(info.Logs[0].Data, &data)
	assert.Nil(t, err, "unmarshal failed")
	fmt.Println(data)
}

func TestAddressFromPrivateKey(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	address, err := tclient.AddressFromPrivateKey(utils.TronPrivateKeyFrom)
	assert.Nil(t, err, "private key failed")
	assert.Equal(t, address, "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ", "address failed")

	address, err = tclient.AddressFromPrivateKey("0x" + utils.TronPrivateKeyFrom)
	assert.Nil(t, err, "private key failed")
	assert.Equal(t, address, "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ", "address failed")
}

func TestIsValidAddress(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	valid := tclient.IsValidAddress("TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ")
	assert.Equal(t, valid, true, "address failed")

	valid = tclient.IsValidAddress("TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8J00")
	assert.Equal(t, valid, false, "address failed")

	valid = tclient.IsValidAddress("")
	assert.False(t, valid, "address failed")
}

func TestFunctionSelector(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	selector := tclient.getFunctionSelector(trc20ABIName, "transferFrom")
	assert.Nil(t, err, "function selector failed")
	assert.Equal(t, selector, "transferFrom(address,address,uint256)", "function selector failed")
}

func TestAllowance(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	value, err := tclient.Allowance("TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t", "TYg7Uh7fG8ZQxRvWRpFziHzWc8YJLX8JtJ", "TYp5ZMYJNSPw8JRmqXRjy8tMVMK1hvDuPe")
	assert.Nil(t, err, "allowance failed")
	assert.Equal(t, value.Int64(), int64(0), "allowance failed")
}

func TestPublicKeyHexToAddress(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	addr, err := tclient.PublicKeyHexToAddress("042f648f8f37f0a108cf4df48a094b4c01d322374a2bb4afbb1afa594280e69e073991ba0aeb1d1a2317088ee14dcf181edd9d46705015aaff0fa2ec366d48cb5a")
	assert.Nil(t, err, "public key to address failed")
	assert.Equal(t, addr, "TY6mooR5J3yeoNo1uANG4sjq4CJyT5UUxq", "public key to address failed")
}

func TestGetLackedGas(t *testing.T) {
	tclient, err := NewTronClient(&config)
	assert.Nil(t, err, "create client failed")
	tclient.GetLackedGas("TSFbrBgDwnU41oLowse5cEZsyQM2fU2mAB", 1000, big.NewInt(420), 1024)
}
