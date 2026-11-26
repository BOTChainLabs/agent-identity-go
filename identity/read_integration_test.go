package identity_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/BOTChainLabs/agent-identity-go/chains"
	"github.com/BOTChainLabs/agent-identity-go/identity"
)

// Optional live read against Bohr testnet. Set AGENT_IDENTITY_AGENT_ID to a
// known agentId (decimal). Skip when unset or when -short.
func TestIntegrationReadTestnet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	rawID := os.Getenv("AGENT_IDENTITY_AGENT_ID")
	if rawID == "" {
		t.Skip("set AGENT_IDENTITY_AGENT_ID to run live testnet read")
	}
	agentID, ok := new(big.Int).SetString(rawID, 10)
	if !ok {
		t.Fatalf("invalid AGENT_IDENTITY_AGENT_ID=%q", rawID)
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = chains.Testnet.RPCURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := identity.NewClient(chains.TestnetChainID, identity.WithRPCURL(rpcURL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	version, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	if version == "" {
		t.Fatal("empty version")
	}
	t.Logf("getVersion=%s", version)

	owner, err := client.OwnerOf(ctx, agentID)
	if err != nil {
		t.Fatalf("OwnerOf: %v", err)
	}
	if (owner == [20]byte{}) {
		t.Fatal("zero owner")
	}
	t.Logf("ownerOf(%s)=%s", agentID, owner.Hex())

	uri, err := client.TokenURI(ctx, agentID)
	if err != nil {
		t.Fatalf("TokenURI: %v", err)
	}
	t.Logf("tokenURI=%s", uri)

	ref, err := client.GetIdentity(ctx, agentID)
	if err != nil {
		t.Fatalf("GetIdentity: %v", err)
	}
	if ref.ChainID != chains.TestnetChainID {
		t.Fatalf("chain_id=%d", ref.ChainID)
	}
	if ref.AgentID.Cmp(agentID) != 0 {
		t.Fatalf("agent_id=%s", ref.AgentID)
	}
}
