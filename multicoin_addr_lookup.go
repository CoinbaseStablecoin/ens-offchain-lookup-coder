package coder

import (
	"bytes"
	"math/big"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

var _ Lookup = (*MulticoinAddrLookup)(nil)

type MulticoinAddrLookup struct {
	name          string
	senderAddress common.Address
	requestData   []byte
	coinType      *big.Int
}

func NewMulticoinAddrLookup(name string, lookupInputs []byte, senderAddress common.Address, requestData []byte) (*MulticoinAddrLookup, error) {
	nh, err := namehash.NameHash(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namehash")
	}

	decoded, err := abi.IMulticoinAddrResolver.Methods["addr"].Inputs.Unpack(lookupInputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode lookup inputs")
	}

	node, ok := decoded[0].([32]byte) // bytes32
	if !ok {
		return nil, errors.New(`failed to decode "node" in lookup inputs`)
	}

	coinType, ok := decoded[1].(*big.Int) // uint256
	if !ok {
		return nil, errors.New(`failed to decode "coinType" in lookup inputs`)
	}

	if !bytes.Equal(node[:], nh[:]) {
		return nil, errors.New("name hash does not match the lookup input")
	}

	return &MulticoinAddrLookup{name, senderAddress, requestData, coinType}, nil
}

func (l *MulticoinAddrLookup) Name() string {
	return l.name
}

func (l *MulticoinAddrLookup) CoinType() *big.Int {
	bi := new(big.Int)
	return bi.Add(l.coinType, bi)
}

func (l *MulticoinAddrLookup) EncodeResult(result []byte, expires uint64) (encodedResult []byte, hash []byte, err error) {
	if encodedResult, err = abi.IMulticoinAddrResolver.Methods["addr"].Outputs.Pack(
		result, // bytes
	); err != nil {
		return nil, nil, errors.Wrap(err, "failed to ABI-encode the result")
	}

	hash = hashResult(l.senderAddress, expires, l.requestData, encodedResult)

	return encodedResult, hash, nil
}
