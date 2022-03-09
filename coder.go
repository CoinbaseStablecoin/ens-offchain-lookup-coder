package coder

import (
	"bytes"
	"encoding/hex"
	"strings"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	dnsname "github.com/petejkim/ens-dnsname"
	"github.com/pkg/errors"
)

func DecodeRequest(sender string, data string) (Lookup, error) {
	senderBytes, err := decodeHex(sender)
	if err != nil || len(senderBytes) != 20 {
		return nil, errors.New("sender is not a valid address")
	}
	senderAddress := common.BytesToAddress(senderBytes)

	requestCallData, err := decodeHex(data)
	if err != nil {
		return nil, errors.New("data is not a valid hex string")
	}

	// check the first four-bytes to ensure that it's calling resolve(bytes,bytes)
	if len(requestCallData) < 4 || !bytes.Equal(requestCallData[0:4], abi.SelectorResolve) {
		return nil, errors.New("data is not a resolve call")
	}

	// decode resolve(bytes,bytes)
	decoded, err := abi.IResolverService.Methods["resolve"].Inputs.Unpack(requestCallData[4:])
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode resolve calldata")
	}

	dnsNameBytes, ok := decoded[0].([]byte)
	if !ok {
		return nil, errors.New("failed to decode resolve calldata")
	}

	lookupCallData, ok := decoded[1].([]byte)
	if !ok {
		return nil, errors.New("failed to decode resolve calldata")
	}

	// decode dns-encoded name
	name, err := dnsname.Decode(dnsNameBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dns-encoded name in the resolve calldata")
	}

	lookupSelector := lookupCallData[0:4]
	lookupInputs := lookupCallData[4:]

	// use the right resolver lookup function based on the selector
	if bytes.Equal(lookupSelector, abi.SelectorAddr) {
		// addr(bytes32)
		return NewAddrLookup(name, lookupInputs, senderAddress, requestCallData)
	} else if bytes.Equal(lookupSelector, abi.SelectorMulticoinAddr) {
		// addr(bytes32,uint256)
		return NewMulticoinAddrLookup(name, lookupInputs, senderAddress, requestCallData)
	} else if bytes.Equal(lookupSelector, abi.SelectorText) {
		// text(bytes32,string)
		return NewTextLookup(name, lookupInputs, senderAddress, requestCallData)
	}

	return nil, errors.Errorf("unsupported lookup: %s", hexutil.Encode(lookupSelector))
}

func EncodeResponse(resultData []byte, expires uint64, signature []byte) (responseData []byte, err error) {
	if len(signature) != 65 {
		return nil, errors.New("signature must be 65 bytes long")
	}

	v := signature[64]
	if v < 27 || v > 34 {
		return nil, errors.New(`invalid "v" value in the signature`)
	}

	if responseData, err = abi.IResolverService.Methods["resolve"].Outputs.Pack(
		resultData, expires, signature,
	); err != nil {
		return nil, errors.Wrap(err, "failed to ABI-encode the result")
	}

	return responseData, nil
}

func decodeHex(str string) ([]byte, error) {
	if strings.HasPrefix(str, "0x") {
		return hex.DecodeString(str[2:])
	}
	return hex.DecodeString(str)
}
