package contractor

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/NebulousLabs/Sia/build"
	"github.com/NebulousLabs/Sia/modules"
	"github.com/NebulousLabs/Sia/types"
)

// newStub is used to test the New function. It implements all of the contractor's
// dependencies.
type newStub struct{}

// consensus set stubs
func (newStub) ConsensusSetPersistentSubscribe(modules.ConsensusSetSubscriber, modules.ConsensusChangeID) error {
	return nil
}

// wallet stubs
func (newStub) NextAddress() (uc types.UnlockConditions, err error) { return }
func (newStub) StartTransaction() modules.TransactionBuilder        { return nil }

// transaction pool stubs
func (newStub) AcceptTransactionSet([]types.Transaction) error { return nil }

// hdb stubs
func (newStub) Host(modules.NetAddress) (settings modules.HostSettings, ok bool) { return }
func (newStub) RandomHosts(int, []modules.NetAddress) []modules.HostSettings     { return nil }

// TestNew tests the New function.
func TestNew(t *testing.T) {
	// Using a stub implementation of the dependencies is fine, as long as its
	// non-nil.
	var stub newStub
	dir := build.TempDir("contractor", "TestNew")

	// Sane values.
	_, err := New(stub, stub, stub, stub, dir)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// Nil consensus set.
	_, err = New(nil, stub, stub, stub, dir)
	if err != errNilCS {
		t.Fatalf("expected %v, got %v", errNilCS, err)
	}

	// Nil wallet.
	_, err = New(stub, nil, stub, stub, dir)
	if err != errNilWallet {
		t.Fatalf("expected %v, got %v", errNilWallet, err)
	}

	// Nil transaction pool.
	_, err = New(stub, stub, nil, stub, dir)
	if err != errNilTpool {
		t.Fatalf("expected %v, got %v", errNilTpool, err)
	}

	// Bad persistDir.
	_, err = New(stub, stub, stub, stub, "")
	if !os.IsNotExist(err) {
		t.Fatalf("expected invalid directory, got %v", err)
	}

	// Corrupted persist file.
	ioutil.WriteFile(filepath.Join(dir, "contractor.json"), []byte{1, 2, 3}, 0666)
	_, err = New(stub, stub, stub, stub, dir)
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Fatalf("expected invalid json, got %v", err)
	}

	// Corrupted logfile.
	os.RemoveAll(filepath.Join(dir, "contractor.log"))
	f, err := os.OpenFile(filepath.Join(dir, "contractor.log"), os.O_CREATE, 0000)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = New(stub, stub, stub, stub, dir)
	if !os.IsPermission(err) {
		t.Fatalf("expected permissions error, got %v", err)
	}
}

// testWalletShim is used to test the walletBridge type.
type testWalletShim struct {
	nextAddressCalled bool
	startTxnCalled    bool
}

// These stub implementations for the walletShim interface set their respective
// booleans to true, allowing tests to verify that they have been called.
func (ws *testWalletShim) NextAddress() (types.UnlockConditions, error) {
	ws.nextAddressCalled = true
	return types.UnlockConditions{}, nil
}
func (ws *testWalletShim) StartTransaction() modules.TransactionBuilder {
	ws.startTxnCalled = true
	return nil
}

// TestWalletBridge tests the walletBridge type.
func TestWalletBridge(t *testing.T) {
	shim := new(testWalletShim)
	bridge := walletBridge{shim}
	bridge.NextAddress()
	if !shim.nextAddressCalled {
		t.Error("NextAddress was not called on the shim")
	}
	bridge.StartTransaction()
	if !shim.startTxnCalled {
		t.Error("StartTransaction was not called on the shim")
	}
}
