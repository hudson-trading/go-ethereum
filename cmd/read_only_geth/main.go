package main

import (
	"C"
	"unsafe"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/internal/debug"
	"github.com/ethereum/go-ethereum/rpc"
)
import (
	"context"

	"github.com/ethereum/go-ethereum/node"
)

var stack *node.ReadOnlyNode

//export open_database
func open_database(datadir *C.char) C.int {
	if stack != nil {
		return C.int(-1)
	}
	go_datadir := C.GoString(datadir)
	stack, _ = makeFullNode(go_datadir)
	if err := startNode(stack, false); err != nil {
		return C.int(-1)
	}
	return C.int(0)
}

//export wrapper_call
func wrapper_call(cargs *C.char, clen C.int) (*C.char, C.int) {
	rawData := C.GoBytes(unsafe.Pointer(cargs), clen)
	server, _ := stack.RPCHandler()
	msg, _ := rpc.ParseMessage(rawData)
	h := rpc.NewReadOnlyHandler(context.Background(), rpc.RandomIDGenerator(), server.Services())
	cproc := rpc.DefaultCallProc()
	res := rpc.HandleCallMsg(h, cproc, msg[0])
	resStr := C.CString(res.String())
	return resStr, C.int(len(res.String()))
}

//export close_database
func close_database() {
	stack.Close()
}

func startNode(stack *node.ReadOnlyNode, isConsole bool) error {
	debug.Memsize.Add("node", stack)

	// Start up the node itself
	utils.StartNodeReadOnly(stack, isConsole)

	// Create a client to interact with local geth node.
	_, err := stack.Attach()
	if err != nil {
		return err
	}
	return nil
}

func main() {
}
