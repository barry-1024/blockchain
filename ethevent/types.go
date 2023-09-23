package ethevent

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type IEventType interface {
	SetContractAddr(contractAddr string)
	SetEventName(eventName string)
	SetEventID(eventID string)

	GetContractAddr() string
	GetEventName() string
	GetEventID() string
}

const (
	EventNameTransfer       = "Transfer"
	EventNameApproval       = "Approval"
	EventNameApprovalForAll = "ApprovalForAll"
	EventNameUxUyTrade      = "Traded"
)

type Origin struct {
	contractAddr string
	eventName    string
	eventID      string
}

type Transfer struct {
	Origin
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value"`
}

type Approval struct {
	Origin
	Owner   common.Address `json:"owner"`
	Spender common.Address `json:"spender"`
	Value   *big.Int       `json:"value"`
}

type ApprovalForAll struct {
	Origin
	Owner    common.Address `json:"owner"`
	Operator common.Address `json:"operator"`
	Approved bool           `json:"approved"`
}

type UxUyTraded struct {
	Origin
	Sender            common.Address `json:"sender"`
	OrderId           *big.Int       `json:"orderId"`
	Recipient         common.Address `json:"recipient"`
	FeeShareRecipient common.Address `json:"feeShareRecipient"`
	TokenIn           common.Address `json:"tokenIn"`
	AmountIn          *big.Int       `json:"amountIn"`
	ChainIDOut        *big.Int       `json:"chainIDOut"`
	TokenOut          common.Address `json:"tokenOut"`
	AmountOut         *big.Int       `json:"amountOut"`
	BridgeTxnID       *big.Int       `json:"bridgeTxnID"`
	FeeToken          common.Address `json:"feeToken"`
	AmountFee         *big.Int       `json:"amountFee"`
	AmountFeeShare    *big.Int       `json:"amountFeeShare"`
	AmountExtraFee    *big.Int       `json:"amountExtraFee"`
}

func (e *Origin) SetContractAddr(contractAddr string) {
	e.contractAddr = contractAddr
}

func (e *Origin) SetEventName(eventName string) {
	e.eventName = eventName
}

func (e *Origin) SetEventID(eventID string) {
	e.eventID = eventID
}

func (e *Origin) GetContractAddr() string {
	return e.contractAddr
}

func (e *Origin) GetEventName() string {
	return e.eventName
}

func (e *Origin) GetEventID() string {
	return e.eventID
}
