package coder

import (
	"fmt"
	mathrand "math/rand"
	"testing"
	"time"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	dnsname "github.com/petejkim/ens-dnsname"
	"github.com/stretchr/testify/require"
)

func prepareTextLookup(t *testing.T) (senderAddress common.Address, requestData []byte, lookup *TextLookup) {
	sender, err := randomAddress()
	require.Nil(t, err)

	name := randomName()
	node, err := namehash.NameHash(name)
	require.Nil(t, err)

	key := fmt.Sprintf("key%d", mathrand.Intn(100000))

	textInputs, err := abi.ITextResolver.Methods["text"].Inputs.Pack(node, key)
	require.Nil(t, err)

	textCallData := make([]byte, len(textInputs)+4)
	copy(textCallData, abi.ITextResolver.Methods["text"].ID)
	copy(textCallData[4:], textInputs)

	dn, err := dnsname.Encode(name)
	require.Nil(t, err)

	resolveInputs, err := abi.IResolverService.Methods["resolve"].Inputs.Pack(dn, textCallData)
	require.Nil(t, err)

	requestData = make([]byte, len(resolveInputs)+4)
	copy(requestData, abi.IResolverService.Methods["resolve"].ID)
	copy(requestData[4:], resolveInputs)

	lookup, err = NewTextLookup(name, textInputs, *sender, requestData)
	require.Nil(t, err)
	require.Equal(t, name, lookup.Name())
	require.Equal(t, key, lookup.Key())

	return *sender, requestData, lookup
}

func TestTextLookupEncodeResult(t *testing.T) {
	sender, requestData, lookup := prepareTextLookup(t)

	result := fmt.Sprintf("result%d", mathrand.Intn(10000))

	expires := uint64(time.Now().Unix() + 300)

	resultData, hash, err := lookup.EncodeResult([]byte(result), expires)
	require.Nil(t, err)

	decoded, err := abi.ITextResolver.Methods["text"].Outputs.Unpack(resultData)
	require.Nil(t, err)

	require.Equal(t, result, decoded[0])
	require.Equal(t, hashResult(sender, expires, requestData, resultData), hash)
}
