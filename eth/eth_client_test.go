package eth

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"git.bipal.space/shared-lib/blockchain/client"
	bclient "git.bipal.space/shared-lib/blockchain/client"
	"git.bipal.space/shared-lib/blockchain/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

var (
	config = client.ChainConfiguration{
		//Endpoint: "https://mainnet.infura.io/v3/648a408b225e4433aa29f0b22534f818",
		Endpoints: []string{"https://sepolia.infura.io/v3/3b85612ab2a5435e950233eee7abd23e"},
		//Endpoint:       "https://rpc.ankr.com/polygon/c671979afc23d0bac7a1c0e398d6cd0a5be2609766e3f208a14ea70600cc6be5",
		ChainID:        big.NewInt(1337),
		SupportEIP1559: true,
	}

	testConfig = bclient.ChainConfiguration{
		//Endpoints: []string{"https://goerli.infura.io/v3/648a408b225e4433aa29f0b22534f818"},
		Endpoints: []string{"https://sepolia.infura.io/v3/3b85612ab2a5435e950233eee7abd23e"},
		//Endpoint:       "https://arb-mainnet.g.alchemy.com/v2/2sKrO8oGFQEeB9iB468lIMHiHu7bYUiU",
		ChainID:        big.NewInt(5),
		SupportEIP1559: true,
	}

	bscConfig = bclient.ChainConfiguration{
		Endpoints:      []string{"https://bsc-dataseed4.ninicoin.io"},
		ChainID:        big.NewInt(56),
		SupportEIP1559: false,
	}
)

const (
	strABI        = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"subtractedValue\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"addedValue\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"transferOwner\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
	strBIN        = "0x60806040523480156200001157600080fd5b506040518060400160405280600a81526020017f4d61676963546f6b656e000000000000000000000000000000000000000000008152506040518060400160405280600381526020017f4d4754000000000000000000000000000000000000000000000000000000000081525081600390816200008f919062000365565b508060049081620000a1919062000365565b50505033600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506200044c565b600081519050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b600060028204905060018216806200016d57607f821691505b60208210810362000183576200018262000125565b5b50919050565b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b600060088302620001ed7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82620001ae565b620001f98683620001ae565b95508019841693508086168417925050509392505050565b6000819050919050565b6000819050919050565b600062000246620002406200023a8462000211565b6200021b565b62000211565b9050919050565b6000819050919050565b620002628362000225565b6200027a62000271826200024d565b848454620001bb565b825550505050565b600090565b6200029162000282565b6200029e81848462000257565b505050565b5b81811015620002c657620002ba60008262000287565b600181019050620002a4565b5050565b601f8211156200031557620002df8162000189565b620002ea846200019e565b81016020851015620002fa578190505b6200031262000309856200019e565b830182620002a3565b50505b505050565b600082821c905092915050565b60006200033a600019846008026200031a565b1980831691505092915050565b600062000355838362000327565b9150826002028217905092915050565b6200037082620000eb565b67ffffffffffffffff8111156200038c576200038b620000f6565b5b62000398825462000154565b620003a5828285620002ca565b600060209050601f831160018114620003dd5760008415620003c8578287015190505b620003d4858262000347565b86555062000444565b601f198416620003ed8662000189565b60005b828110156200041757848901518255600182019150602085019450602081019050620003f0565b8683101562000437578489015162000433601f89168262000327565b8355505b6001600288020188555050505b505050505050565b611662806200045c6000396000f3fe608060405234801561001057600080fd5b50600436106100f55760003560e01c80634fb2e45d11610097578063a0712d6811610066578063a0712d6814610288578063a457c2d7146102a4578063a9059cbb146102d4578063dd62ed3e14610304576100f5565b80634fb2e45d1461020057806370a082311461021c5780638da5cb5b1461024c57806395d89b411461026a576100f5565b806323b872dd116100d357806323b872dd14610166578063313ce5671461019657806339509351146101b457806342966c68146101e4576100f5565b806306fdde03146100fa578063095ea7b31461011857806318160ddd14610148575b600080fd5b610102610334565b60405161010f9190610e7c565b60405180910390f35b610132600480360381019061012d9190610f37565b6103c6565b60405161013f9190610f92565b60405180910390f35b6101506103e9565b60405161015d9190610fbc565b60405180910390f35b610180600480360381019061017b9190610fd7565b6103f3565b60405161018d9190610f92565b60405180910390f35b61019e610422565b6040516101ab9190611046565b60405180910390f35b6101ce60048036038101906101c99190610f37565b61042b565b6040516101db9190610f92565b60405180910390f35b6101fe60048036038101906101f99190611061565b610462565b005b61021a6004803603810190610215919061108e565b610471565b005b6102366004803603810190610231919061108e565b61050f565b6040516102439190610fbc565b60405180910390f35b610254610557565b60405161026191906110ca565b60405180910390f35b61027261057d565b60405161027f9190610e7c565b60405180910390f35b6102a2600480360381019061029d9190611061565b61060f565b005b6102be60048036038101906102b99190610f37565b610698565b6040516102cb9190610f92565b60405180910390f35b6102ee60048036038101906102e99190610f37565b61070f565b6040516102fb9190610f92565b60405180910390f35b61031e600480360381019061031991906110e5565b610732565b60405161032b9190610fbc565b60405180910390f35b60606003805461034390611154565b80601f016020809104026020016040519081016040528092919081815260200182805461036f90611154565b80156103bc5780601f10610391576101008083540402835291602001916103bc565b820191906000526020600020905b81548152906001019060200180831161039f57829003601f168201915b5050505050905090565b6000806103d16107b9565b90506103de8185856107c1565b600191505092915050565b6000600254905090565b6000806103fe6107b9565b905061040b85828561098a565b610416858585610a16565b60019150509392505050565b60006012905090565b6000806104366107b9565b90506104578185856104488589610732565b61045291906111b4565b6107c1565b600191505092915050565b61046e33600083610a16565b50565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104cb57600080fd5b80600560006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b60008060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60606004805461058c90611154565b80601f01602080910402602001604051908101604052809291908181526020018280546105b890611154565b80156106055780601f106105da57610100808354040283529160200191610605565b820191906000526020600020905b8154815290600101906020018083116105e857829003601f168201915b5050505050905090565b600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461066957600080fd5b610695600560009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1682610c8c565b50565b6000806106a36107b9565b905060006106b18286610732565b9050838110156106f6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106ed9061125a565b60405180910390fd5b61070382868684036107c1565b60019250505092915050565b60008061071a6107b9565b9050610727818585610a16565b600191505092915050565b6000600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905092915050565b600033905090565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610830576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610827906112ec565b60405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361089f576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108969061137e565b60405180910390fd5b80600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9258360405161097d9190610fbc565b60405180910390a3505050565b60006109968484610732565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8114610a105781811015610a02576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109f9906113ea565b60405180910390fd5b610a0f84848484036107c1565b5b50505050565b600073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603610a85576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a7c9061147c565b60405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610af4576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610aeb9061150e565b60405180910390fd5b610aff838383610de2565b60008060008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905081811015610b85576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b7c906115a0565b60405180910390fd5b8181036000808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550816000808573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825401925050819055508273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef84604051610c739190610fbc565b60405180910390a3610c86848484610de7565b50505050565b600073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603610cfb576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610cf29061160c565b60405180910390fd5b610d0760008383610de2565b8060026000828254610d1991906111b4565b92505081905550806000808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825401925050819055508173ffffffffffffffffffffffffffffffffffffffff16600073ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051610dca9190610fbc565b60405180910390a3610dde60008383610de7565b5050565b505050565b505050565b600081519050919050565b600082825260208201905092915050565b60005b83811015610e26578082015181840152602081019050610e0b565b60008484015250505050565b6000601f19601f8301169050919050565b6000610e4e82610dec565b610e588185610df7565b9350610e68818560208601610e08565b610e7181610e32565b840191505092915050565b60006020820190508181036000830152610e968184610e43565b905092915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610ece82610ea3565b9050919050565b610ede81610ec3565b8114610ee957600080fd5b50565b600081359050610efb81610ed5565b92915050565b6000819050919050565b610f1481610f01565b8114610f1f57600080fd5b50565b600081359050610f3181610f0b565b92915050565b60008060408385031215610f4e57610f4d610e9e565b5b6000610f5c85828601610eec565b9250506020610f6d85828601610f22565b9150509250929050565b60008115159050919050565b610f8c81610f77565b82525050565b6000602082019050610fa76000830184610f83565b92915050565b610fb681610f01565b82525050565b6000602082019050610fd16000830184610fad565b92915050565b600080600060608486031215610ff057610fef610e9e565b5b6000610ffe86828701610eec565b935050602061100f86828701610eec565b925050604061102086828701610f22565b9150509250925092565b600060ff82169050919050565b6110408161102a565b82525050565b600060208201905061105b6000830184611037565b92915050565b60006020828403121561107757611076610e9e565b5b600061108584828501610f22565b91505092915050565b6000602082840312156110a4576110a3610e9e565b5b60006110b284828501610eec565b91505092915050565b6110c481610ec3565b82525050565b60006020820190506110df60008301846110bb565b92915050565b600080604083850312156110fc576110fb610e9e565b5b600061110a85828601610eec565b925050602061111b85828601610eec565b9150509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b6000600282049050600182168061116c57607f821691505b60208210810361117f5761117e611125565b5b50919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006111bf82610f01565b91506111ca83610f01565b92508282019050808211156111e2576111e1611185565b5b92915050565b7f45524332303a2064656372656173656420616c6c6f77616e63652062656c6f7760008201527f207a65726f000000000000000000000000000000000000000000000000000000602082015250565b6000611244602583610df7565b915061124f826111e8565b604082019050919050565b6000602082019050818103600083015261127381611237565b9050919050565b7f45524332303a20617070726f76652066726f6d20746865207a65726f2061646460008201527f7265737300000000000000000000000000000000000000000000000000000000602082015250565b60006112d6602483610df7565b91506112e18261127a565b604082019050919050565b60006020820190508181036000830152611305816112c9565b9050919050565b7f45524332303a20617070726f766520746f20746865207a65726f20616464726560008201527f7373000000000000000000000000000000000000000000000000000000000000602082015250565b6000611368602283610df7565b91506113738261130c565b604082019050919050565b600060208201905081810360008301526113978161135b565b9050919050565b7f45524332303a20696e73756666696369656e7420616c6c6f77616e6365000000600082015250565b60006113d4601d83610df7565b91506113df8261139e565b602082019050919050565b60006020820190508181036000830152611403816113c7565b9050919050565b7f45524332303a207472616e736665722066726f6d20746865207a65726f20616460008201527f6472657373000000000000000000000000000000000000000000000000000000602082015250565b6000611466602583610df7565b91506114718261140a565b604082019050919050565b6000602082019050818103600083015261149581611459565b9050919050565b7f45524332303a207472616e7366657220746f20746865207a65726f206164647260008201527f6573730000000000000000000000000000000000000000000000000000000000602082015250565b60006114f8602383610df7565b91506115038261149c565b604082019050919050565b60006020820190508181036000830152611527816114eb565b9050919050565b7f45524332303a207472616e7366657220616d6f756e742065786365656473206260008201527f616c616e63650000000000000000000000000000000000000000000000000000602082015250565b600061158a602683610df7565b91506115958261152e565b604082019050919050565b600060208201905081810360008301526115b98161157d565b9050919050565b7f45524332303a206d696e7420746f20746865207a65726f206164647265737300600082015250565b60006115f6601f83610df7565b9150611601826115c0565b602082019050919050565b60006020820190508181036000830152611625816115e9565b905091905056fea2646970667358221220e1c4dd16a135635248f819d884c36ddc4c035310a1fa1c9d081d35b20618e6c964736f6c63430008110033"
	strUniswapABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenIn\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"tokenOut\",\"type\":\"address\"},{\"internalType\":\"uint24\",\"name\":\"fee\",\"type\":\"uint24\"},{\"internalType\":\"uint256\",\"name\":\"amountIn\",\"type\":\"uint256\"},{\"internalType\":\"uint160\",\"name\":\"sqrtPriceLimitX96\",\"type\":\"uint160\"}],\"name\":\"quoteExactInputSingle\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amountOut\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
)

var UXUXBridgeABI = `[{"inputs":[{"internalType":"address[]","name":"acceptedTokens","type":"address[]"},{"internalType":"address[]","name":"uagents","type":"address[]"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"bool","name":"accepted","type":"bool"}],"name":"AcceptedTokenChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"caller","type":"address"},{"indexed":false,"internalType":"bool","name":"allowed","type":"bool"}],"name":"AllowedCallerChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"uagent","type":"address"},{"indexed":false,"internalType":"address","name":"outAddress","type":"address"},{"indexed":false,"internalType":"address","name":"outToken","type":"address"},{"indexed":false,"internalType":"uint256","name":"outChainId","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"outMinAmount","type":"uint256"},{"indexed":false,"internalType":"address","name":"inToken","type":"address"},{"indexed":false,"internalType":"uint256","name":"inAmount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"orderId","type":"uint256"}],"name":"Transferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"},{"indexed":false,"internalType":"bool","name":"uagent","type":"bool"}],"name":"UAgentChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"token","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"Withdrawn","type":"event"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function","constant":true},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"caller","type":"address"},{"internalType":"bool","name":"allowed","type":"bool"}],"name":"updateAllowedCaller","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"address","name":"recipient","type":"address"}],"name":"withdrawNativeAsset","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"address","name":"recipient","type":"address"}],"name":"withdrawToken","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive","payable":true},{"inputs":[{"internalType":"address[]","name":"tokens","type":"address[]"},{"internalType":"bool","name":"accepted","type":"bool"}],"name":"updateAcceptedTokens","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address[]","name":"accounts","type":"address[]"},{"internalType":"bool","name":"uagent","type":"bool"}],"name":"updateUAgents","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"supportSwap","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"pure","type":"function","constant":true},{"inputs":[{"components":[{"internalType":"address","name":"tokenIn","type":"address"},{"internalType":"uint256","name":"chainIDOut","type":"uint256"},{"internalType":"address","name":"tokenOut","type":"address"},{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint256","name":"minAmountOut","type":"uint256"},{"internalType":"address","name":"recipient","type":"address"},{"internalType":"bytes","name":"data","type":"bytes"}],"internalType":"struct IBridgeAdapter.BridgeParams","name":"params","type":"tuple"}],"name":"bridge","outputs":[{"internalType":"uint256","name":"","type":"uint256"},{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"payable","type":"function","payable":true}]`
var UPoolUSDTAddress = map[uint64]map[string]interface{}{
	137: {
		"contract":        "0xc2132d05d31c914a87c6611c10748aeb04b58e8f",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87",
		"bridge_contract": "0x9774ef62cf1f5985f1f2ee21d1e8c631a470f626",
		"fee":             "0.0003",
	},
	56: {
		"contract":        "0x55d398326f99059ff775485246999027b3197955",
		"decimals":        18,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "0xd4fdD4588b4202411548C3DbDD5f4Fb763bC2E9E",                       // need to check
		"fee":             "0.0003",
	},
	1: {
		"contract":        "0xdac17f958d2ee523a2206206994597c13d831ec7",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "0xec7Db1A2e39b532b84f2f5549330eb1485953743",                       // need to check
		"fee":             "0.0003",
	},
	250: {
		"contract":        "0x049d68029688eAbF473097a2fC38ef61633A3C7A",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "0xc8e49A5e0447b99C13aad79542812858D4220b24",                       // need to check
		"fee":             "0.0003",
	},
	42161: {
		"contract":        "0xff970a61a04b1ca14834a43f5de4533ebddb5cc8",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "0xEEb24183819F5c36475736e4815d172E826D197b",                       // need to check
		"fee":             "0.0003",
	},
	43114: {
		"contract":        "0xc7198437980c041c805a1edcba50c1ce5db95118",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "0x905F0f32007dfD9833aeA796d22D181b206a5dB6",                       // need to check
		"fee":             "0.0003",
	},
	10: {
		"contract":        "0x94b008aA00579c1307B0EF2c499aD98a8ce58e58",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "0xc4f6d317163eF2A0C7a9Da1eACE3Eb73eA7C4Af8",                       // need to check
		"fee":             "0.0003",
	},
	1000001: {
		"contract":        "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		"decimals":        6,
		"coin":            "USDT",
		"transfer_topic":  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
		"bridge_topic":    "76272a4bde5e5589ca452fd617567674ea2fcedd0572b269554b55002d158c87", // need to check
		"bridge_contract": "TDuTc9J2RpD7M3vaXtSUuvxHFBQ9BRtsLe",                               // need to check
		"fee":             "0.0003",
	},
}

func TestBalanceOf(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	// address of USDC
	erc20Address := "0xCA1d7dE02439eec7727AeE15cD8bF36cCD9728c7"
	// from is magic's address
	from := "0xC8bD5B1aD2FD42Ef9D92B32F38E9b0DFAC875Be4"
	value, err := client.BalanceOf(erc20Address, from)
	assert.Nil(t, err, "balance failed")
	assert.Equal(t, value.Cmp(big.NewInt(100000)), 0, "value should equal")
}

func TestBalanceAt(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	from := "0xC8bD5B1aD2FD42Ef9D92B32F38E9b0DFAC875Be4"
	value, err := client.BalanceAt(from)
	fmt.Println("BalanceAt ", value)
	assert.Nil(t, err, "balance failed")
	assert.True(t, value.Cmp(big.NewInt(0)) > 0, "value should bigger than 0")
}

func TestDecimalsOf(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	usdcAddress := "0x3506424F91fD33084466F402d5D97f05F8e3b4AF"
	decimals, err := client.DecimalsOf(usdcAddress)
	assert.Nil(t, err, "get decimals should success")
	assert.Equal(t, decimals, uint8(18), "decimals should be 6")
}

func TestTotalSupplyOf(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	erc20Address := "0xCA1d7dE02439eec7727AeE15cD8bF36cCD9728c7"
	supply, err := client.TotalSupplyOf(erc20Address)
	fmt.Println("erc20 ", erc20Address)
	assert.Nil(t, err, "get total supply should success")
	assert.True(t, supply.Cmp(big.NewInt(1000000)) > 0, "value should be big")
}

func TestSymbolOf(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	usdcAddress := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	symbol, err := client.SymbolOf(usdcAddress)
	assert.Nil(t, err, "get total supply should success")
	assert.Equal(t, symbol, "USDC", "value should be big")
}

func TestGetNonde(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	from := "0xC8bD5B1aD2FD42Ef9D92B32F38E9b0DFAC875Be4"
	nonce, err := client.GetNonce(from)
	assert.Nil(t, err, "get nonce failed")
	assert.Greater(t, nonce, uint64(1), "nonce should > 1")
}

func TestGetTransaction(t *testing.T) {
	priKey, addr, cli := simulateClient()
	client, err := NewEthClient(&config)
	client.SetClient(cli)
	assert.Nil(t, err, "create client failed")
	td := bclient.Transaction{}
	td.ChainID = new(big.Int).SetUint64(1)
	//td.Contract = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	td.Method = "transferFrom"
	td.From = addr.Hex()
	td.To = "0x715d2B5aD8821BCabDE74EcEea85eA0296328Cb5"
	td.Amount = new(big.Int).SetUint64(10)
	td.Nonce = 0
	td.Fee = &bclient.FeeLimit{}
	td.Fee.GasFeeCap = big.NewInt(875000000)
	td.Fee.Gas = new(big.Int).SetUint64(210000)
	td.Fee.GasTipCap = new(big.Int).SetUint64(0)
	message, hash, err := client.GetTransaction(&td)
	assert.Nil(t, err, "get transaction failed")
	tx := types.Transaction{}
	assert.Nil(t, tx.UnmarshalBinary(message), "parse binary failed")
	signer := types.NewLondonSigner(config.ChainID)
	h := signer.Hash(&tx)
	assert.Equal(t, bytes.Compare(h[:], hash[:]), 0, "should equal")

	sig, err := crypto.Sign(hash, priKey)
	assert.Nil(t, err, "signature failed")
	_, err = client.BroadcastTransaction(message, sig)
	assert.Nil(t, err, "broadcast faled")
}

func simulateClient() (*ecdsa.PrivateKey, common.Address, *backends.SimulatedBackend) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, common.Address{}, nil
	}
	auth := bind.NewKeyedTransactor(privateKey)
	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei

	address := auth.From
	genesisAlloc := map[common.Address]core.GenesisAccount{
		address: {
			Balance: balance,
		},
	}
	blockGasLimit := uint64(4712388)
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	return privateKey, addr, backends.NewSimulatedBackend(genesisAlloc, blockGasLimit)
}
func TestDeployContract(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	priKey, from, cli := simulateClient()
	client.SetClient(cli)
	defer cli.Close()
	td := bclient.Transaction{}
	td.ChainID = new(big.Int).SetUint64(1)
	td.From = from.Hex()
	td.To = ""
	td.Amount = big.NewInt(0)
	td.Nonce = 0
	td.Fee = &bclient.FeeLimit{}
	td.Fee.GasFeeCap = new(big.Int).SetUint64(875000001)
	td.Fee.Gas = new(big.Int).SetUint64(210000)
	td.Fee.GasTipCap = new(big.Int).SetUint64(0)
	message, hash, _, err := client.DeployContract(strABI, strBIN, &td)
	sig, err := crypto.Sign(hash, priKey)
	assert.Nil(t, err, "signature failed")
	_, err = client.BroadcastTransaction(message, sig)
	assert.Nil(t, err, "broadcast failed")
	//fmt.Println(hex.EncodeToString(txID))
}

func TestCallContract(t *testing.T) {
	client, err := NewEthClient(&config)
	assert.Nil(t, err, "create client failed")
	tokenIn := common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
	tokenOut := common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	amountIn, _ := new(big.Int).SetString("10000000000000000000000", 10)
	data, err := client.GetTransactionData("quoteExactInputSingle", strUniswapABI, tokenIn, tokenOut, big.NewInt(3000), amountIn, big.NewInt(0))
	assert.Nil(t, err, "data failed")
	td := bclient.Transaction{}
	td.ChainID = new(big.Int).SetUint64(1)
	td.To = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"
	td.Data = data
	res, err := client.CallContract(&td)
	assert.Nil(t, err, "call contract failed")
	assert.NotEqual(t, bytes.Compare(res, []byte("00000000000000000000000000000000")), 0, "should not be empty")
}

func TestAbiPack(t *testing.T) {
	u256, err := abi.NewType("uint256", "", nil)
	assert.Nil(t, err, "wrong type")
	args := abi.Arguments{
		{Type: u256},
	}
	data, err := args.Pack(big.NewInt(128))
	assert.Nil(t, err, "wrong pack")
	assert.Equal(t, hex.EncodeToString(data), "0000000000000000000000000000000000000000000000000000000000000080", "wrong packed data")
}

func TestGetOnChainTransaction(t *testing.T) {
	client, _ := NewEthClient(&testConfig)
	//info, err := client.GetTransactionByHash("0x292de558a3490160f106c12ca5e98164fa78c8533f11ef6984bca8b4248a8b85")
	info, err := client.GetTransactionByHash("0xfda97728d22c89bb23c58a051ae8278beeeb7c6cadc324f866c28435fda2b245")
	fmt.Println("---", info.Logs)
	for _, log := range info.Logs {
		fmt.Println("log.Address", log.Address)
		fmt.Println("log.Address", log.Topics)
		bridgeAbi, err := abi.JSON(strings.NewReader(UXUXBridgeABI))
		if err != nil {
			fmt.Println("---bridgeAbi err", err)
		}
		fmt.Println("--- len", len(log.Data))
		fmt.Println("--- log.Topics", log.Topics)
		fmt.Println("bridge_topic:", strings.ToLower(UPoolUSDTAddress[137]["transfer_topic"].(string)))
		fmt.Println("bridge_contract:", strings.ToLower(UPoolUSDTAddress[137]["bridge_contract"].(string)))
		for _, topic := range log.Topics {
			fmt.Println("topic:", hex.EncodeToString(topic))
		}

		bridgeLogs, err := bridgeAbi.Unpack("Transferred", log.Data)
		if err != nil {
			fmt.Println("Unpack err", err)
		}
		fmt.Println("---bridgeLogs", bridgeLogs)

	}
	assert.Nil(t, err, "get transaction error")
	//assert.False(t, info.IsPending, "info pending status is wrong")
	//assert.Equal(t, uint64(1), info.Status, "transaction status wrong")
	//content, err := ioutil.ReadFile("./protocol2.abi")
	//assert.Nil(t, err, "load file failed")
	//abiName := "protocol"
	//assert.Nil(t, client.RegisterABI(abiName, string(content)), "register abi failed")
	//for i := range info.Logs {
	//	el := info.Logs[i]
	//	if el.Address == info.Tx.To {
	//		fields, err := client.ParseEventLog(abiName, el)
	//		assert.Nil(t, err, "failed to parse event")
	//		tokenIn := client.AbiConvertToAddress(fields[0])
	//		assert.Equal(t, tokenIn, "0x0000000000000000000000000000000000000000", "tokenIn not match")
	//	}
	//}
}

func TestTransferEvent(t *testing.T) {
	client, _ := NewEthClient(&config)
	info, err := client.GetTransactionByHash("0xfda97728d22c89bb23c58a051ae8278beeeb7c6cadc324f866c28435fda2b245")
	assert.Nil(t, err, "get transaction error")
	assert.Greater(t, len(info.Logs), 0, "no logs found")
	assert.Equal(t, len(info.Logs[0].Topics), 3, "wrong topic length")
	addrType, _ := abi.NewType("address", "", nil)
	a := abi.Argument{Type: addrType}
	addr := abi.Arguments{a}
	v, err := addr.Unpack((info.Logs[0].Topics[1]))
	assert.Nil(t, err, "unpack failed")
	from := client.AbiConvertToAddress(v[0])
	assert.Equal(t, from, "0xC8bD5B1aD2FD42Ef9D92B32F38E9b0DFAC875Be4")

	intType, _ := abi.NewType("uint256", "", nil)
	intArg := abi.Arguments{abi.Argument{Type: intType}}
	val, err := intArg.Unpack(info.Logs[0].Data)
	amount := client.AbiConvertToInt(val[0])
	fmt.Println(amount.String())
}

func TestNoramlizeAddress(t *testing.T) {
	client, _ := NewEthClient(&config)
	addr := client.NormalizeAddress("0x0000000022D53366457F9d5e68ec105046fc4383")
	assert.Equal(t, addr, "0x0000000022D53366457F9d5E68Ec105046FC4383", "normalize address failed")
}

func TestGetGasPrice(t *testing.T) {
	client, _ := NewEthClient(&testConfig)
	base, tip, err := client.GetGasPrice()
	assert.Nil(t, err, "get gas price failed")
	fmt.Println(base, tip)
}

func TestAddressFromPrivateKey(t *testing.T) {
	client, _ := NewEthClient(&config)
	address, err := client.AddressFromPrivateKey(utils.EthPrivateKey2)
	assert.Nil(t, err, "should success")
	assert.Equal(t, address, "0xa70fdFd8a32b6c0f32e246B53Fa45B3B372A73D8", "address not match")

	address, err = client.AddressFromPrivateKey("0x" + utils.EthPrivateKey2)
	assert.Nil(t, err, "should success")
	assert.Equal(t, address, "0xa70fdFd8a32b6c0f32e246B53Fa45B3B372A73D8", "address not match")
}

func TestTransferData(t *testing.T) {
	client, _ := NewEthClient(&config)
	data, err := client.TransferData("0xa70fdFd8a32b6c0f32e246B53Fa45B3B372A73D8", big.NewInt(1000000000000000000))
	assert.Nil(t, err, "should success")
	assert.Greater(t, len(data), 0, "data should not be empty")
}

func TestAllowance(t *testing.T) {
	client, _ := NewEthClient(&config)
	amount, err := client.Allowance("0xdAC17F958D2ee523a2206206994597C13D831ec7", "0xf1D7BEe92F49EAfc36b09b9953C05a2F4673cB40",
		"0x90fcDAE23d01e916b5FF0ce36CA9E4887DAEBDb5")
	assert.Nil(t, err, "should success")
	assert.Equal(t, amount.String(), "0", "amount not match")
}

func TestIsValidAddress(t *testing.T) {
	client, _ := NewEthClient(&config)
	valid := client.IsValidAddress("0xa70fdFd8a32b6c0f32e246B53Fa45B3B372A73D8")
	assert.True(t, valid, "address should be valid")

	valid = client.IsValidAddress("")
	assert.False(t, valid, "address should be invalid")

	valid = client.IsValidAddress("0x0000000000000000000000000000000000000000")
	assert.False(t, valid, "address should be valid")

	valid = client.IsValidAddress("0x000000000000000000000000000000000000000000")
	assert.False(t, valid, "address should be valid")
}

// TestUxuyCall is used for debuging contracts, can leave this function commented
func TestUxuyCall(t *testing.T) {
	client, _ := NewEthClient(&testConfig)
	//data, err := hex.DecodeString("893419ca00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000002a13907000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002a0000000000000000000000000cef4a0531b24319d7df091967b94c6b33716b1670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001ccbf0000000000000000000000000000000000000000000000000000000000000380000000000000000000000000000000000000000000000000000000006431b1ea00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020aa443a48000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000005338640ff4849900000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000002000000000000000000000000ff970a61a04b1ca14834a43f5de4533ebddb5cc8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000001f40000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000020aa443a4800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000003691d6afc00000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000002000000000000000000000000ff970a61a04b1ca14834a43f5de4533ebddb5cc8000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000001f4")
	data, err := hex.DecodeString("9fbf10fc000000000000000000000000000000000000000000000000000000000000279400000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000f1d7bee92f49eafc36b09b9953c05a2f4673cb400000000000000000000000000000000000000000000000000000000001312d0000000000000000000000000000000000000000000000000000000000013042a0000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000001e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000014f1d7bee92f49eafc36b09b9953c05a2f4673cb400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	assert.Nil(t, err, "failed")
	td := bclient.Transaction{}
	td.ChainID = new(big.Int).SetUint64(137)
	td.From = "0xf1D7BEe92F49EAfc36b09b9953C05a2F4673cB40"
	td.To = "0x7612ae2a34e5a363e137de748801fb4c86499152"
	//td.To = "0xce9bAe29Ed0dECb3e078289d99D15Ad80f5728c6"
	//td.To = "0xce9bAe29Ed0dECb3e078289d99D15Ad80f5728c6"
	//td.To = "0x60A4030c3F3882fd5e786195F4EEB56F98ACD6C8"
	td.Data = data
	td.Amount, _ = new(big.Int).SetString("20000000000000000", 10)
	result, err := client.CallContract(&td)
	fmt.Println(result, err)
}

func TestBSC(t *testing.T) {
	client, err := NewEthClient(&bscConfig)
	assert.Nil(t, err, "failed")
	data, err := client.ApproveData("0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3", "0x3ea040d8c646A3BF91914121f6e9594b172d6BaF", "0xf1D7BEe92F49EAfc36b09b9953C05a2F4673cB40", big.NewInt(1000000000000000000))
	assert.Nil(t, err, "failed")
	nonce, err := client.GetNonce("0x3ea040d8c646A3BF91914121f6e9594b172d6BaF")
	assert.Nil(t, err, "failed")
	td := bclient.Transaction{
		From:   "0x3ea040d8c646A3BF91914121f6e9594b172d6BaF",
		To:     "0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3",
		Data:   data,
		Nonce:  nonce,
		Amount: big.NewInt(0),
	}
	fee, err := client.GetSuggestFee(&td)
	assert.Nil(t, err, "failed")
	td.Fee = fee
	tx, hash, err := client.GetTransaction(&td)
	assert.Nil(t, err, "failed")
	pk := utils.EthPrivateKey
	privateKey, err := crypto.HexToECDSA(pk)
	assert.Nil(t, err, "failed")
	signature, err := crypto.Sign(hash, privateKey)
	assert.Nil(t, err, "failed")
	fmt.Println(hex.EncodeToString(signature))
	//signature[64] = 1
	result, err := client.BroadcastTransaction(tx, signature)
	fmt.Println(hex.EncodeToString(result), err)
}

func TestPublicKeyHexToAddress(t *testing.T) {
	key := "042f648f8f37f0a108cf4df48a094b4c01d322374a2bb4afbb1afa594280e69e073991ba0aeb1d1a2317088ee14dcf181edd9d46705015aaff0fa2ec366d48cb5a"
	client, err := NewEthClient(&bscConfig)
	assert.Nil(t, err, "failed")
	address, err := client.PublicKeyHexToAddress(key)
	assert.Nil(t, err, "failed")
	assert.Equal(t, address, "0xF2c1105fb02A1acC3C25EE1AeDb46639BC424857", "address failed")
}

const stargateFeeLibraryABI = `[{
	"inputs": [
		{"internalType": "uint256","name": "_srcPoolId","type": "uint256"},
		{"internalType": "uint256","name": "_dstPoolId","type": "uint256"},
		{"internalType": "uint16","name": "_dstChainId","type": "uint16"},
		{"internalType": "address","name": "_from","type": "address"},
		{"internalType": "uint256","name": "_amountSD","type": "uint256"}
	],
	"name": "getFees",
	"outputs": [
		{
		"components": [
			{"internalType": "uint256","name": "amount","type": "uint256"},
			{"internalType": "uint256","name": "eqFee","type": "uint256"},
			{"internalType": "uint256","name": "eqReward","type": "uint256"},
			{"internalType": "uint256","name": "lpFee","type": "uint256"},
			{"internalType": "uint256","name": "protocolFee","type": "uint256"},
			{"internalType": "uint256","name": "lkbRemove","type": "uint256"}
		],
		"internalType": "struct Pool.SwapObj","name": "s","type": "tuple"
		}
	],
	"stateMutability": "view",
	"type": "function"
	}]`

func TestStargateFee(t *testing.T) {
	c, err := NewEthClient(&testConfig)
	assert.Nil(t, err, "create client failed")
	err = c.RegisterABI("stargateFee", stargateFeeLibraryABI)
	assert.Nil(t, err, "register abi failed")
	srcPoolId := big.NewInt(1)
	dstPoolId := srcPoolId
	chainId := uint32(11155111)
	amount := big.NewInt(3000000000000)
	from := "0xca1d7de02439eec7727aee15cd8bf36ccd9728c7"
	data, err := c.GetTransactionDataByABI("getFees", "stargateFee", srcPoolId, dstPoolId, chainId, common.HexToAddress(from), amount)
	assert.Nil(t, err, "get data failed")
	td := client.Transaction{
		From:    from,
		To:      "0xC8bD5B1aD2FD42Ef9D92B32F38E9b0DFAC875Be4",
		Data:    data,
		Amount:  big.NewInt(0),
		ChainID: big.NewInt(1),
	}
	result, err := c.CallContract(&td)
	assert.Nil(t, err, "call contract failed")
	r, err := c.UnpackByABI("getFees", "stargateFee", result)
	fmt.Printf("----r: %#v\n", r)
	if r == nil {
		return
	}
	ri := reflect.ValueOf(r[0])
	for i := 0; i < ri.NumField(); i++ {
		field := ri.Field(i)
		elem := field.Elem()
		for j := 0; j < elem.NumField(); j++ {
			elemField := elem.Field(j)
			if elemField.Kind() == reflect.Slice {
				for k := 0; k < elemField.Len(); k++ {
					v := uint64(elemField.Index(k).Uint())
					fmt.Println(v)
				}
			}
		}
	}
}

func TestParseCalldata(t *testing.T) {
	c, _ := NewEthClient(&config)
	data, err := c.GetTransactionDataByABI("transfer", erc20ABIName, common.HexToAddress("0xf1D7BEe92F49EAfc36b09b9953C05a2F4673cB40"), big.NewInt(1000))
	assert.Nil(t, err, "failed to get data")
	contractABI, err := c.GetABIByName(erc20ABIName)
	assert.Nil(t, err, "failed to get abi")
	vs, err := contractABI.Methods["transfer"].Inputs.UnpackValues(data[4:])
	assert.Nil(t, err, "failed to unpack data")
	addr := c.AbiConvertToAddress(vs[0])
	amount := c.AbiConvertToInt(vs[1])
	assert.Equal(t, addr, "0xf1D7BEe92F49EAfc36b09b9953C05a2F4673cB40")
	assert.Equal(t, amount, big.NewInt(1000))
}
