package coder

import (
	"bytes"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var _ Lookup = (*TextLookup)(nil)

type TextLookup struct {
	name          string
	senderAddress common.Address
	requestData   []byte
	key           string
}

func NewTextLookup(name string, lookupInputs []byte, senderAddress common.Address, requestData []byte) (*TextLookup, error) {
	nh, err := namehash.NameHash(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namehash")
	}

	decoded, err := abi.ITextResolver.Methods["text"].Inputs.Unpack(lookupInputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode lookup inputs")
	}

	node, ok := decoded[0].([32]byte) // bytes32
	if !ok {
		return nil, errors.New(`failed to decode "node" in lookup inputs`)
	}

	key, ok := decoded[1].(string) // string
	if !ok {
		return nil, errors.New(`failed to decode "key" in lookup inputs`)
	}

	if !bytes.Equal(node[:], nh[:]) {
		return nil, errors.New("name hash does not match the lookup input")
	}

	return &TextLookup{name, senderAddress, requestData, key}, nil
}

func (l *TextLookup) Name() string {
	return l.name
}

func (l *TextLookup) Key() string {
	return l.key
}

func (l *TextLookup) EncodeResult(result []byte, expires uint64) (encodedResult []byte, hash []byte, err error) {
	if encodedResult, err = abi.ITextResolver.Methods["text"].Outputs.Pack(
		string(result), // string
	); err != nil {
		return nil, nil, errors.Wrap(err, "failed to ABI-encode the result")
	}

	hash = hashResult(l.senderAddress, expires, l.requestData, encodedResult)

	return encodedResult, hash, nil
}
