from web3 import Web3
from absl import flags
import eth_abi
flags.DEFINE_string("private_key", "", "private_key for this transaction")
flags.DEFINE_string("address", "", "address for this transaction")
import json, sys

endpoint = "https://rpc.ankr.com/avalanche_fuji"
endpoint = "https://polygon-rpc.com"
endpoint = "https://bsc-dataseed.binance.org"
arguments = flags.FLAGS
arguments(sys.argv)
web3 = Web3(Web3.HTTPProvider(endpoint))

abiFile = "./izumi_quoter.abi"
contractAddress = Web3.toChecksumAddress("0x64b005eD986ed5D6aeD7125F49e61083c46b8e02")

with open(abiFile) as f:
    izumiAbi = json.load(f)
contract = web3.eth.contract(address=contractAddress, abi=izumiAbi)
data = eth_abi.encode_abi(["address", "uint24", "address"], ["0x55d398326f99059ff775485246999027b3197955", 
    2000, "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d"])
contract.functions.swapAmount(1000000000, data).call()