package identity

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/BOTChainLabs/agent-identity-go/chains"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestPrivateKeySignerAddress(t *testing.T) {
	t.Parallel()

	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	signer, err := NewPrivateKeySigner(key)
	if err != nil {
		t.Fatal(err)
	}
	want := crypto.PubkeyToAddress(key.PublicKey)
	if signer.Address() != want {
		t.Fatalf("addr=%s want=%s", signer.Address().Hex(), want.Hex())
	}
}

func TestPackRegisterSelectors(t *testing.T) {
	t.Parallel()

	c := &Client{}
	noURI, err := c.packRegister(RegisterParams{})
	if err != nil {
		t.Fatal(err)
	}
	if common.Bytes2Hex(noURI[:4]) != "1aa3a008" {
		t.Fatalf("register() selector=%x", noURI[:4])
	}

	withURI, err := c.packRegister(RegisterParams{AgentURI: "ipfs://x"})
	if err != nil {
		t.Fatal(err)
	}
	if common.Bytes2Hex(withURI[:4]) != "f2c298be" {
		t.Fatalf("register(string) selector=%x", withURI[:4])
	}

	withMeta, err := c.packRegister(RegisterParams{
		AgentURI: "ipfs://x",
		Metadata: []MetadataEntry{{Key: "k", Value: []byte("v")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if common.Bytes2Hex(withMeta[:4]) != "8ea42286" {
		t.Fatalf("register(string,metadata) selector=%x", withMeta[:4])
	}
}

func TestClassifyTxErr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		msg  string
		kind TxErrorKind
	}{
		{"user rejected the request", TxErrorRejected},
		{"execution reverted: foo", TxErrorRevert},
		{"dial tcp timeout", TxErrorRPC},
	}
	for _, tc := range cases {
		got := classifyTxErr(errors.New(tc.msg))
		if got != tc.kind {
			t.Fatalf("%q -> %v want %v", tc.msg, got, tc.kind)
		}
	}
}

func TestApplyGasBuffer(t *testing.T) {
	t.Parallel()

	if got := applyGasBuffer(1000, 0); got != 1200 {
		t.Fatalf("default buffer=%d", got)
	}
	if got := applyGasBuffer(1000, 15000); got != 1500 {
		t.Fatalf("custom buffer=%d", got)
	}
}

func TestWrapTxPreservesKind(t *testing.T) {
	t.Parallel()

	err := wrapTx("register", errors.New("execution reverted: x"))
	var txErr *TxError
	if !errors.As(err, &txErr) {
		t.Fatalf("got %T", err)
	}
	if txErr.Kind != TxErrorRevert {
		t.Fatalf("kind=%v", txErr.Kind)
	}
	if !strings.Contains(txErr.Error(), "register") {
		t.Fatalf("msg=%s", txErr.Error())
	}
}

func TestWrapTxHashSurfacesForRecovery(t *testing.T) {
	t.Parallel()

	err := wrapTxHash("wait receipt", "0xabc", context.DeadlineExceeded)
	var txErr *TxError
	if !errors.As(err, &txErr) {
		t.Fatalf("got %T", err)
	}
	if txErr.TxHash != "0xabc" {
		t.Fatalf("hash=%s", txErr.TxHash)
	}
	if !strings.Contains(txErr.Error(), "0xabc") {
		t.Fatalf("msg=%s", txErr.Error())
	}
	if !strings.Contains(txErr.Error(), "query receipt before retrying Register") {
		t.Fatalf("msg=%s", txErr.Error())
	}
}

func TestParseExactHash(t *testing.T) {
	t.Parallel()

	full := "0x" + strings.Repeat("ab", 32)
	hash, err := parseExactHash(full)
	if err != nil {
		t.Fatal(err)
	}
	if hash.Hex() != full {
		t.Fatalf("got %s", hash.Hex())
	}

	if _, err := parseExactHash("0x1"); err == nil {
		t.Fatal("expected short hash error")
	}
	if _, err := parseExactHash("0x" + strings.Repeat("ab", 33)); err == nil {
		t.Fatal("expected long hash error")
	}
	if _, err := parseExactHash("0x" + strings.Repeat("zz", 32)); err == nil {
		t.Fatal("expected non-hex error")
	}
}

func TestParseRegisteredNilReceipt(t *testing.T) {
	t.Parallel()

	c := &Client{chain: chains.Testnet}
	if _, err := c.parseRegistered(nil); err == nil || !strings.Contains(err.Error(), "nil") {
		t.Fatalf("got %v", err)
	}
}

func TestRefAfterRegisterShape(t *testing.T) {
	t.Parallel()

	// Ensure RegisterResult JSON-facing fields stay stable.
	res := RegisterResult{
		Ref: IdentityRef{
			ChainID:         968,
			RegistryAddress: "0xec8fFbC3c9A34AdDCbB3A14F91db2bd26A8b99c0",
			AgentID:         big.NewInt(7),
		},
		TxHash: "0xabc",
	}
	if res.Ref.AgentID.Cmp(big.NewInt(7)) != 0 {
		t.Fatal(res.Ref.AgentID)
	}
}
