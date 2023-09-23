package tron

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/stretchr/testify/assert"
)

func TestAddress(t *testing.T) {
	//data := "410000000000000000000000000000000000000000"
	addrByte, err := hex.DecodeString("f90f35f998e9aa74924f69c958ad9cd400a58f22")
	assert.Nil(t, err, "decode address error")
	addr := make([]byte, 0, 32)
	addr = append(addr, addressPrefix)
	addr = append(addr, addrByte...)
	addrValue := address.HexToAddress(hex.EncodeToString(addr))
	fmt.Println(addrValue.Hex(), addrValue.String())
}

func TestBase58Address(t *testing.T) {
	addr := "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
	value, err := address.Base58ToAddress(addr)
	assert.Nil(t, err, "error failed")
	fmt.Println(value.Hex(), addr)
}

func TestAddressConversion(t *testing.T) {
	addr := "TVnFbxVHgu5EgCocuSB4AwKVWyscPgAodE"
	client, err := NewTronClient(&tConfig)
	assert.Nil(t, err, "error failed")
	addrValue, _ := client.AddressFromString(addr)
	a := client.AddressToString(addrValue)
	assert.Equal(t, addr, a, "address not equal")
}
