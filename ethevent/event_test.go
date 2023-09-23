package ethevent

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"math/big"
	"reflect"
	"testing"
)

func TestTransfer(t *testing.T) {
	tf := &Transfer{}
	data, _ := hexutil.Decode("0x00000000000000000000000000000000000000000000000000000000000cd140")
	log := &EventLog{
		Address: common.HexToAddress("0xc2132d05d31c914a87c6611c10748aeb04b58e8f"),
		Topics: []common.Hash{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			common.HexToHash("0x000000000000000000000000e7804c37c13166ff0b37f5ae0bb07a3aebb6e245"),
			common.HexToHash("0x00000000000000000000000098116fc6ca32399d3835e24720255ae3c6172fa0"),
		},
		Data: data,
	}

	_, err := ParseEventToStruct(tf, log)
	assert.Nil(t, err, "transfer convert failed")
	if tf != nil {
		assert.Equal(t, "0xe7804c37c13166fF0b37F5aE0BB07A3aEbb6e245", tf.From.String())
		assert.Equal(t, "0x98116fC6Ca32399d3835e24720255Ae3C6172FA0", tf.To.String())
		assert.Equal(t, "840000", tf.Value.String())
	}
}

func TestTransferInputNil(t *testing.T) {
	data, _ := hexutil.Decode("0x00000000000000000000000000000000000000000000000000000000000cd140")
	log := &EventLog{
		Address: common.HexToAddress("0xc2132d05d31c914a87c6611c10748aeb04b58e8f"),
		Topics: []common.Hash{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			common.HexToHash("0x000000000000000000000000e7804c37c13166ff0b37f5ae0bb07a3aebb6e245"),
			common.HexToHash("0x00000000000000000000000098116fc6ca32399d3835e24720255ae3c6172fa0"),
		},
		Data: data,
	}

	tf, err := ParseEventToStruct(nil, log)
	assert.Equal(t, EventNameTransfer, tf.GetEventName(), "transfer convert failed")
	assert.Nil(t, err, "transfer convert failed")
	assert.IsType(t, &Transfer{}, tf, "transfer type parsed failed")

	if reflect.TypeOf(tf).String() == "*ethevent.Transfer" {
		tff := tf.(*Transfer)
		if tff != nil {
			fmt.Println("xxxx")
			assert.Equal(t, "0xe7804c37c13166fF0b37F5aE0BB07A3aEbb6e245", tff.From.String())
			assert.Equal(t, "0x98116fC6Ca32399d3835e24720255Ae3C6172FA0", tff.To.String())
			assert.Equal(t, "840000", tff.Value.String())
		}
	}
}

func TestApprove(t *testing.T) {
	//todo::
}

func TestUxUyTraded(t *testing.T) {
	tf := &UxUyTraded{}
	data, _ := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c2132d05d31c914a87c6611c10748aeb04b58e8f00000000000000000000000000000000000000000000000000000000000f4240000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001377d460e64b298e00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000eff1932c35f2100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	log := &EventLog{
		Address: common.HexToAddress("0xc2132d05d31c914a87c6611c10748aeb04b58e8f"),
		Topics: []common.Hash{
			common.HexToHash("0xfad3e93c5edf363a6a7db9e8df2679b14559c44e44a66a163359056050ab806f"),
			common.HexToHash("0x0000000000000000000000000b57341bca5a9f6e640b89fb19d06e36e48bcfad"),
			common.HexToHash("0x0000000000000000000000000b57341bca5a9f6e640b89fb19d06e36e48bcfad"),
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		},
		Data: data,
	}

	_, err := ParseEventToStruct(tf, log)
	assert.Nil(t, err, "traded convert failed")
	if tf != nil {
		assert.Equal(t, "0x0b57341BcA5a9F6e640b89FB19D06E36E48BcFad", tf.Sender.String())
		assert.Equal(t, "0x0b57341BcA5a9F6e640b89FB19D06E36E48BcFad", tf.Recipient.String())
		assert.Equal(t, "0xc2132D05D31c914a87C6611C10748AEb04B58e8F", tf.TokenIn.String())
		assert.Equal(t, big.NewInt(1000000), tf.AmountIn)
		t.Logf("tf: %+v", tf)
	}
}
