package main

/*
#include <memory.h>
*/
import "C"

import (
	"encoding/hex"
	pb "github.com/pqabelian/abelian-sdk-go-lib-v2/proto"
	"github.com/pqabelian/abelian-sdk-go-v2/abelian"
	core "github.com/pqabelian/abelian-sdk-go-v2/abelian/crypto"

	"errors"
	"unsafe"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func packToOutData(data []byte, outData []byte) error {
	size := len(data)
	if size > len(outData) {
		return errors.New("outData is too small")
	}

	C.memcpy(unsafe.Pointer(&outData[0]), unsafe.Pointer(&size), 4)
	if size > 0 {
		C.memcpy(unsafe.Pointer(&outData[4]), unsafe.Pointer(&data[0]), C.size_t(size))
	}

	return nil
}

func packToRetData(data []byte) *C.char {
	size := int32(len(data))
	retData := C.malloc(C.size_t(size) + 4)

	C.memcpy(retData, unsafe.Pointer(&size), 4)
	if size > 0 {
		C.memcpy(unsafe.Pointer(uintptr(retData)+4), unsafe.Pointer(&data[0]), C.size_t(size))
	}

	return (*C.char)(retData)
}

func panicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func ignore(args interface{}) {
}

func unmarshalArgs(argsData []byte, args protoreflect.ProtoMessage) {
	err := proto.Unmarshal(argsData, args)
	panicIf(err)
}

func marshalResultAndPackToRetData(result protoreflect.ProtoMessage) *C.char {
	resultData, err := proto.Marshal(result)
	panicIf(err)
	return (*C.char)(packToRetData(resultData))
}

//export GenerateSafeCryptoSeed
func GenerateSafeCryptoSeed(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GenerateSafeCryptoSeedArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	ignore(args)

	// Do real work.
	cryptoSeed, err := core.GenerateSeed(core.CryptoSchemePQRingCTX, core.PrivacyLevel(args.PrivacyLevel))
	panicIf(err)

	// Marshal result and return it.
	result := &pb.GenerateSafeCryptoSeedResult{
		SpendKeyRootSeed:    cryptoSeed.CoinSpendKeySeed(),
		SerialNoKeyRootSeed: cryptoSeed.CoinSerialNumberKeySeed(),
		ViewKeyRootSeed:     cryptoSeed.CoinValueKeySeed(),
		DetectorRootKey:     cryptoSeed.CoinDetectorKey(),
	}
	return marshalResultAndPackToRetData(result)
}

//export GenerateCryptoKeysAndAddressByRootSeeds
func GenerateCryptoKeysAndAddressByRootSeeds(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GenerateCryptoKeysAndAddressByRootSeedsArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	cryptoSeed, err := core.NewRootSeeds(
		core.CryptoSchemePQRingCTX,
		core.PrivacyLevel(args.PrivacyLevel),
		args.GetSpendKeyRootSeed(),
		args.GetSerialNoKeyRootSeed(),
		args.GetViewKeyRootSeed(),
		args.GetDetectorRootKey(),
	)
	panicIf(err)

	// Do real work.
	ckaa, err := core.GenerateCryptoKeysAndAddressByRootSeeds(cryptoSeed)
	panicIf(err)

	// Marshal result and return it.
	result := &pb.GenerateCryptoKeysAndAddressByRootSeedsResult{
		SpendSecretKey:    ckaa.SpendSecretKey,
		SerialNoSecretKey: ckaa.SerialNoSecretKey,
		ViewSecretKey:     ckaa.ViewSecretKey,
		DetectorKey:       ckaa.DetectorKey,
		CryptoAddress:     ckaa.CryptoAddress.Data(),
	}
	return marshalResultAndPackToRetData(result)
}

//export GetAbelAddressFromCryptoAddress
func GetAbelAddressFromCryptoAddress(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GetAbelAddressFromCryptoAddressArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	cryptoAddress, err := core.NewCryptoAddress(args.GetCryptoAddress())
	panicIf(err)
	chainID := args.GetChainID()

	// Do real work.
	abelAddress := abelian.NewAbelAddressFromCryptoAddress(abelian.NetworkID(chainID), cryptoAddress)

	// Marshal result and return it.
	result := &pb.GetAbelAddressFromCryptoAddressResult{
		AbelAddress: abelAddress.Data(),
	}
	return marshalResultAndPackToRetData(result)
}

//export GetCryptoAddressFromAbelAddress
func GetCryptoAddressFromAbelAddress(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GetCryptoAddressFromAbelAddressArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	abelAddress, err := abelian.NewAbelAddress(args.GetAbelAddress())
	panicIf(err)

	// Do real work.
	cryptoAddress := abelAddress.GetCryptoAddress()

	// Marshal result and return it.
	result := &pb.GetCryptoAddressFromAbelAddressResult{
		CryptoAddress: cryptoAddress.Data(),
	}
	return marshalResultAndPackToRetData(result)
}

//export GetShortAbelAddressFromAbelAddress
func GetShortAbelAddressFromAbelAddress(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GetShortAbelAddressFromAbelAddressArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	abelAddress, err := abelian.NewAbelAddress(args.GetAbelAddress())
	panicIf(err)

	// Do real work.
	shortAbelAddress, err := abelian.GetShortAbelAddressFromAbelAddress(abelAddress)
	panicIf(err)

	// Marshal result and return it.
	result := &pb.GetShortAbelAddressFromAbelAddressResult{
		ShortAbelAddress: shortAbelAddress.Data(),
	}
	return marshalResultAndPackToRetData(result)
}

//export CoinReceiveFromTxOutData
func CoinReceiveFromTxOutData(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.CoinReceiveFromTxOutDataArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	txVersion := args.GetTxVersion()
	serializedTxOut := args.GetTxOutData()
	accountPrivacyLevel := args.GetAccountPrivacyLevel()
	coinDetectorRootKey := args.GetCoinDetectorRootKey()
	coinViewSecretRootSeed := args.GetCoinViewSecretRootSeed()

	result := &pb.CoinReceiveFromTxOutDataResult{
		Success:   false,
		CoinValue: 0,
	}

	// Do real work.
	privacyLevel, err := core.GetTxoPrivacyLevel(txVersion, serializedTxOut)
	panicIf(err)
	if core.PrivacyLevel(accountPrivacyLevel) == privacyLevel {
		success, err := core.TxoCoinDetectByCoinDetectorRootKey(txVersion, serializedTxOut, coinDetectorRootKey)
		panicIf(err)
		if success {
			success, coinValue, err := core.TxoCoinReceiveByRootSeeds(txVersion, serializedTxOut, coinViewSecretRootSeed, coinDetectorRootKey)
			panicIf(err)
			// Marshal result and return it.
			result.Success = success
			result.CoinValue = coinValue
		}
	}

	return marshalResultAndPackToRetData(result)
}

//export GenerateRawTxRequest
func GenerateRawTxRequest(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GenerateRawTxRequestArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	txDesc := newTxDescFromArgs(args)

	// Do real work.
	unsignedRawTx, err := abelian.GenerateUnsignedRawTx(txDesc)
	panicIf(err)

	// Marshal result and return it.
	result := &pb.GenerateRawTxRequestResult{
		Data: unsignedRawTx.Data,
	}
	return marshalResultAndPackToRetData(result)
}

func newTxDescFromArgs(args *pb.GenerateRawTxRequestArgs) *abelian.TxDesc {
	// Create txInDescs.
	txInDescMessages := args.GetTxInDescs()
	txInDescs := make([]*abelian.TxInDesc, 0, len(txInDescMessages))
	for _, txInDescMessage := range txInDescMessages {
		txInDesc := newTxInDescFromMessage(txInDescMessage)
		txInDescs = append(txInDescs, txInDesc)
	}

	// Create txOutDescs.
	txOutDescMessages := args.GetTxOutDescs()
	txOutDescs := make([]*abelian.TxOutDesc, 0, len(txOutDescMessages))
	for _, txOutDescMessage := range txOutDescMessages {
		txOutDesc := newTxOutDescFromMessage(txOutDescMessage)
		txOutDescs = append(txOutDescs, txOutDesc)
	}

	// Create txRingBlockDescs.
	txRingBlockDescMessages := args.GetTxRingBlockDescs()
	txRingBlockDescs := make(map[int64]*abelian.TxBlockDesc)
	for _, txRingBlockDescMessage := range txRingBlockDescMessages {
		txRingBlockDesc := newTxRingBlockDescFromMessage(txRingBlockDescMessage)
		txRingBlockDescs[txRingBlockDesc.Height] = txRingBlockDesc
	}

	// Create txFee.
	txFee := args.GetTxFee()

	// Create and return txDesc.
	return abelian.NewTxDesc(txInDescs, txOutDescs, txFee, txRingBlockDescs)
}

func newTxInDescFromMessage(txInDescMessage *pb.TxInDescMessage) *abelian.TxInDesc {
	return &abelian.TxInDesc{
		BlockHeight:      txInDescMessage.GetHeight(),
		BlockID:          hex.EncodeToString(txInDescMessage.GetBlockID()),
		TxVersion:        txInDescMessage.GetTxVersion(),
		TxID:             hex.EncodeToString(txInDescMessage.GetTxID()),
		TxOutIndex:       uint8(txInDescMessage.GetIndex()),
		TxOutData:        txInDescMessage.GetTxOutData(),
		CoinValue:        txInDescMessage.GetValue(),
		CoinSerialNumber: txInDescMessage.GetCoinSerialNumber(),
	}
}

func newTxOutDescFromMessage(txOutDescMessage *pb.TxOutDescMessage) *abelian.TxOutDesc {
	abelAddress, err := abelian.NewAbelAddress(txOutDescMessage.GetAbelAddress())
	panicIf(err)
	return &abelian.TxOutDesc{
		AbelAddress: abelAddress,
		CoinValue:   txOutDescMessage.GetValue(),
	}
}

func newTxRingBlockDescFromMessage(txRingBlockDescMessage *pb.BlockDescMessage) *abelian.TxBlockDesc {
	return &abelian.TxBlockDesc{
		BinData: txRingBlockDescMessage.GetBinData(),
		Height:  txRingBlockDescMessage.GetHeight(),
	}
}

//export GenerateRawTxData
func GenerateRawTxData(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GenerateRawTxDataArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	serializedTxRequest := args.GetSerializedTxRequest()
	privacyLevels := args.GetPrivacyLevels()
	spendKeyRootSeeds := args.GetSpendKeyRootSeeds()
	serialNoKeyRootSeeds := args.GetSerialNoKeyRootSeeds()
	viewKeyRootSeeds := args.GetViewKeyRootSeeds()
	detectorRootKeys := args.GetDetectorRootKeys()
	numCryptoSeeds := len(privacyLevels)
	if len(spendKeyRootSeeds) != numCryptoSeeds ||
		len(spendKeyRootSeeds) != numCryptoSeeds ||
		len(spendKeyRootSeeds) != numCryptoSeeds ||
		len(spendKeyRootSeeds) != numCryptoSeeds {
		panicIf(errors.New("mismatch length for root seeds"))
	}

	signerCryptoSeeds := make([][]byte, numCryptoSeeds)
	for i := 0; i < len(privacyLevels); i++ {
		rootSeeds, err := core.NewRootSeeds(core.CryptoSchemePQRingCTX, core.PrivacyLevel(privacyLevels[i]),
			spendKeyRootSeeds[i],
			serialNoKeyRootSeeds[i],
			viewKeyRootSeeds[i],
			detectorRootKeys[i])
		panicIf(err)
		signerCryptoSeeds[i], err = rootSeeds.Serialize()
		panicIf(err)
	}

	// Do real work.
	serializedTx, txID, err := core.GenerateTransferTransactionByRootSeeds(serializedTxRequest, signerCryptoSeeds)
	panicIf(err)

	// Marshal result and return it.
	result := &pb.GenerateRawTxDataResult{
		Data: serializedTx,
		Txid: txID,
	}
	return marshalResultAndPackToRetData(result)
}

//export GenerateCoinSerialNumber
func GenerateCoinSerialNumber(argsData []byte) *C.char {
	// Unmarshal args.
	args := &pb.GenerateCoinSerialNumberArgs{}
	unmarshalArgs(argsData, args)

	// Prepare data.
	outpoint, err := core.NewOutPointFromTxId(hex.EncodeToString(args.GetTxid()), uint8(args.GetIndex()))
	panicIf(err)

	serialNoSecretRootSeed := args.GetSerialNoSecretRootSeed()

	ringBlockDescs := args.GetRingBlockDescs()
	serializedRingBlockDescs := make([][]byte, len(ringBlockDescs))
	for i := 0; i < len(ringBlockDescs); i++ {
		serializedRingBlockDescs[i] = ringBlockDescs[i].GetBinData()
	}

	// Do real work.
	serialNumbers, err := core.GenerateCoinSerialNumberByRootSeeds([]*core.OutPoint{outpoint}, serializedRingBlockDescs, serialNoSecretRootSeed)
	panicIf(err)

	// Marshal result and return it.
	result := &pb.GenerateCoinSerialNumberResult{
		SerialNumber: serialNumbers[0],
	}
	return marshalResultAndPackToRetData(result)
}
