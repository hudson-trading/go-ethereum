package node

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type ReadOnlyNode struct {
	Node
}

// Overrite New() for ReadOnlyNode
func NewReadOnly(conf *Config) (*ReadOnlyNode, error) {
	if conf.DataDir != "" {
		absdatadir, err := filepath.Abs(conf.DataDir)
		if err != nil {
			return nil, err
		}
		conf.DataDir = absdatadir
	}
	if conf.Logger == nil {
		conf.Logger = log.New()
	}
	read_node := &ReadOnlyNode{Node{
		config:        conf,
		inprocHandler: rpc.NewServer(),
		eventmux:      new(event.TypeMux),
		log:           conf.Logger,
		stop:          make(chan struct{}),
		server:        nil,
		databases:     make(map[*closeTrackingDB]struct{}),
	}}
	// Register built-in APIs.
	read_node.rpcAPIs = append(read_node.rpcAPIs, read_node.apis()...)
	return read_node, nil
}

// Overrite Close() for ReadOnlyNode
func (n *ReadOnlyNode) Close() error {
	n.startStopLock.Lock()
	defer n.startStopLock.Unlock()

	n.lock.Lock()
	state := n.state
	n.lock.Unlock()
	switch state {
	case initializingState:
		// The node was never started.
		return n.doClose(nil)
	case runningState:
		// The node was started, release resources acquired by Start().
		var errs []error
		if err := n.stopServices(n.lifecycles); err != nil {
			errs = append(errs, err)
		}
		return n.doClose(errs)
	case closedState:
		return ErrNodeStopped
	default:
		panic(fmt.Sprintf("node is in unknown state %d", state))
	}
}

func (n *ReadOnlyNode) doClose(errs []error) error {
	// Close databases. This needs the lock because it needs to
	// synchronize with OpenDatabase*.
	n.lock.Lock()
	n.state = closedState
	errs = append(errs, n.closeDatabases()...)
	n.lock.Unlock()

	if n.keyDirTemp {
		if err := os.RemoveAll(n.keyDir); err != nil {
			errs = append(errs, err)
		}
	}
	n.closeDataDir()
	// Unblock n.Wait.
	close(n.stop)

	// Report any errors that might have occurred.
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return fmt.Errorf("%v", errs)
	}
}

func (n *ReadOnlyNode) openEndpoints() error {
	// start RPC endpoints
	var err error
	if err = n.startInProc(); err != nil {
		n.server.Stop()
		return err
	}
	return err
}

func (n *ReadOnlyNode) stopServices(running []Lifecycle) error {
	// Stop running lifecycles in reverse order.
	failure := &StopError{Services: make(map[reflect.Type]error)}
	for i := len(running) - 1; i >= 0; i-- {
		if err := running[i].Stop(); err != nil {
			failure.Services[reflect.TypeOf(running[i])] = err
		}
	}
	if len(failure.Services) > 0 {
		return failure
	}
	return nil
}

func (n *ReadOnlyNode) ResolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	c := n.config
	if c.DataDir == "" {
		return ""
	}
	return filepath.Join(c.DataDir, path)
}

func (n *ReadOnlyNode) Start() error {
	n.startStopLock.Lock()
	defer n.startStopLock.Unlock()

	n.lock.Lock()
	switch n.state {
	case runningState:
		n.lock.Unlock()
		return ErrNodeRunning
	case closedState:
		n.lock.Unlock()
		return ErrNodeStopped
	}
	n.state = runningState
	// open networking and RPC endpoints
	err := n.openEndpoints()
	lifecycles := make([]Lifecycle, len(n.lifecycles))
	copy(lifecycles, n.lifecycles)
	n.lock.Unlock()

	// Check if endpoint startup failed.
	if err != nil {
		n.doClose(nil)
		return err
	}
	// Start all registered lifecycles.
	var started []Lifecycle
	for _, lifecycle := range lifecycles {
		if err = lifecycle.Start(); err != nil {
			break
		}
		started = append(started, lifecycle)
	}
	// Check if any lifecycle failed to start.
	if err != nil {
		n.stopServices(started)
		n.doClose(nil)
	}
	return err
}

func (n *ReadOnlyNode) OpenDatabase(name string, cache, handles int, namespace string, readonly bool) (ethdb.Database, error) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.state == closedState {
		return nil, ErrNodeStopped
	}
	var db ethdb.Database
	var err error
	if n.config.DataDir == "" {
		db = rawdb.NewMemoryDatabase()
	} else {
		db, err = rawdb.NewLevelDBDatabase(n.ResolvePath(name), cache, handles, namespace, readonly)
	}

	if err == nil {
		db = n.wrapDatabase(db)
	}
	return db, err
}
