package identity

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/BOTChainLabs/agent-identity-go/chains"
	"github.com/BOTChainLabs/agent-identity-go/contracts/identityregistry"
)

func TestNewClientUnsupportedChain(t *testing.T) {
	t.Parallel()

	_, err := NewClient(1)
	if err == nil {
		t.Fatal("expected error")
	}
	var unsupported *ErrUnsupportedChain
	if !errors.As(err, &unsupported) || unsupported.ChainID != 1 {
		t.Fatalf("got %v", err)
	}
}

func TestNewClientPresetAddresses(t *testing.T) {
	t.Parallel()

	// Dial is deferred until needed only when EthClient is injected; use a
	// nil-safe path by checking Ref/presets without network via chains package
	// and binding metadata.
	cfg, err := chains.ByChainID(chains.TestnetChainID)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.IdentityRegistryAddress != "0xec8fFbC3c9A34AdDCbB3A14F91db2bd26A8b99c0" {
		t.Fatalf("registry=%s", cfg.IdentityRegistryAddress)
	}

	abi, err := identityregistry.IdentityRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ownerOf", "tokenURI", "getVersion", "register", "setAgentURI"} {
		if _, ok := abi.Methods[name]; !ok {
			t.Fatalf("missing ABI method %s", name)
		}
	}
}

func TestRef(t *testing.T) {
	t.Parallel()

	c := &Client{chain: chains.Testnet}
	ref := c.Ref(big.NewInt(42))
	if ref.ChainID != 968 {
		t.Fatalf("chain=%d", ref.ChainID)
	}
	if ref.RegistryAddress != chains.Testnet.IdentityRegistryAddress {
		t.Fatalf("registry=%s", ref.RegistryAddress)
	}
	if ref.AgentID.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("agent=%s", ref.AgentID)
	}
	ref.AgentID.SetInt64(1)
	if c.Ref(big.NewInt(42)).AgentID.Cmp(big.NewInt(42)) != 0 {
		t.Fatal("Ref must copy agentID")
	}
}

func TestBindingHasRegisteredEvent(t *testing.T) {
	t.Parallel()

	abi, err := identityregistry.IdentityRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := abi.Events["Registered"]; !ok {
		t.Fatal("missing Registered event")
	}
	if !strings.Contains(identityregistry.IdentityRegistryABI, "getVersion") {
		t.Fatal("ABI string missing getVersion")
	}
}
