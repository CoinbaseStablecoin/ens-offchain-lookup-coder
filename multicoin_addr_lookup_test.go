package coder

import (
	"math/big"
	mathrand "math/rand"
	"testing"
	"time"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	dnsname "github.com/petejkim/ens-dnsname"
	"github.com/stretchr/testify/require"
)

func prepareMulticoinAddrLookup(t *testing.T) (senderAddress common.Address, requestData []byte, lookup *MulticoinAddrLookup) {
	sender, err := randomAddress()
	require.Nil(t, err)

	name := randomName()
	node, err := namehash.NameHash(name)
	require.Nil(t, err)

	coinType := big.NewInt(int64(mathrand.Intn(100000)))

	multicoinAddrInputs, err := abi.IMulticoinAddrResolver.Methods["addr"].Inputs.Pack(node, coinType)
	require.Nil(t, err)

	multicoinAddrCallData := make([]byte, len(multicoinAddrInputs)+4)
	copy(multicoinAddrCallData, abi.IMulticoinAddrResolver.Methods["addr"].ID)
	copy(multicoinAddrCallData[4:], multicoinAddrInputs)

	dn, err := dnsname.Encode(name)
	require.Nil(t, err)

	resolveInputs, err := abi.IResolverService.Methods["resolve"].Inputs.Pack(dn, multicoinAddrCallData)
	require.Nil(t, err)

	requestData = make([]byte, len(resolveInputs)+4)
	copy(requestData, abi.IResolverService.Methods["resolve"].ID)
	copy(requestData[4:], resolveInputs)

	lookup, err = NewMulticoinAddrLookup(name, multicoinAddrInputs, *sender, requestData)
	require.Nil(t, err)
	require.Equal(t, name, lookup.Name())
	require.Equal(t, coinType, lookup.CoinType())

	return *sender, requestData, lookup
}

func TestMulticoinAddrLookupEncodeResult(t *testing.T) {
	sender, requestData, lookup := prepareMulticoinAddrLookup(t)

	result, err := randomBytes(32)
	require.Nil(t, err)

	expires := uint64(time.Now().Unix() + 300)

	resultData, hash, err := lookup.EncodeResult(result, expires)
	require.Nil(t, err)

	decoded, err := abi.IMulticoinAddrResolver.Methods["addr"].Outputs.Unpack(resultData)
	require.Nil(t, err)

	require.Equal(t, result, decoded[0])
	require.Equal(t, hashResult(sender, expires, requestData, resultData), hash)
}
