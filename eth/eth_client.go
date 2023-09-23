package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ecrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"reflect"
	"strings"
	"sync"

	"git.bipal.space/shared-lib/blockchain/client"
)

// EthClient implements BlockChain interface
type EthClient struct {
	client   *ethclient.Client
	abiMap   sync.Map
	erc20Abi *abi.ABI

	chainID        *big.Int
	debugClient    *backends.SimulatedBackend
	SupportEIP1559 bool
}

const (
	nativeAsset         = "0x0000000000000000000000000000000000000000"
	maticNativeAsset    = "0x0000000000000000000000000000000000001010"
	erc20ABIName        = "erc20"
	nativeAssetDecimals = 18
	// erc20Abi is generate from erc20.abi file, just remove spaces and escape the double quotes
	erc20Abi = "[{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

// NewEthClient creates and init the client for ethereum
func NewEthClient(config *client.ChainConfiguration) (*EthClient, error) {
	client := &EthClient{}
	ethClient, err := ethclient.Dial(config.Endpoints[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to endpoint=%s", config.Endpoints[0])
	}
	client.client = ethClient
	client.abiMap = sync.Map{}
	if err := client.RegisterABI(erc20ABIName, erc20Abi); err != nil {
		return nil, fmt.Errorf("register erc20 abi failed, err=%s", err)
	}
	erc20, err := abi.JSON(strings.NewReader(erc20Abi))
	if err != nil {
		return nil, fmt.Errorf("failed to parse the abi, err=%s", err)
	}
	client.SupportEIP1559 = config.SupportEIP1559
	client.erc20Abi = &erc20
	client.chainID = config.ChainID
	return client, nil
}

func (e *EthClient) TransferData(to string, value *big.Int) ([]byte, error) {
	method := "transfer"
	toAddr := common.HexToAddress(to)
	return e.GetTransactionDataByABI(method, erc20ABIName, toAddr, value)
}

func (e *EthClient) RegisterABI(name, abiStr string) error {
	compiled, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return err
	}
	e.abiMap.Store(name, &compiled)
	return nil
}

func (e *EthClient) UnpackByABI(method, name string, data []byte) ([]interface{}, error) {
	compiled, err := e.GetABIByName(name)
	if err != nil {
		return nil, err
	}
	return compiled.Unpack(method, data)
}

func (e *EthClient) GetABIByName(name string) (*abi.ABI, error) {
	cabi, ok := e.abiMap.Load(name)
	if !ok {
		return nil, fmt.Errorf("abi=%s not found", name)
	}
	compiled := cabi.(*abi.ABI)
	return compiled, nil
}

// SetClient can be used for mock purpose
func (e *EthClient) SetClient(cli *backends.SimulatedBackend) {
	e.debugClient = cli
}

// BalanceAt reads the balance of eth
func (e *EthClient) BalanceAt(address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	value, err := e.client.BalanceAt(context.Background(), addr, nil)
	return value, err
}

func (e *EthClient) Allowance(contract, owner, address string) (*big.Int, error) {
	method := "allowance"
	contractAddr := common.HexToAddress(contract)
	ownerAddr := common.HexToAddress(owner)
	addressAddr := common.HexToAddress(address)
	input, err := e.GetTransactionDataByABI(method, erc20ABIName, ownerAddr, addressAddr)
	if err != nil {
		return nil, fmt.Errorf("pack message failed, err=%s", err)
	}
	msg := ethereum.CallMsg{From: ownerAddr, To: &contractAddr, Data: input}
	output, err := e.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("call contract failed, err=%s", err)
	}
	// unpack result
	result, err := e.UnpackByABI(method, erc20ABIName, output)
	if err != nil {
		return nil, fmt.Errorf("unpack result failed, err=%s", err)
	}
	if len(result) != 1 {
		return nil, fmt.Errorf("invalid result, result=%v", result)
	}
	return e.AbiConvertToInt(result[0]), nil
}

func (e *EthClient) ApproveData(contract, owner, spender string, amount *big.Int) ([]byte, error) {
	method := "approve"
	spenderAddr := common.HexToAddress(spender)
	return e.GetTransactionDataByABI(method, erc20ABIName, spenderAddr, amount)
}

// BalanceOf reads the balance of
func (e *EthClient) BalanceOf(contract, from string) (*big.Int, error) {
	method := "balanceOf"
	// convert address
	fromAddr := common.HexToAddress(from)
	contractAddr := common.HexToAddress(contract)
	// pack params
	input, err := e.GetTransactionDataByABI(method, erc20ABIName, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("pack message failed, err=%s", err)
	}
	// call contract
	msg := ethereum.CallMsg{From: fromAddr, To: &contractAddr, Data: input}
	output, err := e.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("call contract failed, err=%s", err)
	}
	// parse result
	// res, err := e.erc20Abi.Unpack(method, output)
	fmt.Println("output: ", hex.EncodeToString(output))
	res, err := e.UnpackByABI(method, erc20ABIName, output)
	if err != nil {
		return nil, fmt.Errorf("unpack returned message failed, err=%s", err)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("wrong return format")
	}
	balance := *abi.ConvertType(res[0], new(*big.Int)).(**big.Int)
	return balance, nil
}

func (e *EthClient) callERC20(contract, method string) ([]interface{}, error) {
	contractAddr := common.HexToAddress(contract)
	input, err := e.GetTransactionDataByABI(method, erc20ABIName)
	if err != nil {
		return nil, fmt.Errorf("pack message failed, err=%s", err)
	}
	msg := ethereum.CallMsg{From: common.Address{}, To: &contractAddr, Data: input}
	output, err := e.client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return nil, fmt.Errorf("call contract failed, err=%s", err)
	}
	res, err := e.UnpackByABI(method, erc20ABIName, output)
	if err != nil {
		return nil, fmt.Errorf("unpack returned message failed, err=%s", err)
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("wrong return format")
	}
	return res, nil
}

// DecimalOf returns the decimals of a contract
func (e *EthClient) DecimalsOf(contract string) (uint8, error) {
	method := "decimals"
	res, err := e.callERC20(contract, method)
	if err != nil {
		return 0, err
	}
	if reflect.ValueOf(res[0]).Kind() == reflect.Ptr {
		value := *abi.ConvertType(res[0], new(*big.Int)).(**big.Int)
		return uint8(value.Uint64()), nil
	}
	decimals := *abi.ConvertType(res[0], new(uint8)).(*uint8)
	return decimals, nil
}

// TotalSupplyOf returns the total supply of a contract
func (e *EthClient) TotalSupplyOf(contract string) (*big.Int, error) {
	method := "totalSupply"
	res, err := e.callERC20(contract, method)
	if err != nil {
		return nil, err
	}
	supply := *abi.ConvertType(res[0], new(*big.Int)).(**big.Int)
	return supply, nil
}

// SymbolOf returns the symbol of a contract
func (e *EthClient) SymbolOf(contract string) (string, error) {
	method := "symbol"
	res, err := e.callERC20(contract, method)
	if err != nil {
		return "", err
	}
	symbol := *abi.ConvertType(res[0], new(string)).(*string)
	return symbol, nil
}

// GetNonce returns the nonce for an address
func (e *EthClient) GetNonce(address string) (uint64, error) {
	addr := common.HexToAddress(address)
	return e.client.PendingNonceAt(context.Background(), addr)
}

// GetNonceByNumber returns the nonce for an address at a block number
func (e *EthClient) GetNonceByNumber(address string, blockNumber *big.Int) (uint64, error) {
	addr := common.HexToAddress(address)
	return e.client.NonceAt(context.Background(), addr, blockNumber)
}

// GetSuggestFee returns the suggested fee for a transaction
func (e *EthClient) GetSuggestFee(td *client.Transaction) (*client.FeeLimit, error) {
	feeLimit := &client.FeeLimit{}
	// tip cap
	tipCap, err := e.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get suggest gas tip failed, err=%s", err)
	}
	feeCap, err := e.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get suggest gas price failed, err=%s", err)
	}
	contractAddr := common.HexToAddress(td.To)
	// gas limit
	fromAddress := common.HexToAddress(td.From)
	gas, err := e.client.EstimateGas(context.Background(), ethereum.CallMsg{From: fromAddress, To: &contractAddr,
		Data: td.Data, Value: td.Amount})
	if err != nil {
		return nil, fmt.Errorf("estimate gas failed, err=%s", err)
	}

	//gas limit = gas * 120% / 100%
	multiGas := new(big.Int).Mul(big.NewInt(int64(gas)), big.NewInt(120))
	gasLimit := new(big.Int).Div(multiGas, big.NewInt(100))

	feeLimit.GasTipCap = tipCap
	feeLimit.GasFeeCap = feeCap
	feeLimit.Gas = gasLimit
	return feeLimit, nil
}

func (e *EthClient) EstimateGas(td *client.Transaction) (uint64, error) {
	//todo::后面优化，主币的gas_limit 限制为21000
	if len(td.Data) <= 0 {
		return 21000, nil
	}

	toAddr := (*common.Address)(nil)
	if td.To != "" {
		contractAddr := common.HexToAddress(td.To)
		toAddr = &contractAddr
	}

	gas, err := e.client.EstimateGas(context.Background(), ethereum.CallMsg{From: common.HexToAddress(td.From),
		To: toAddr, Data: td.Data, Value: td.Amount})
	if err != nil {
		return 0, err
	}

	//强制增加20%的gas_limit
	return gas * 120 / 100, nil
}

// DeployContract generate the transactions that deploy an contract
// The address can be calculated by calling ContractAddressOf function
func (e *EthClient) DeployContract(contractAbi, contractBin string, td *client.Transaction) (
	[]byte, []byte, string, error) {
	parsed, err := abi.JSON(strings.NewReader(contractAbi))
	if err != nil {
		return nil, nil, "", fmt.Errorf("parse abi failed, err=%s", err)
	}
	byteCode := common.FromHex(contractBin)
	input, err := parsed.Pack("")
	if err != nil {
		return nil, nil, "", fmt.Errorf("pack message failed, err=%s", err)
	}
	data := append(byteCode, input...)
	baseTx := &types.DynamicFeeTx{
		ChainID:   e.chainID,
		Nonce:     td.Nonce,
		GasFeeCap: td.Fee.GasFeeCap,
		GasTipCap: td.Fee.GasTipCap,
		Gas:       td.Fee.Gas.Uint64(),
		To:        nil,
		Value:     big.NewInt(0),
		Data:      data,
	}
	tx := types.NewTx(baseTx)
	message, err := tx.MarshalBinary()
	if err != nil {
		return nil, nil, "", fmt.Errorf("encode message failed, err=%s", err)
	}
	signer := types.NewLondonSigner(e.chainID)
	hash := signer.Hash(tx)
	contractAddr, err := e.contractAddressOf(contractAbi, contractBin, td)
	if err != nil {
		return nil, nil, "", fmt.Errorf("calculate contract address failed, err=%s", err)
	}
	return message, hash.Bytes(), contractAddr, nil
}

// GetTransactionData generates the data of the transaction
func (e *EthClient) GetTransactionData(method string, abiDesc string, args ...interface{}) ([]byte, error) {
	if abiDesc == "" {
		return nil, fmt.Errorf("empty abi")
	}
	methodAbi, err := abi.JSON(strings.NewReader(abiDesc))
	if err != nil {
		return nil, fmt.Errorf("parse abi failed, err=%s", err)
	}
	return methodAbi.Pack(method, args...)
}

// GetTransactionDataByABI generate the data from registered abi
func (e *EthClient) GetTransactionDataByABI(method string, abiName string, args ...interface{}) ([]byte, error) {
	compiled, err := e.GetABIByName(abiName)
	if err != nil {
		return nil, err
	}
	return compiled.Pack(method, args...)
}

func (e *EthClient) generateTx(toAddr *common.Address, td *client.Transaction) *types.Transaction {
	var tx *types.Transaction
	if e.SupportEIP1559 {
		baseTx := &types.DynamicFeeTx{
			ChainID:   e.chainID,
			Nonce:     td.Nonce,
			GasFeeCap: td.Fee.GasFeeCap,
			GasTipCap: td.Fee.GasTipCap,
			Gas:       td.Fee.Gas.Uint64(),
			To:        toAddr,
			Value:     td.Amount,
			Data:      td.Data,
		}
		tx = types.NewTx(baseTx)
	} else {
		baseTx := &types.LegacyTx{
			Nonce:    td.Nonce,
			GasPrice: td.Fee.GasFeeCap,
			Gas:      td.Fee.Gas.Uint64(),
			To:       toAddr,
			Value:    td.Amount,
			Data:     td.Data,
		}
		tx = types.NewTx(baseTx)
	}
	return tx
}

func (e *EthClient) GetLatestBlockNumber() (*big.Int, error) {
	header, err := e.client.HeaderByNumber(context.Background(), nil)
	return header.Number, err
}

// GetTransaction generate a transaction for transfer
func (e *EthClient) GetTransaction(td *client.Transaction) ([]byte, []byte, error) {
	toAddr := common.HexToAddress(td.To)
	tx := e.generateTx(&toAddr, td)
	var hash common.Hash
	if e.SupportEIP1559 {
		signer := types.NewLondonSigner(e.chainID)
		hash = signer.Hash(tx)
	} else {
		signer := types.NewEIP155Signer(e.chainID)
		hash = signer.Hash(tx)
	}
	message, err := tx.MarshalBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("encode message failed, err=%s", err)
	}
	return message, hash.Bytes(), nil
}

// BroadcastTransaction will broad the signed transaction to chain
func (e *EthClient) BroadcastTransaction(trans []byte, signature []byte) ([]byte, error) {
	tx := &types.Transaction{}
	if err := tx.UnmarshalBinary(trans); err != nil {
		return nil, fmt.Errorf("parse transaction failed, err=%s", err)
	}
	if len(signature) != crypto.SignatureLength {
		return nil, fmt.Errorf("invalid signature length, expect=%d, got=%d", crypto.SignatureLength, len(signature))
	}
	signer := types.NewLondonSigner(e.chainID)
	signedTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, fmt.Errorf("combine with signature failed, err=%s", err)
	}
	hash := signedTx.Hash()
	// if debugClient is set then go debug logic only
	if e.debugClient != nil {
		return hash[:], e.debugClient.SendTransaction(context.Background(), signedTx)
	}
	return hash[:], e.client.SendTransaction(context.Background(), signedTx)
}

func (e *EthClient) CallContract(td *client.Transaction) ([]byte, error) {
	from := common.HexToAddress(td.From)
	to := common.HexToAddress(td.To)
	msg := ethereum.CallMsg{From: from, To: &to, Value: td.Amount, Data: td.Data}
	return e.client.CallContract(context.Background(), msg, nil)
}

// Helper functions these functions may different on different chains
func (e *EthClient) contractAddressOf(strABI, strBIN string, td *client.Transaction) (string, error) {
	address := common.HexToAddress(td.From)
	contractAddr := ecrypto.CreateAddress(address, td.Nonce)
	return contractAddr.Hex(), nil
}

func (e *EthClient) AbiConvertToInt(v interface{}) *big.Int {
	return *abi.ConvertType(v, new(*big.Int)).(**big.Int)
}

func (e *EthClient) AbiConvertToString(v interface{}) string {
	return *abi.ConvertType(v, new(string)).(*string)
}

func (e *EthClient) AbiConvertToBytes(v interface{}) []byte {
	value := abi.ConvertType(v, new([]byte)).(*[]byte)
	return *value
}

func (e *EthClient) AbiConvertToAddress(v interface{}) string {
	value := abi.ConvertType(v, new(common.Address)).(*common.Address)
	return value.Hex()
}

// GetTransactionByHash gets the transaction information from chain
func (e *EthClient) GetTransactionByHash(transactionHash string) (*client.TransactionInfo, error) {
	hash := common.HexToHash(transactionHash)
	tx, isPending, err := e.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("get transaction failed, hash=%s, err=%s", transactionHash, err)
	}
	txReceipt, err := e.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("get transaction receipt failed, hash=%s, err=%s", transactionHash, err)
	}
	info := client.TransactionInfo{}
	transaction := client.Transaction{}
	transaction.To = tx.To().Hex()
	transaction.Nonce = tx.Nonce()
	transaction.ChainID = tx.ChainId()
	transaction.Amount = tx.Value()
	transaction.Data = tx.Data()
	// gas information
	fee := client.FeeLimit{}
	fee.Gas = big.NewInt(0).SetUint64(tx.Gas())
	fee.GasFeeCap = tx.GasPrice()
	fee.GasTipCap = tx.GasTipCap()

	// gasPrice 优先从receipt中获取
	gasPrice := tx.GasPrice()
	if txReceipt.EffectiveGasPrice.Cmp(big.NewInt(0)) > 0 {
		gasPrice = txReceipt.EffectiveGasPrice
	}

	gasUsed := big.NewInt(0).SetUint64(txReceipt.GasUsed)
	info.Gas = &client.TxGasInfo{
		Fee:      big.NewInt(0).Mul(gasUsed, gasPrice),
		GasPrice: gasPrice,
		GasUsed:  gasUsed,
	}

	transaction.Fee = &fee
	info.Tx = &transaction
	info.IsPending = isPending
	if txReceipt.Status == 1 {
		info.Status = client.TransactionStatusSuccess
	} else {
		info.Status = client.TransactionStatusFailed
	}
	sender, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
	if err != nil {
		info.Status, info.Error = client.TransactionStatusInvalid, "get_from_failed"
		return nil, fmt.Errorf("get from failed, err=%s", err)
	}
	transaction.From = sender.String()
	events := make([]*client.EventLog, 0, len(txReceipt.Logs))
	for i := range txReceipt.Logs {
		event := client.EventLog{
			Address: txReceipt.Logs[i].Address.Hex(),
			Data:    txReceipt.Logs[i].Data,
			Removed: txReceipt.Logs[i].Removed,
			Topics:  make([][]byte, 0, len(txReceipt.Logs[i].Topics)),
		}
		for j := range txReceipt.Logs[i].Topics {
			topic := txReceipt.Logs[i].Topics[j]
			event.Topics = append(event.Topics, topic[:])
		}
		events = append(events, &event)
	}
	info.Logs = events
	if info.Status != client.TransactionStatusSuccess {
		_, err := e.getRevertReason(sender, tx, txReceipt)
		info.Error = err.Error()
	}
	return &info, nil
}

func (e *EthClient) getRevertReason(sender common.Address, tx *types.Transaction, reciept *types.Receipt) ([]byte, error) {
	msg := ethereum.CallMsg{
		From:     sender,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}
	return e.client.CallContract(context.Background(), msg, reciept.BlockNumber)
}

func (e *EthClient) ParseEventLog(abiName string, eventLog *client.EventLog) ([]interface{}, error) {
	compiled, err := e.GetABIByName(abiName)
	if err != nil {
		return nil, err
	}
	if len(eventLog.Topics) == 0 {
		return nil, fmt.Errorf("no topic found")
	}
	eventID := common.Hash{}
	eventID.SetBytes(eventLog.Topics[0])
	event, err := compiled.EventByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("get event from id failed, err=%s", err)
	}
	return event.Inputs.Unpack(eventLog.Data)
}

func (e *EthClient) AddressFromPrivateKey(privateKey string) (string, error) {
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}
	key, err := ecrypto.HexToECDSA(privateKey)
	if err != nil {
		return "", fmt.Errorf("wrong private key=%s, err=%s", privateKey, err)
	}
	return e.AddressFromPublicKey(&key.PublicKey)
}

func (e *EthClient) AddressFromPublicKey(pubKey *ecdsa.PublicKey) (string, error) {
	address := ecrypto.PubkeyToAddress(*pubKey)
	return address.Hex(), nil
}

func (e *EthClient) IsValidAddress(address string) bool {
	if address == nativeAsset {
		return false
	}
	return common.IsHexAddress(address)
}

func (e *EthClient) AddressFromString(addr string) (common.Address, error) {
	if !common.IsHexAddress(addr) {
		return common.Address{}, fmt.Errorf("invalid address=%s", addr)
	}
	return common.HexToAddress(addr), nil
}

func (e *EthClient) AddressToString(addr common.Address) string {
	return addr.Hex()
}

func (e *EthClient) ContractAddress(addr common.Address) (bool, error) {
	code, err := e.client.CodeAt(context.Background(), addr, nil)
	if err != nil {
		return false, fmt.Errorf("get code failed, err=%s", err)
	}
	return len(code) > 0, nil
}

func (e *EthClient) IsNativeAsset(asset string) bool {
	if e.chainID == big.NewInt(137) {
		return asset == maticNativeAsset || asset == nativeAsset
	}
	return asset == nativeAsset
}

func (e *EthClient) NormalizeAddress(address string) string {
	return common.HexToAddress(address).String()
}

// GetGasPrice  Deprecate!! 尽量使用 GetSuggestGasPrice 替代
func (e *EthClient) GetGasPrice() (*big.Int, *big.Int, error) {
	feeCap, err := e.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("get gas price failed, err=%s", err)
	}
	tipCap, err := e.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("get gas tip failed, err=%s", err)
	}
	return feeCap, tipCap, nil
}

// GetSuggestGasPrice 获取建议的gasPrice, tipCap 以及返回当前最新块的baseFee。
// todo:: 这里需要优化为并发请求2次接口
func (e *EthClient) GetSuggestGasPrice() (*big.Int, *big.Int, *big.Int, error) {
	//获取建议的gas
	gasPrice, err := e.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get gas price failed, err=%s", err)
	}

	//如果不支持EIP1559，直接返回
	if !e.SupportEIP1559 {
		return big.NewInt(0), big.NewInt(0), gasPrice, nil
	}

	//获取建议的tip
	tipCap, err := e.client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get gas tip failed, err=%s", err)
	}

	//根据建议推算当前的basePrice
	baseFee := &big.Int{}
	if gasPrice.Cmp(tipCap) > 0 {
		baseFee = big.NewInt(0).Sub(gasPrice, tipCap)
	}
	return baseFee, tipCap, gasPrice, nil
}

func (e *EthClient) NativeAssetAddress() string {
	return nativeAsset
}

func (e *EthClient) PublicKeyHexToAddress(key string) (string, error) {
	buffer, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("decode public key failed, err=%s", err)
	}
	pubKey, err := ecrypto.UnmarshalPubkey(buffer)
	if err != nil {
		return "", fmt.Errorf("unmarshal public key failed, err=%s", err)
	}
	addr := ecrypto.PubkeyToAddress(*pubKey)
	return addr.Hex(), nil
}

func (e *EthClient) GetLackedGas(address string, gas uint64, gasPrice *big.Int, txSize uint64) (*big.Int, error) {
	balance, err := e.BalanceAt(address)
	if err != nil {
		return big.NewInt(0), fmt.Errorf("get balance failed, err=%s", err)
	}
	need := new(big.Int).Mul(gasPrice, big.NewInt(int64(gas)))
	if balance.Cmp(need) >= 0 {
		return big.NewInt(0), nil
	}
	need = need.Sub(need, balance)
	return need, nil
}

func (e *EthClient) NativeAssetDecimals() uint8 {
	return 18
}
