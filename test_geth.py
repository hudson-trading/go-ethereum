import site
site.addsitedir("/scratch/haowang/remote/quarinterns/pyimports")
from defi.geth_shared_lib import GethReadOnlyProvider
from defi.base import sha_geth_read_only
from web3 import Web3
from defi.erc20 import ERC20, ERC20Addr
from defi.contract import Contract
with sha_geth_read_only() as prov:
    c = Contract(abi=ERC20.artifact().get("abi"), address=ERC20Addr.DAI, provider_in=prov)
    print(c.event_filter.Transfer(from_block= 0 ))
