// This is an integration test for showing how to use IFOC
// kernel with btcwallet. The smart property gets issued and
// goes through three different addresses.

package gochroma_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/conformal/btcnet"
	"github.com/conformal/btcrpcclient"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/davecgh/go-spew/spew"
	"github.com/jimmysong/gochroma"
)

var (
	seed = make([]byte, 32)
)

var _ = spew.Dump

func setUp() (*gochroma.BlockExplorer, *gochroma.Wallet, func(), error) {
	// create some temporary files and directories on the system
	systemTmp := os.TempDir()
	walletTmp, err := ioutil.TempDir(systemTmp, "cc_test")
	if err != nil {
		return nil, nil, nil, err
	}
	tempLocation := filepath.Join(walletTmp, "wallet.bin")
	btcd1Tmp, err := ioutil.TempDir(systemTmp, "cc_test")
	if err != nil {
		return nil, nil, nil, err
	}
	btcd2Tmp, err := ioutil.TempDir(systemTmp, "cc_test")
	if err != nil {
		return nil, nil, nil, err
	}

	// create some addresses to work with
	net := &btcnet.SimNetParams

	wallet, err := gochroma.CreateWallet(seed, tempLocation, net)
	if err != nil {
		return nil, nil, nil, err
	}

	addr, err := wallet.NewUncoloredAddress()

	// run btcd so that it starts mining
	path := os.Getenv("GOPATH")
	btcd := filepath.Join(path, "bin", "btcd")
	btcctl := filepath.Join(path, "bin", "btcctl")

	btcd1RPCPort := "47321"
	btcd1Port := "47311"
	btcd1 := exec.Command(btcd, "--datadir="+btcd1Tmp+"/",
		"--logdir="+btcd1Tmp+"/", "--debuglevel=debug", "--simnet",
		"--listen=:"+btcd1Port, "--rpclisten=:"+btcd1RPCPort, "--rpcuser=user",
		"--rpcpass=pass", "--rpccert=rpc.cert", "--rpckey=rpc.key")
	btcd2RPCPort := "47322"
	btcd2Port := "47312"
	miningStr := fmt.Sprintf("--miningaddr=%v", addr.String())
	btcd2 := exec.Command(btcd, "--datadir="+btcd2Tmp+"/",
		"--logdir="+btcd2Tmp+"/", "--debuglevel=debug", "--simnet",
		"--listen=:"+btcd2Port, "--rpclisten=:"+btcd2RPCPort,
		"--connect=localhost:"+btcd1Port, "--rpcuser=user", "--rpcpass=pass",
		"--rpccert=rpc.cert", "--rpckey=rpc.key", "--generate", miningStr)

	if err = btcd1.Start(); err != nil {
		return nil, nil, nil, err
	}
	if err = btcd2.Start(); err != nil {
		return nil, nil, nil, err
	}
	time.Sleep(1000 * time.Millisecond)

	// Now make a connection to the btcd for colored coins
	certs, err := ioutil.ReadFile("rpc.cert")
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
		exec.Command(btcctl, "-C", "btcctl2.conf", "stop").Output()
		exec.Command(btcctl, "-C", "btcctl1.conf", "stop").Output()
		os.RemoveAll(btcd1Tmp)
		os.RemoveAll(btcd2Tmp)
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
