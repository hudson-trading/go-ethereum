from cffi import FFI
import os
from web3 import Web3
import json
ffi = FFI()
ffi.cdef("""
extern void open_database(char* datadir);
struct wrapper_call_return {
	char* r0;
	int r1;
};
extern struct wrapper_call_return wrapper_call(char* cargs, int clen);
extern void close_database();""")
path = os.path.join("./build/bin/libcgeth.so")
cbl = ffi.dlopen(path)
path = "/home/dmaclennan/share/geth_backup_20220113/geth"
path = path.encode("utf-8")
cbl.open_database(path)
data = json.dumps({"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest", True],"id":1})
data = data.encode("utf-8")
len_of_data = ffi.cast("int", len(data))
res = cbl.wrapper_call(data, len_of_data)
print(ffi.unpack(res.r0, res.r1))
cbl.close_database()
