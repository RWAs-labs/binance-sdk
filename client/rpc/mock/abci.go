package mock

import (
	abci "github.com/tendermint/tendermint/abci/types"
	libbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	"gitlab.com/thorchain/binance-sdk/client/rpc"
)

// ABCIApp will send all abci related request to the named app,
// so you can test app behavior from a client without needing
// an entire tendermint node
type ABCIApp struct {
	App abci.Application
}

var (
	_ rpc.ABCIClient = ABCIApp{}
	_ rpc.ABCIClient = ABCIMock{}
	_ rpc.ABCIClient = (*ABCIRecorder)(nil)
)

func (a ABCIApp) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return &ctypes.ResultABCIInfo{Response: a.App.Info(proxy.RequestInfo)}, nil
}

func (a ABCIApp) ABCIQuery(path string, data libbytes.HexBytes) (*ctypes.ResultABCIQuery, error) {
	return a.ABCIQueryWithOptions(path, data, client.DefaultABCIQueryOptions)
}

func (a ABCIApp) ABCIQueryWithOptions(path string, data libbytes.HexBytes, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	q := a.App.Query(abci.RequestQuery{
		Data:   data,
		Path:   path,
		Height: opts.Height,
		Prove:  opts.Prove,
	})
	return &ctypes.ResultABCIQuery{Response: q}, nil
}

// NOTE: Caller should call a.App.Commit() separately,
// this function does not actually wait for a commit.
// TODO: Make it wait for a commit and set res.Height appropriately.
func (a ABCIApp) BroadcastTxCommit(tx types.Tx) (*rpc.ResultBroadcastTxCommit, error) {
	return &rpc.ResultBroadcastTxCommit{}, nil
}

func (a ABCIApp) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	c := a.App.CheckTx(abci.RequestCheckTx{Tx: tx})
	// and this gets written in a background thread...
	if !c.IsErr() {
		go func() { a.App.DeliverTx(abci.RequestDeliverTx{Tx: tx}) }() // nolint: errcheck
	}
	return &ctypes.ResultBroadcastTx{Code: c.Code, Data: c.Data, Log: c.Log, Hash: tx.Hash()}, nil
}

func (a ABCIApp) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	c := a.App.CheckTx(abci.RequestCheckTx{Tx: tx})
	// and this gets written in a background thread...
	if !c.IsErr() {
		go func() { a.App.DeliverTx(abci.RequestDeliverTx{Tx: tx}) }() // nolint: errcheck
	}
	return &ctypes.ResultBroadcastTx{Code: c.Code, Data: c.Data, Log: c.Log, Hash: tx.Hash()}, nil
}

// ABCIMock will send all abci related request to the named app,
// so you can test app behavior from a client without needing
// an entire tendermint node
type ABCIMock struct {
	Info            Call
	Query           Call
	BroadcastCommit Call
	Broadcast       Call
}

func (m ABCIMock) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	res, err := m.Info.GetResponse(nil)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultABCIInfo{Response: res.(abci.ResponseInfo)}, nil
}

func (m ABCIMock) ABCIQuery(path string, data libbytes.HexBytes) (*ctypes.ResultABCIQuery, error) {
	return m.ABCIQueryWithOptions(path, data, client.DefaultABCIQueryOptions)
}

func (m ABCIMock) ABCIQueryWithOptions(path string, data libbytes.HexBytes, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	res, err := m.Query.GetResponse(QueryArgs{path, data, opts.Height, opts.Prove})
	if err != nil {
		return nil, err
	}
	resQuery := res.(abci.ResponseQuery)
	return &ctypes.ResultABCIQuery{Response: resQuery}, nil
}

func (m ABCIMock) BroadcastTxCommit(tx types.Tx) (*rpc.ResultBroadcastTxCommit, error) {
	return &rpc.ResultBroadcastTxCommit{}, nil
}

func (m ABCIMock) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := m.Broadcast.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTx), nil
}

func (m ABCIMock) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := m.Broadcast.GetResponse(tx)
	if err != nil {
		return nil, err
	}
	return res.(*ctypes.ResultBroadcastTx), nil
}

// ABCIRecorder can wrap another type (ABCIApp, ABCIMock, or Client)
// and record all ABCI related calls.
type ABCIRecorder struct {
	Client rpc.ABCIClient
	Calls  []Call
}

func NewABCIRecorder(client rpc.ABCIClient) *ABCIRecorder {
	return &ABCIRecorder{
		Client: client,
		Calls:  []Call{},
	}
}

type QueryArgs struct {
	Path   string
	Data   libbytes.HexBytes
	Height int64
	Prove  bool
}

func (r *ABCIRecorder) addCall(call Call) {
	r.Calls = append(r.Calls, call)
}

func (r *ABCIRecorder) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	res, err := r.Client.ABCIInfo()
	r.addCall(Call{
		Name:     "abci_info",
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *ABCIRecorder) ABCIQuery(path string, data libbytes.HexBytes) (*ctypes.ResultABCIQuery, error) {
	return r.ABCIQueryWithOptions(path, data, client.DefaultABCIQueryOptions)
}

func (r *ABCIRecorder) ABCIQueryWithOptions(path string, data libbytes.HexBytes, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	res, err := r.Client.ABCIQueryWithOptions(path, data, opts)
	r.addCall(Call{
		Name:     "abci_query",
		Args:     QueryArgs{path, data, opts.Height, opts.Prove},
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *ABCIRecorder) BroadcastTxCommit(tx types.Tx) (*rpc.ResultBroadcastTxCommit, error) {
	return &rpc.ResultBroadcastTxCommit{}, nil
}

func (r *ABCIRecorder) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := r.Client.BroadcastTxAsync(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_async",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}

func (r *ABCIRecorder) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	res, err := r.Client.BroadcastTxSync(tx)
	r.addCall(Call{
		Name:     "broadcast_tx_sync",
		Args:     tx,
		Response: res,
		Error:    err,
	})
	return res, err
}
