package main

import (
	"C"

	"github.com/ethereum/go-ethereum/internal/debug"
)
import (
	"context"
	"unsafe"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
)

var stack *node.Node

//export open_database
func open_database(datadir *C.char) C.int {
	// glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	// glogger.Verbosity(1)
	// log.Root().SetHandler(glogger)
	stack, _ = makeReadOnlyNode("/mnt/pepper/intern/home/dmaclennan/share/geth_archive_20220707")
	if err := startNode(stack, false); err != nil {
		return C.int(-1)
	}
	return C.int(0)
}

//export wrapper_call
func wrapper_call(cargs *C.void, clen C.int) (*C.char, C.int) {
	rawData := C.GoBytes(unsafe.Pointer(cargs), clen)
	log.Info(string(rawData))
	server, err := stack.RPCHandler()
	if err == nil {
		res, err := server.ServeRawRequest(context.Background(), rawData)
		if err == nil {
			log.Info("Good")
			c_res := C.CString(res)
			return c_res, C.int(len(res))
		} else {
			log.Error("There is an error")
			return nil, 0
		}
	} else {
		log.Error("Could not get the RPC Handler")
	}
	return nil, 0
}

//export close_database
func close_database() {
	stack.Close()
}

func startNode(stack *node.Node, isConsole bool) error {
	debug.Memsize.Add("node", stack)
	StartNode(stack)
	return nil
}

func main() {

}
