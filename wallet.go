package gochroma

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/conformal/bolt"
	"github.com/conformal/btcnet"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcutil/hdkeychain"
	"github.com/conformal/btcwire"
)

var (
	keyBucketName                  = []byte("keys")
	idBucketName                   = []byte("id counter")
	accountBucketName              = []byte("account info")
	colorDefinitionBucketName      = []byte("color definitions")
	colorOutPointBucketName        = []byte("color outpoints")
	outPointIndexBucketName        = []byte("out point index")
	scriptToAccountIndexBucketName = []byte("script account index")

	colorIdKey    = []byte("color id")
	outPointIdKey = []byte("out point id")
	pubKeyName    = []byte("extended pubkey")
	privKeyName   = []byte("encrypted extended privkey")
	netName       = []byte("net")

	uncoloredAcctNum = uint32(0)
	issuingAcctNum   = uint32(1)

	uncoloredColorId = []byte{0, 0, 0, 0}
)

type ColorId []byte
type OutPointId []byte

func increment(id []byte) []byte {
	x := DeserializeUint32(id) + 1
	return SerializeUint32(x)
}

func openColorDB(dbPath string) (*bolt.DB, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		fmt.Printf("failed to open database\n")
		return nil, err
	}

	// Initialize the buckets and main db fields as needed.
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(keyBucketName)
		if err != nil {
			fmt.Printf("failed to create key bucket\n")
			return err
		}
		_, err = tx.CreateBucketIfNotExists(idBucketName)
		if err != nil {
			fmt.Printf("failed to create id bucket\n")
			return err
		}
		_, err = tx.CreateBucketIfNotExists(accountBucketName)
		if err != nil {
			fmt.Printf("failed to create account bucket\n")
			return err
		}
		_, err := tx.CreateBucketIfNotExists(colorDefinitionBucketName)
		if err != nil {
			fmt.Printf("failed to create colordef bucket\n")
			return err
		}
		_, err = tx.CreateBucketIfNotExists(colorOutPointBucketName)
		if err != nil {
			fmt.Printf("failed to create outpoint bucket\n")
			return err
		}
		_, err = tx.CreateBucketIfNotExists(outPointIndexBucketName)
		if err != nil {
			fmt.Printf("failed to create outpoint index bucket\n")
			return err
		}
		_, err = tx.CreateBucketIfNotExists(scriptToAccountIndexBucketName)
		if err != nil {
			fmt.Printf("failed to create script to account/index bucket\n")
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("failed to update database\n")
		return nil, err
	}

	return db, nil
}

type Wallet struct {
	ColorDB *bolt.DB
	pubKey  *hdkeychain.ExtendedKey
	privKey *hdkeychain.ExtendedKey
	Net     *btcnet.Params
}

func CreateWallet(seed []byte, path string, net *btcnet.Params) (*Wallet, error) {
	var err error
	if seed == nil {
		seed, err = hdkeychain.GenerateSeed(32)
		if err != nil {
			return nil, errors.New("failed to generate seed")
		}
	}
	if len(seed) != 32 {
		return nil, errors.New("Need a 32 byte seed")
	}

	// get the hd root
	root, err := hdkeychain.NewMaster(seed)
	if err != nil {
		return nil, errors.New("failed to derive master extended key")
	}
	pub, err := root.Neuter()
	if err != nil {
		return nil, errors.New("failed to get extended public key")
	}

	colorDB, err := openColorDB(path)
	if err != nil {
		return nil, err
	}

	err = colorDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(keyBucketName)
		err := b.Put(pubKeyName, []byte(pub.String()))
		if err != nil {
			return err
		}
		err = b.Put(privKeyName, []byte(root.String()))
		if err != nil {
			return err
		}
		err = b.Put(netName, []byte(net.Name))
		if err != nil {
			return err
		}
		b2 := tx.Bucket(accountBucketName)
		err = b2.Put(SerializeUint32(uncoloredAcctNum), SerializeUint32(0))
		if err != nil {
			return err
		}
		err = b2.Put(SerializeUint32(issuingAcctNum), SerializeUint32(0))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Wallet{colorDB, pub, root, net}, nil
}

func OpenWallet(path string) (*Wallet, error) {
	var err error
	colorDB, err := openColorDB(path)
	if err != nil {
		return nil, err
	}
	var pub, priv *hdkeychain.ExtendedKey
	var net *btcnet.Params
	err = colorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(keyBucketName)
		pubStr := string(b.Get(pubKeyName))
		pub, err = hdkeychain.NewKeyFromString(pubStr)
		if err != nil {
			return err
		}
		privStr := string(b.Get(privKeyName))
		priv, err = hdkeychain.NewKeyFromString(privStr)
		if err != nil {
			return err
		}
		netStr := string(b.Get(netName))
		if netStr == "mainnet" {
			net = &btcnet.MainNetParams
		} else if netStr == "testnet" {
			net = &btcnet.TestNet3Params
		} else if netStr == "simnet" {
			net = &btcnet.SimNetParams
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Wallet{colorDB, pub, priv, net}, nil
}

func (w *Wallet) CurrentId(bucketKey []byte) ([]byte, error) {
	var id []byte
	err := w.ColorDB.View(func(tx *bolt.Tx) error {
		var err error
		bucket := tx.Bucket(idBucketName)
		id = bucket.Get(bucketKey)
		if len(id) == 0 {
			id = SerializeUint32(uint32(1))
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (w *Wallet) NewId(bucketKey []byte) ([]byte, error) {
	id, err := w.CurrentId(bucketKey)
	if err != nil {
		return nil, err
	}

	err = w.ColorDB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(idBucketName)
		return bucket.Put(bucketKey, increment(id))
	})
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (w *Wallet) NewColorId() (ColorId, error) {
	return w.NewId(colorIdKey)
}

func (w *Wallet) NewOutPointId() (OutPointId, error) {
	return w.NewId(outPointIdKey)
}

func (w *Wallet) FetchOrAddDefinition(def string) (ColorId, error) {
	// check definition is valid
	cd, err := NewColorDefinitionFromStr(def)
	if err != nil {
		return nil, err
	}
	// check to see if it's already in db
	var colorId ColorId
	err = w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(colorDefinitionBucketName)
		colorId = b.Get([]byte(cd.HashString()))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(colorId) != 0 {
		return colorId, nil
	}

	// grab a new color id
	colorId, err = w.NewColorId()
	if err != nil {
		return nil, err
	}
	err = w.ColorDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(colorDefinitionBucketName)
		err := b.Put([]byte(cd.HashString()), colorId)
		if err != nil {
			return err
		}
		b2 := tx.Bucket(accountBucketName)
		return b2.Put(SerializeUint32(cd.AccountNumber()), SerializeUint32(0))
	})
	if err != nil {
		return nil, err
	}

	return colorId, nil
}

func (w *Wallet) Close() error {
	w.pubKey = nil
	w.privKey = nil
	return w.ColorDB.Close()
}

func (w *Wallet) Sign(pkScript []byte, tx *btcwire.MsgTx, txIndex int) error {
	var acct, index uint32
	err := w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(scriptToAccountIndexBucketName)
		raw := b.Get(pkScript)
		if len(raw) == 0 {
			str := fmt.Sprintf("wallet can't sign script: %x", pkScript)
			return errors.New(str)
		}
		acct = DeserializeUint32(raw[:4])
		index = DeserializeUint32(raw[4:])
		return nil
	})
	if err != nil {
		return err
	}
	acctKey, err := w.privKey.Child(acct)
	if err != nil {
		return err
	}
	indexKey, err := acctKey.Child(index)
	if err != nil {
		return err
	}
	privateKey, err := indexKey.ECPrivKey()
	if err != nil {
		return err
	}

	sigScript, err := btcscript.SignatureScript(
		tx, txIndex, pkScript, btcscript.SigHashAll, privateKey, true)
	if err != nil {
		str := fmt.Sprintf("cannot create sigScript: %s", err)
		return errors.New(str)
	}
	tx.TxIn[index].SignatureScript = sigScript
	return nil
}

func (w *Wallet) NewAddress(acct uint32) (btcutil.Address, error) {
	subKey, err := w.pubKey.Child(acct)
	if err != nil {
		return nil, err
	}
	var index uint32
	err = w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(accountBucketName)
		raw := b.Get(SerializeUint32(acct))
		if len(raw) == 0 {
			str := fmt.Sprintf("Account %d doesn't exist", acct)
			return errors.New(str)
		}
		index = DeserializeUint32(raw)
		return nil
	})
	if err != nil {
		return nil, err
	}

	key, err := subKey.Child(index)
	if err != nil {
		return nil, err
	}
	addr, err := key.Address(w.Net)
	if err != nil {
		return nil, err
	}

	err = w.ColorDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(accountBucketName)
		err = b.Put(SerializeUint32(acct), SerializeUint32(index+1))
		if err != nil {
			return err
		}
		b2 := tx.Bucket(scriptToAccountIndexBucketName)
		val := append(SerializeUint32(acct), SerializeUint32(index)...)
		pkScript, err := btcscript.PayToAddrScript(addr)
		if err != nil {
			return err
		}
		return b2.Put(pkScript, val)
	})
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func (w *Wallet) NewUncoloredAddress() (btcutil.Address, error) {
	return w.NewAddress(uncoloredAcctNum)
}

func (w *Wallet) NewIssuingAddress() (btcutil.Address, error) {
	return w.NewAddress(issuingAcctNum)
}

func (w *Wallet) NewColorAddress(cd *ColorDefinition) (btcutil.Address, error) {
	return w.NewAddress(cd.AccountNumber())
}

type ColorOutPoint struct {
	Id            OutPointId
	Tx            []byte
	Index         uint32
	Value         uint64
	Color         ColorId
	ColorValue    ColorValue
	SpendingTx    []byte
	SpendingIndex uint32
	Spent         bool
	PkScript      []byte
}

func (w *Wallet) FetchOutPointId(outPoint *btcwire.OutPoint) (*OutPointId, error) {
	var outPointId OutPointId
	serializedOutPoint, err := SerializeOutPoint(outPoint)
	if err != nil {
		return nil, err
	}
	err = w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(outPointIndexBucketName)
		outPointId = b.Get(serializedOutPoint)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(outPointId) == 0 {
		return nil, nil
	}
	return &outPointId, nil
}

func (w *Wallet) FetchColorId(cd *ColorDefinition) (*ColorId, error) {
	var colorId ColorId
	err := w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(colorDefinitionBucketName)
		colorId = b.Get([]byte(cd.HashString()))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(colorId) == 0 {
		return nil, nil
	}
	return &colorId, nil
}

func (w *Wallet) NewUncoloredOutPoint(b *BlockExplorer, outPoint *btcwire.OutPoint) (*ColorOutPoint, error) {
	// look up the outpoint and see if it's already in the db
	exists, err := w.FetchOutPointId(outPoint)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		str := fmt.Sprintf("out point %v already in db", outPoint)
		return nil, errors.New(str)
	}
	outPointId, err := w.NewOutPointId()
	if err != nil {
		return nil, err
	}

	// get tx data
	tx, err := b.OutPointTx(outPoint)
	if err != nil {
		return nil, err
	}
	msgTx := tx.MsgTx()
	txOut := msgTx.TxOut[outPoint.Index]
	satoshiValue := uint64(txOut.Value)
	pkScript := txOut.PkScript

	// construct the outpoint
	colorOutPoint := &ColorOutPoint{
		Id:         outPointId,
		Tx:         BigEndianBytes(&outPoint.Hash),
		Index:      outPoint.Index,
		Value:      satoshiValue,
		Color:      uncoloredColorId,
		ColorValue: ColorValue(satoshiValue),
		PkScript:   pkScript,
	}

	// store outpoint in DB
	err = w.storeOutPoint(colorOutPoint)
	if err != nil {
		return nil, err
	}

	return colorOutPoint, nil
}

func (w *Wallet) NewColorOutPoint(b *BlockExplorer, outPoint *btcwire.OutPoint, cd *ColorDefinition) (*ColorOutPoint, error) {
	// look up the outpoint and see if it's already in the db
	exists, err := w.FetchOutPointId(outPoint)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		str := fmt.Sprintf("out point %v already in db", outPoint)
		return nil, errors.New(str)
	}
	outPointId, err := w.NewOutPointId()
	if err != nil {
		return nil, err
	}

	// get tx data
	tx, err := b.OutPointTx(outPoint)
	if err != nil {
		return nil, err
	}
	msgTx := tx.MsgTx()
	txOut := msgTx.TxOut[outPoint.Index]
	satoshiValue := uint64(txOut.Value)
	pkScript := txOut.PkScript

	// get color data
	color, err := w.FetchOrAddDefinition(cd.String())
	if err != nil {
		return nil, err
	}
	colorValue, err := cd.ColorValue(b, outPoint)
	if err != nil {
		return nil, err
	}

	// construct the outpoint
	colorOutPoint := &ColorOutPoint{
		Id:         outPointId,
		Tx:         BigEndianBytes(&outPoint.Hash),
		Index:      outPoint.Index,
		Value:      satoshiValue,
		Color:      color,
		ColorValue: *colorValue,
		PkScript:   pkScript,
	}

	// store outpoint in DB
	err = w.storeOutPoint(colorOutPoint)
	if err != nil {
		return nil, err
	}

	return colorOutPoint, nil
}

func (w *Wallet) storeOutPoint(cop *ColorOutPoint) error {
	// TODO: update various indices
	return w.ColorDB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(colorOutPointBucketName)
		s, err := SerializeColorOutPoint(cop)
		if err != nil {
			return err
		}
		return b.Put(cop.Id, s)
	})
}

func (w *Wallet) allOutPoints() ([]*ColorOutPoint, error) {
	currentOutPointId, err := w.CurrentId(outPointIdKey)
	limit := DeserializeUint32(currentOutPointId)

	outPoints := make([]*ColorOutPoint, limit-1)
	err = w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(colorOutPointBucketName)
		for i := 0; i < int(limit-1); i++ {
			key := SerializeUint32(uint32(i + 1))
			raw := b.Get(key)
			outPoints[i], err = DeserializeColorOutPoint(raw)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return outPoints, nil
}

func (w *Wallet) fetchSpendable(b *BlockExplorer, colorId ColorId, needed ColorValue) ([]*ColorOutPoint, error) {
	outPoints, err := w.allOutPoints()
	if err != nil {
		return nil, err
	}
	var ret []*ColorOutPoint

	sum := ColorValue(0)
	for _, outPoint := range outPoints {
		if !outPoint.Spent && bytes.Compare(outPoint.Color, colorId) == 0 {
			// check again if the outPoint has been spent
			op, err := outPoint.OutPoint()
			if err != nil {
				return nil, err
			}
			spent, err := b.OutPointSpent(op)
			if err != nil {
				return nil, err
			}
			if *spent {
				// update this outpoint
				outPoint.Spent = true
				err = w.storeOutPoint(outPoint)
				if err != nil {
					return nil, err
				}

				continue
			}
			ret = append(ret, outPoint)
			sum += outPoint.ColorValue
			if sum >= needed {
				break
			}
		}
	}
	if sum < needed {
		str := fmt.Sprintf("Need %d value, only have %d value", needed, sum)
		return nil, errors.New(str)
	}

	return ret, nil
}

func (w *Wallet) AllColors() (map[*ColorDefinition]ColorId, error) {
	maxColorId, err := w.CurrentId(colorIdKey)
	if err != nil {
		return nil, err
	}
	numColors := DeserializeUint32(maxColorId) - 1
	cds := make(map[*ColorDefinition]ColorId, int(numColors))

	err = w.ColorDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(colorDefinitionBucketName)
		c := b.Cursor()
		i := 0
		for k, v := c.First(); k != nil; k, v = c.Next() {
			cd, err := NewColorDefinitionFromStr(string(k) + ":0")
			if err != nil {
				return err
			}
			cds[cd] = v
			i++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return cds, nil
}

func (w *Wallet) ColorBalance(cd *ColorDefinition) (*ColorValue, error) {
	colorId, err := w.FetchColorId(cd)
	outPoints, err := w.allOutPoints()
	if err != nil {
		return nil, err
	}
	sum := ColorValue(0)
	for _, outPoint := range outPoints {
		if bytes.Compare(outPoint.Color, *colorId) == 0 {
			sum += outPoint.ColorValue
		}
	}

	return &sum, nil
}

func (w *Wallet) IssueColor(b *BlockExplorer, kernel ColorKernel, value ColorValue, fee int64) (*ColorDefinition, error) {
	needed := ColorValue(kernel.IssuingSatoshiNeeded(value) + fee)
	ins, err := w.fetchSpendable(b, uncoloredColorId, needed)
	if err != nil {
		return nil, err
	}
	inputs := make([]*btcwire.OutPoint, len(ins))
	for i, in := range ins {
		inputs[i], err = in.OutPoint()
		if err != nil {
			return nil, err
		}
	}

	outAddr, err := w.NewIssuingAddress()
	if err != nil {
		return nil, err
	}
	pkScript, err := btcscript.PayToAddrScript(outAddr)
	if err != nil {
		return nil, err
	}
	outputs := []*ColorOut{&ColorOut{pkScript, value}}
	changeAddr, err := w.NewUncoloredAddress()
	if err != nil {
		return nil, err
	}
	changeScript, err := btcscript.PayToAddrScript(changeAddr)
	if err != nil {
		return nil, err
	}

	tx, err := kernel.IssuingTx(b, inputs, outputs, changeScript, fee)
	if err != nil {
		return nil, err
	}

	// sign everything
	for i, in := range ins {
		err = w.Sign(in.PkScript, tx, i)
		if err != nil {
			return nil, err
		}
	}

	txHash, err := b.PublishTx(tx)
	if err != nil {
		return nil, err
	}
	// mark everything as spent
	for i, in := range ins {
		in.Spent = true
		in.SpendingTx = BigEndianBytes(txHash)
		in.SpendingIndex = uint32(i)
		err = w.storeOutPoint(in)
		if err != nil {
			return nil, err
		}

	}

	// make the new color definition
	genesis := btcwire.NewOutPoint(txHash, 0)
	cd, err := NewColorDefinition(kernel, genesis, 0)
	if err != nil {
		return nil, err
	}
	cid, err := w.FetchOrAddDefinition(cd.String())
	if err != nil {
		return nil, err
	}

	// add outpoint of this tx
	outPointId, err := w.NewOutPointId()
	if err != nil {
		return nil, err
	}
	colorOutPoint := &ColorOutPoint{
		Id:         outPointId,
		Tx:         BigEndianBytes(&genesis.Hash),
		Index:      genesis.Index,
		Value:      uint64(tx.TxOut[genesis.Index].Value),
		Color:      cid,
		ColorValue: value,
		PkScript:   pkScript,
	}
	err = w.storeOutPoint(colorOutPoint)
	if err != nil {
		return nil, err
	}

	return cd, nil
}

func (w *Wallet) Send(b *BlockExplorer, cd *ColorDefinition, addrMap map[btcutil.Address]ColorValue, fee int64) (*btcwire.MsgTx, error) {
	colorId, err := w.FetchColorId(cd)
	if err != nil {
		return nil, err
	}
	needed := ColorValue(0)
	var outputs []*ColorOut
	for addr, cv := range addrMap {
		needed += cv
		pkScript, err := btcscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, &ColorOut{pkScript, cv})
	}
	coloredInputs, err := w.fetchSpendable(b, *colorId, needed)
	if err != nil {
		return nil, err
	}
	var inputs []*ColorIn
	inSum := ColorValue(0)
	for _, ci := range coloredInputs {
		inSum += ci.ColorValue
		colorIn, err := ci.ColorIn()
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, colorIn)
	}
	// see if we need colored change
	if inSum > needed {
		addr, err := w.NewColorAddress(cd)
		if err != nil {
			return nil, err
		}
		pkScript, err := btcscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, &ColorOut{pkScript, inSum - needed})
	}
	uncoloredInputs, err := w.fetchSpendable(b, uncoloredColorId, ColorValue(fee))
	if err != nil {
		return nil, err
	}
	for _, ui := range uncoloredInputs {
		outPoint, err := ui.OutPoint()
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, &ColorIn{outPoint, ColorValue(0)})
	}
	changeAddr, err := w.NewUncoloredAddress()
	if err != nil {
		return nil, err
	}
	changeScript, err := btcscript.PayToAddrScript(changeAddr)
	if err != nil {
		return nil, err
	}

	tx, err := cd.TransferringTx(b, inputs, outputs, changeScript, fee, false)

	if err != nil {
		return nil, err
	}

	// sign everything
	for i, in := range coloredInputs {
		err = w.Sign(in.PkScript, tx, i)
		if err != nil {
			return nil, err
		}
	}
	for i, in := range uncoloredInputs {
		err = w.Sign(in.PkScript, tx, i+len(coloredInputs))
		if err != nil {
			return nil, err
		}
	}

	txHash, err := b.PublishTx(tx)
	if err != nil {
		return nil, err
	}
	// mark everything as spent
	for i, in := range coloredInputs {
		in.Spent = true
		in.SpendingTx = BigEndianBytes(txHash)
		in.SpendingIndex = uint32(i)
		err = w.storeOutPoint(in)
		if err != nil {
			return nil, err
		}

	}
	for i, in := range uncoloredInputs {
		in.Spent = true
		in.SpendingTx = BigEndianBytes(txHash)
		in.SpendingIndex = uint32(i + len(coloredInputs))
		err = w.storeOutPoint(in)
		if err != nil {
			return nil, err
		}
	}
	return tx, nil
}

func (cop *ColorOutPoint) IsUncolored() bool {
	return bytes.Compare(cop.Color, uncoloredColorId) == 0
}

func (cop *ColorOutPoint) OutPoint() (*btcwire.OutPoint, error) {
	shaHash, err := NewShaHash(cop.Tx)
	if err != nil {
		return nil, err
	}
	return btcwire.NewOutPoint(shaHash, cop.Index), nil
}

func (cop *ColorOutPoint) ColorIn() (*ColorIn, error) {
	if cop.IsUncolored() {
		return nil, errors.New("Cannot make an uncolored out point into a color in")
	}
	outPoint, err := cop.OutPoint()
	if err != nil {
		return nil, err
	}

	return &ColorIn{outPoint, cop.ColorValue}, nil
}
