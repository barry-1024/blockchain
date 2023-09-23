from web3 import Web3, Account
from absl import flags
import json, sys, yaml
flags.DEFINE_string("private_key", "", "private_key for this transaction")
flags.DEFINE_string("address", "", "address for this transaction")
flags.DEFINE_string("foc_file", "", "foc accounts file path")
flags.DEFINE_string("metadata_dir", "../../../onchain/contracts-metadata", "root directory of metadata")
flags.DEFINE_integer("chain_id", 0, "chain id")
arguments = flags.FLAGS
arguments(sys.argv)
abiFile = "./protocol.abi"

def loadContract(dir, contractName, chainId):
    file = dir + "/deployment/contracts.yaml"
    with open(file) as f:
        contracts = yaml.load(f, Loader=yaml.FullLoader)
    for c in contracts:
        if c["contract_name"] == contractName:
            abiFile = dir + "/deployment/" + c["abi_file"]
            with open(abiFile) as f:
                abi = json.load(f)
            for d in c["deployment"]:
                if d["chain_id"] == chainId:
                    return abi, d["address"]
            
    return None, None

def loadEndpoints(dir):
    file = dir + "/chains/chains.yaml"
    with open(file) as f:
        chains = yaml.load(f, Loader=yaml.FullLoader)
    endpoints = {}
    for chain in chains:
        endpoints[chain["id"]] = chain["endpoints"]
    return endpoints

def ProtocolUpdateFOCAccounts(web3, address, abi, focs):
    contract = web3.eth.contract(address=address, abi=abi)
    fromAddress = Account.from_key(arguments.private_key).address
    tx = contract.functions.updateFOCAccounts(focs, True).buildTransaction({
        "from": fromAddress,
        "nonce": web3.eth.getTransactionCount(fromAddress),
        "gas": 8000000,
        "chainId": arguments.chain_id,
        "maxFeePerGas": 26000000000,
        "maxPriorityFeePerGas": 1000000000,
    })
    signedTxn = web3.eth.account.sign_transaction(tx, private_key=arguments.private_key)
    txHash = web3.eth.sendRawTransaction(signedTxn.rawTransaction)
    return txHash.hex()

def loadFOCS(file):
    with open(file) as f:
        focs = yaml.load(f, Loader=yaml.FullLoader)
    for chain in focs:
        if chain["chain_id"] == arguments.chain_id:
            return chain["accounts"]
    return []

if __name__ == "__main__":
    endpoints = loadEndpoints(arguments.metadata_dir)
    endpoint = endpoints[arguments.chain_id]
    web3 = Web3(Web3.HTTPProvider(endpoint))
    abi, address = loadContract(arguments.metadata_dir, "UxuyProtocol", arguments.chain_id)
    focs = loadFOCS(arguments.foc_file)
    if len(focs) == 0:
        print("no focs")
        exit()
    print(ProtocolUpdateFOCAccounts(web3, address, abi, focs))