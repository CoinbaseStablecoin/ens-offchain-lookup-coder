package coder

import (
	"crypto/rand"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"testing"
	"time"

	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/abi"
	"github.com/CoinbaseStablecoin/ens-offchain-lookup-coder/pkg/namehash"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	dnsname "github.com/petejkim/ens-dnsname"
	"github.com/stretchr/testify/require"
)

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func randomAddress() (*common.Address, error) {
	b, err := randomBytes(20)
	if err != nil {
		return nil, err
	}
	addr := common.BytesToAddress(b)
	return &addr, nil
}

func randomName() string {
	return fmt.Sprintf("pete%d.cbdev%d.eth", mathrand.Intn(10000), mathrand.Intn(10000))
}

func makeExpires() uint64 {
	return uint64(time.Now().Unix() + 300)
}

func TestDecodeRequestAddrLookup(t *testing.T) {
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

	resolveCallData := make([]byte, len(resolveInputs)+4)
	copy(resolveCallData, abi.IResolverService.Methods["resolve"].ID)
	copy(resolveCallData[4:], resolveInputs)

	req, err := DecodeRequest(sender.Hex(), hexutil.Encode(resolveCallData))
	require.Nil(t, err)

	lookup, ok := req.(*AddrLookup)
	require.True(t, ok, "expected the decoded lookup to be a AddrLookup")

	require.Equal(t, name, lookup.Name())
	require.Equal(t, *sender, lookup.senderAddress)
	require.Equal(t, resolveCallData, lookup.requestData)
}

func TestDecodeRequestMulticoinAddrLookup(t *testing.T) {
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

	resolveCallData := make([]byte, len(resolveInputs)+4)
	copy(resolveCallData, abi.IResolverService.Methods["resolve"].ID)
	copy(resolveCallData[4:], resolveInputs)

	req, err := DecodeRequest(sender.Hex(), hexutil.Encode(resolveCallData))
	require.Nil(t, err)

	lookup, ok := req.(*MulticoinAddrLookup)
	require.True(t, ok, "expected the decoded lookup to be a MulticoinAddrLookup")

	require.Equal(t, name, lookup.Name())
	require.Equal(t, coinType, lookup.CoinType())
	require.Equal(t, *sender, lookup.senderAddress)
	require.Equal(t, resolveCallData, lookup.requestData)
}

func TestDecodeRequestTextLookup(t *testing.T) {
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

	resolveCallData := make([]byte, len(resolveInputs)+4)
	copy(resolveCallData, abi.IResolverService.Methods["resolve"].ID)
	copy(resolveCallData[4:], resolveInputs)

	req, err := DecodeRequest(sender.Hex(), hexutil.Encode(resolveCallData))
	require.Nil(t, err)

	lookup, ok := req.(*TextLookup)
	require.True(t, ok, "expected the decoded lookup to be a MulticoinAddrLookup")

	require.Equal(t, name, lookup.Name())
	require.Equal(t, key, lookup.Key())
	require.Equal(t, *sender, lookup.senderAddress)
	require.Equal(t, resolveCallData, lookup.requestData)
}

func TestDecodeRequestInvalidAddress(t *testing.T) {
	req, err := DecodeRequest("0xcafebabe", "0x")
	require.Nil(t, req)
	require.EqualError(t, err, "sender is not a valid address")
}

func TestDecodeRequestNonHexData(t *testing.T) {
	sender, err := randomAddress()
	require.Nil(t, err)

	req, err := DecodeRequest(sender.Hex(), "zebra")
	require.Nil(t, req)
	require.EqualError(t, err, "data is not a valid hex string")
}

func TestDecodeRequestEmptyData(t *testing.T) {
	sender, err := randomAddress()
	require.Nil(t, err)

	req, err := DecodeRequest(sender.Hex(), "0x")
	require.Nil(t, req)
	require.EqualError(t, err, "data is not a resolve call")
}

func TestDecodeRequestMalformedData(t *testing.T) {
	sender, err := randomAddress()
	require.Nil(t, err)

	name := randomName()
	node, err := namehash.NameHash(name)
	require.Nil(t, err)

	// resolve call, but put addr callData in it to make it be malformed
	addrInputs, err := abi.IAddrResolver.Methods["addr"].Inputs.Pack(node)
	require.Nil(t, err)

	malformedCallData := make([]byte, len(addrInputs)+4)
	copy(malformedCallData, abi.IResolverService.Methods["resolve"].ID)
	copy(malformedCallData[4:], addrInputs)

	req, err := DecodeRequest(sender.Hex(), hexutil.Encode(malformedCallData))
	require.Nil(t, req)
	require.Contains(t, err.Error(), "failed to decode resolve calldata")
}

func TestDecodeRequestInvalidDnsEncodedName(t *testing.T) {
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

	// pass raw name instead of dns-encoded name
	resolveInputs, err := abi.IResolverService.Methods["resolve"].Inputs.Pack([]byte(name), addrCallData)
	require.Nil(t, err)

	resolveCallData := make([]byte, len(resolveInputs)+4)
	copy(resolveCallData, abi.IResolverService.Methods["resolve"].ID)
	copy(resolveCallData[4:], resolveInputs)

	req, err := DecodeRequest(sender.Hex(), hexutil.Encode(resolveCallData))
	require.Nil(t, req)
	require.Contains(t, err.Error(), "failed to parse dns-encoded name")
}

func TestDecodeRequestUnsupportedLookup(t *testing.T) {
	sender, err := randomAddress()
	require.Nil(t, err)

	name := randomName()
	node, err := namehash.NameHash(name)
	require.Nil(t, err)

	addrInputs, err := abi.IAddrResolver.Methods["addr"].Inputs.Pack(node)
	require.Nil(t, err)

	randomSelector, err := randomBytes(4)
	require.Nil(t, err)

	addrCallData := make([]byte, len(addrInputs)+4)
	// use a random selector instead of the addr selector
	copy(addrCallData, randomSelector)
	copy(addrCallData[4:], addrInputs)

	dn, err := dnsname.Encode(name)
	require.Nil(t, err)

	resolveInputs, err := abi.IResolverService.Methods["resolve"].Inputs.Pack(dn, addrCallData)
	require.Nil(t, err)

	resolveCallData := make([]byte, len(resolveInputs)+4)
	copy(resolveCallData, abi.IResolverService.Methods["resolve"].ID)
	copy(resolveCallData[4:], resolveInputs)

	req, err := DecodeRequest(sender.Hex(), hexutil.Encode(resolveCallData))
	require.Nil(t, req)
	require.EqualError(t, err, fmt.Sprintf("unsupported lookup: %s", hexutil.Encode(randomSelector)))
}

func TestEncodeResponse(t *testing.T) {
	result, err := randomBytes(32)
	require.Nil(t, err)

	expires := makeExpires()

	mockSignature, err := randomBytes(65)
	require.Nil(t, err)
	mockSignature[64] = 27

	responseData, err := EncodeResponse(result, expires, mockSignature)
	require.Nil(t, err)

	decoded, err := abi.IResolverService.Methods["resolve"].Outputs.Unpack(responseData)
	require.Nil(t, err)

	require.Equal(t, result, decoded[0])
	require.Equal(t, expires, decoded[1])
	require.Equal(t, mockSignature, decoded[2])
}

func TestEncodeInvalidSignature(t *testing.T) {
	result, err := randomBytes(32)
	require.Nil(t, err)

	expires := makeExpires()

	responseData, err := EncodeResponse(result, expires, []byte{})
	require.Nil(t, responseData)
	require.EqualError(t, err, "signature must be 65 bytes long")

	mockSignature, err := randomBytes(65)
	mockSignature[64] = 0
	require.Nil(t, err)

	responseData, err = EncodeResponse(result, expires, mockSignature)
	require.Nil(t, responseData)
	require.EqualError(t, err, `invalid "v" value in the signature`)
}
