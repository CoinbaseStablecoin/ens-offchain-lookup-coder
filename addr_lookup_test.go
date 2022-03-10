package coder

import (
	"testing"
	"time"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	dnsname "github.com/petejkim/ens-dnsname"
	"github.com/stretchr/testify/require"
)

func prepareAddrLookup(t *testing.T) (senderAddress common.Address, requestData []byte, lookup *AddrLookup) {
	sender, err := randomAddress()
	require.Nil(t, err)

	name := randomName()
	node, err := namehash.NameHash(name)
	require.Nil(t, err)

	addrInputs, err := abi.IAddrResolver.Methods["addr"].Inputs.Pack(node)
	require.Nil(t, err)

	addrCallData := make([]byte, len(addrInputs)+4)
	copy(addrCallData, abi.IAddrResolver.Methods["addr"].ID)
	copy(addrCallData[4:], addrInputs)

	dn, err := dnsname.Encode(name)
	require.Nil(t, err)

	resolveInputs, err := abi.IResolverService.Methods["resolve"].Inputs.Pack(dn, addrCallData)
	require.Nil(t, err)

	requestData = make([]byte, len(resolveInputs)+4)
	copy(requestData, abi.IResolverService.Methods["resolve"].ID)
	copy(requestData[4:], resolveInputs)

	lookup, err = NewAddrLookup(name, addrInputs, *sender, requestData)
	require.Nil(t, err)
	require.Equal(t, name, lookup.Name())

	return *sender, requestData, lookup
}

func TestAddrLookupEncodeResult(t *testing.T) {
	sender, requestData, lookup := prepareAddrLookup(t)

	resultAddress, err := randomAddress()
	require.Nil(t, err)

	result := resultAddress.Bytes()
	expires := uint64(time.Now().Unix() + 300)

	resultData, hash, err := lookup.EncodeResult(result, expires)
	require.Nil(t, err)

	decoded, err := abi.IAddrResolver.Methods["addr"].Outputs.Unpack(resultData)
	require.Nil(t, err)

	require.Equal(t, *resultAddress, decoded[0])

	require.Equal(t, hashResult(sender, expires, requestData, resultData), hash)
}

func TestAddrLookupEncodeResultInvalidAddress(t *testing.T) {
	_, _, lookup := prepareAddrLookup(t)

	invalidAddress1, err := randomBytes(19)
	require.Nil(t, err)

	invalidAddress2, err := randomBytes(21)
	require.Nil(t, err)

	invalidAddress3 := []byte{}

	expires := uint64(time.Now().Unix() + 300)

	for _, invalidAddress := range [][]byte{invalidAddress1, invalidAddress2, invalidAddress3} {
		resultData, hash, err := lookup.EncodeResult(invalidAddress, expires)
		require.Nil(t, resultData)
		require.Nil(t, hash)
		require.EqualError(t, err, "address must be 20 bytes long")
	}
}
