package coder

import (
	"bytes"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var _ Lookup = (*AddrLookup)(nil)

type AddrLookup struct {
	name          string
	senderAddress common.Address
	requestData   []byte
}

func NewAddrLookup(name string, lookupInputs []byte, senderAddress common.Address, requestData []byte) (*AddrLookup, error) {
	nh, err := namehash.NameHash(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namehash")
	}

	decoded, err := abi.IAddrResolver.Methods["addr"].Inputs.Unpack(lookupInputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode lookup inputs")
	}

	node, ok := decoded[0].([32]byte) // bytes32
	if !ok {
		return nil, errors.New(`failed to decode "node" in lookup inputs`)
	}

	if !bytes.Equal(node[:], nh[:]) {
		return nil, errors.New("name hash does not match the lookup input")
	}

	return &AddrLookup{name, senderAddress, requestData}, nil
}

func (l *AddrLookup) Name() string {
	return l.name
}

func (l *AddrLookup) EncodeResult(result []byte, expires uint64) (encodedResult []byte, hash []byte, err error) {
	if len(result) != 20 {
		return nil, nil, errors.New("address must be 20 bytes long")
	}

	if encodedResult, err = abi.IAddrResolver.Methods["addr"].Outputs.Pack(
		common.BytesToAddress(result), // address
	); err != nil {
		return nil, nil, errors.Wrap(err, "failed to ABI-encode the result")
	}

	hash = hashResult(l.senderAddress, expires, l.requestData, encodedResult)

	return encodedResult, hash, nil
}
