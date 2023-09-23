from web3 import Web3
from absl import flags
flags.DEFINE_string("private_key", "", "private_key for this transaction")
flags.DEFINE_string("address", "", "address for this transaction")
import json, sys

endpoint = "https://rpc.ankr.com/avalanche_fuji"
endpoint = "https://polygon-rpc.com"
arguments = flags.FLAGS
arguments(sys.argv)

abiFile = "./swap_adapter.abi"
contractAddress = Web3.toChecksumAddress("0xb14C56E30eDaB4b2D6B15C7b1A4d31a5b104570a")

web3 = Web3(Web3.HTTPProvider(endpoint))
def swap(contractAddress, abiFile):
    with open(abiFile) as f:
        swapAbi = json.load(f)
    contract = web3.eth.contract(address=contractAddress, abi=swapAbi)
    swapParam = [["0xAb231A5744C8E6c45481754928cCfFFFD4aa0732", "0xB6076C93701D6a07266c31066B298AeC6dd65c2d"], 233000, 220000, "0xf1D7BEe92F49EAfc36b09b9953C05a2F4673cB40", "0x"]
    tx = contract.functions.swap(swapParam).buildTransaction({
        "from": arguments.address,
        "nonce": web3.eth.getTransactionCount(arguments.address),
        "gas": 8000000,
        #"gasPrice": web3.eth.gasPrice,
        "chainId": 43113,
        "maxFeePerGas": 26000000000,
        "maxPriorityFeePerGas": 1000000000,
    })
    signedTxn = web3.eth.account.sign_transaction(tx, private_key=arguments.private_key)
    txHash = web3.eth.sendRawTransaction(signedTxn.rawTransaction)
    print(txHash.hex())

def getAmoutOut(contractAddress, abiFile):
    with open(abiFile) as f:
        swapAbi = json.load(f)
    contract = web3.eth.contract(address=contractAddress, abi=swapAbi)
    param = [Web3.toChecksumAddress("0x55bb4d4b4545a886df159354e5fa5791c2d13496"), Web3.toChecksumAddress("0xc2132d05d31c914a87c6611c10748aeb04b58e8f")]
    result = contract.functions.getAmountOut(param, 4171864654411494500000).call()
    print(result)

if __name__ == "__main__":
    #swap(contractAddress, abiFile)
    getAmoutOut(contractAddress, abiFile)