// This is an integration test for showing how to use IFOC
// kernel with btcwallet. The smart property gets issued and
// goes through three different addresses.

package gochroma_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/jimmysong/gochroma"
	"github.com/monetas/tools/btclib"
)

var (
	seed = make([]byte, 32)
)

func setUp() (*gochroma.BlockExplorer, *gochroma.Wallet, func(), error) {

	// create some temporary files and directories on the system
	systemTmp := os.TempDir()
	walletTmp, err := ioutil.TempDir(systemTmp, "integration_test")
	if err != nil {
		return nil, nil, nil, err
	}
	tempLocation := filepath.Join(walletTmp, "wallet.bin")
	rpcCertLoc := "rpc.cert"
	rpcKeyLoc := "rpc.key"

	// create a wallet
	net := &btcnet.SimNetParams
	wallet, err := gochroma.CreateWallet(seed, tempLocation, net)
	if err != nil {
		return nil, nil, nil, err
	}

	addr, err := wallet.NewUncoloredAddress()

	// run btcd so that it starts mining
	path := os.Getenv("GOPATH")
	btcd := filepath.Join(path, "bin", "btcd")

	btcd1RPCPort, btcd2RPCPort := "47321", "47322"
	btcd1Port, btcd2Port := "47311", "47312"
	user, pass := "user", "pass"
	miningStr := fmt.Sprintf("--miningaddr=%v", addr.String())
	connectStr := fmt.Sprintf("--connect=localhost:%v", btcd1Port)

	btcd1, err := btclib.StartBTCD(btclib.BtcdConfig{
		ExecPath:    btcd,
		Listen:      "127.0.0.1:" + btcd1Port,
		RPCUser:     user,
		RPCPassword: pass,
		RPCListen:   "127.0.0.1:" + btcd1RPCPort,
		RPCCert:     rpcCertLoc,
		RPCKey:      rpcKeyLoc,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	btcd2, err := btclib.StartBTCD(btclib.BtcdConfig{
		ExecPath:         btcd,
		Listen:           "127.0.0.1:" + btcd2Port,
		RPCUser:          user,
		RPCPassword:      pass,
		RPCListen:        "127.0.0.1:" + btcd2RPCPort,
		RPCCert:          rpcCertLoc,
		RPCKey:           rpcKeyLoc,
		AdditionalParams: []string{connectStr, "--generate", miningStr},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	time.Sleep(1000 * time.Millisecond)

	// Now make a connection to the btcd for colored coins
	certs, err := ioutil.ReadFile(rpcCertLoc)
	if err != nil {
		return nil, nil, nil, err
	}
	connConfig := &btcrpcclient.ConnConfig{
		Host:         "localhost:" + btcd2RPCPort,
		User:         "user",
		Pass:         "pass",
		Certificates: certs,
		HttpPostMode: true,
	}
	blockReaderWriter, err := gochroma.NewBtcdBlockExplorer(net, connConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	b := &gochroma.BlockExplorer{blockReaderWriter}

	tearDown := func() {
		btcd1.Process.Kill()
		btcd2.Process.Kill()
		os.RemoveAll(walletTmp)
	}

	return b, wallet, tearDown, nil
}

func TestCC(t *testing.T) {

	b, wallet, tearDown, err := setUp()
	if err != nil {
		fmt.Printf("Couldn't set up in order to run test: %v\nSkipping...", err)
		return
	}
	defer tearDown()

	// grab an outpoint we can sign (we're mining to the first addr)
	block, err := b.BlockAtHeight(1)
	if err != nil {
		t.Fatalf("%v\n", err)
	}
	coinbaseTx := block.Transactions()[0]
	outPoint := btcwire.NewOutPoint(coinbaseTx.Sha(), 0)
	_, err = wallet.NewUncoloredOutPoint(b, outPoint)
	if err != nil {
		t.Fatal(err)
	}

	v, err := wallet.ColorBalance(gochroma.UncoloredColorId)
	want := gochroma.ColorValue(5000000000)
	if *v != want {
		t.Fatalf("unexpected balance: want %d, got %d", want, *v)
	}

	// grab the kernel
	ifoc, err := gochroma.GetColorKernel("IFOC")
	if err != nil {
		t.Fatalf("error getting ifoc kernel: %v", err)
	}

	// Create the issuing tx
	cd, err := wallet.IssueColor(b, ifoc, gochroma.ColorValue(1), 10000)
	if err != nil {
		t.Fatalf("failed to issue color: %v", err)
	}

	// check that the color value of the outpoint is what we expect
	currentOut := cd.Genesis
	colorIn, err := ifoc.OutPointToColorIn(b, cd.Genesis, currentOut)
	if err != nil {
		t.Fatalf("cannot get color value: %v, %v", cd.Genesis, err)
	}
	if colorIn.ColorValue != 1 {
		t.Fatalf("wrong color value: %v", colorIn)
	}

	// Hop 3 more times
	for i := 1; i < 4; i++ {
		// transfer to another address we control
		addr, err := wallet.NewColorAddress(cd)
		if err != nil {
			t.Fatalf("cannot get color address for cd %v: %v", cd, err)
		}
		addrMap := map[btcutil.Address]gochroma.ColorValue{
			addr: gochroma.ColorValue(1),
		}
		tx, err := wallet.Send(b, cd, addrMap, 10000)
		if err != nil {
			t.Fatalf("sending failed: %v", err)
		}

		// old endpoint should have a value of 0
		cv, err := cd.ColorValue(b, currentOut)
		if err != nil {
			t.Fatalf("cannot get color value: %v", err)
		}
		if *cv != 0 {
			t.Fatalf("wrong color value: %v", colorIn)
		}
		// new endpoint should have a value of 1
		shaHash, err := tx.TxSha()
		if err != nil {
			t.Fatalf("failed to compute hash: %v", err)
		}
		currentOut = btcwire.NewOutPoint(&shaHash, 0)
		cv, err = cd.ColorValue(b, currentOut)
		if err != nil {
			t.Fatalf("cannot get color value: %s", err)
		}
		if *cv != 1 {
			t.Fatalf("wrong color value: %v", colorIn)
		}

		// add this back to our wallet since it's ours anyway
		_, err = wallet.NewColorOutPoint(b, currentOut, cd)
		if err != nil {
			t.Fatal(err)
		}
	}
}
