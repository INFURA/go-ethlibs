package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/justinwongcn/go-ethlibs/eth"
	"github.com/stretchr/testify/require"
)

type BlockSpecifierTestCase struct {
	Payload       interface{}
	Expected      eth.BlockSpecifier
	Marshalled    []byte
	MarshalledRaw []byte
}

func GetBlockSpecifierTestCases() []BlockSpecifierTestCase {
	return []BlockSpecifierTestCase{
		{
			Payload: "0x0",
			Expected: eth.BlockSpecifier{
				Number: eth.MustQuantity("0x0"),
			},
			Marshalled:    []byte(`{"blockNumber":"0x0"}`),
			MarshalledRaw: []byte(`"0x0"`),
		},
		{
			Payload: "latest",
			Expected: eth.BlockSpecifier{
				Tag: eth.MustTag("latest"),
			},
			Marshalled:    []byte(`"latest"`),
			MarshalledRaw: []byte(`"latest"`),
		},
		{
			Payload: "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
			Expected: eth.BlockSpecifier{
				Hash: eth.MustHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
			},
			Marshalled:    []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`),
			MarshalledRaw: []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`),
		},
		{
			Payload: map[string]interface{}{
				"blockNumber": "0x0",
			},
			Expected: eth.BlockSpecifier{
				Number: eth.MustQuantity("0x0"),
			},
			Marshalled:    []byte(`{"blockNumber":"0x0"}`),
			MarshalledRaw: []byte(`"0x0"`),
		},
		{
			Payload: map[string]interface{}{
				"blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
			},
			Expected: eth.BlockSpecifier{
				Hash:             eth.MustHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
				RequireCanonical: false,
			},
			Marshalled:    []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`),
			MarshalledRaw: []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`),
		},
		{
			Payload: map[string]interface{}{
				"blockHash":        "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
				"requireCanonical": false,
			},
			Expected: eth.BlockSpecifier{
				Hash:             eth.MustHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
				RequireCanonical: false,
			},
			Marshalled:    []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`),
			MarshalledRaw: []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`),
		},
		{
			Payload: map[string]interface{}{
				"blockHash":        "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
				"requireCanonical": true,
			},
			Expected: eth.BlockSpecifier{
				Hash:             eth.MustHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
				RequireCanonical: true,
			},
			Marshalled:    []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":true}`),
			MarshalledRaw: []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`),
		},
	}
}

func TestNewBlockSpecifier(t *testing.T) {
	for _, tc := range GetBlockSpecifierTestCases() {
		spec, err := eth.NewBlockSpecifier(tc.Payload)
		require.Nil(t, err)
		require.Equal(t, tc.Expected, *spec)
	}
}

func TestBlockSpecifierMarshalUnmarshal(t *testing.T) {
	var spec eth.BlockSpecifier
	for _, tc := range GetBlockSpecifierTestCases() {
		jsonPayload, err := json.Marshal(tc.Payload)
		spec.UnmarshalJSON(jsonPayload)
		require.Equal(t, tc.Expected, spec)

		spec.Raw = false
		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, tc.Marshalled, m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, tc.MarshalledRaw, m)
	}
}

func TestGetTag(t *testing.T) {
	payload := "latest"
	spec := eth.MustBlockSpecifier(payload)
	expectedTag := eth.MustTag("latest")

	tag, isTag := spec.GetTag()
	require.True(t, isTag)
	require.Equal(t, expectedTag, tag)

	qty, isNumber := spec.GetQuantity()
	require.False(t, isNumber)
	require.Nil(t, qty)

	hash, isHash := spec.GetHash()
	require.False(t, isHash)
	require.Nil(t, hash)
}

func TestGetQuantity(t *testing.T) {
	payload := "0x0"
	spec := eth.MustBlockSpecifier(payload)
	expectedQuantity := eth.MustQuantity("0x0")

	qty, isNumber := spec.GetQuantity()
	require.True(t, isNumber)
	require.Equal(t, expectedQuantity, qty)

	tag, isTag := spec.GetTag()
	require.False(t, isTag)
	require.Nil(t, tag)

	hash, isHash := spec.GetHash()
	require.False(t, isHash)
	require.Nil(t, hash)
}

func TestGetHash(t *testing.T) {
	payload := map[string]interface{}{
		"blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
	}
	spec := eth.MustBlockSpecifier(payload)
	expectedHash := eth.MustHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")

	hash, isHash := spec.GetHash()
	require.True(t, isHash)
	require.Equal(t, expectedHash, hash)

	qty, isNumber := spec.GetQuantity()
	require.False(t, isNumber)
	require.Nil(t, qty)

	tag, isTag := spec.GetTag()
	require.False(t, isTag)
	require.Nil(t, tag)
}
