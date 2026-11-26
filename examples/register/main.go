// Example: register (or read) an Agent identity on Bohr testnet (968).
//
// Env:
//
//	RPC_URL                 optional RPC override
//	PRIVATE_KEY             hex key without 0x — enables Register
//	AGENT_URI               optional initial tokenURI (default empty register())
//	AGENT_IDENTITY_AGENT_ID decimal agentId — read-only ownerOf/tokenURI path
package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/BOTChainLabs/agent-identity-go/identity"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	rpcURL := os.Getenv("RPC_URL")
	var opts []identity.Option
	if rpcURL != "" {
		opts = append(opts, identity.WithRPCURL(rpcURL))
	}

	client, err := identity.NewClient(968, opts...)
	if err != nil {
		fail(err)
	}
	defer client.Close()

	cfg := client.Chain()
	fmt.Printf("chain_id=%d name=%s registry=%s rpc=%s\n",
		cfg.ChainID, cfg.Name, cfg.IdentityRegistryAddress, cfg.RPCURL)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	version, err := client.GetVersion(ctx)
	if err != nil {
		fail(err)
	}
	fmt.Printf("getVersion=%s\n", version)

	if rawKey := strings.TrimSpace(os.Getenv("PRIVATE_KEY")); rawKey != "" {
		register(ctx, client, rawKey)
		return
	}

	rawID := os.Getenv("AGENT_IDENTITY_AGENT_ID")
	if rawID == "" {
		fmt.Println("set PRIVATE_KEY to register, or AGENT_IDENTITY_AGENT_ID to read")
		return
	}
	agentID, ok := new(big.Int).SetString(rawID, 10)
	if !ok {
		fail(fmt.Errorf("invalid AGENT_IDENTITY_AGENT_ID=%q", rawID))
	}
	owner, err := client.OwnerOf(ctx, agentID)
	if err != nil {
		fail(err)
	}
	uri, err := client.TokenURI(ctx, agentID)
	if err != nil {
		fail(err)
	}
	ref := client.Ref(agentID)
	fmt.Printf("ref=(%d, %s, %s)\n", ref.ChainID, ref.RegistryAddress, ref.AgentID)
	fmt.Printf("ownerOf=%s\n", owner.Hex())
	fmt.Printf("tokenURI=%s\n", uri)
}

func register(ctx context.Context, client *identity.Client, rawKey string) {
	rawKey = strings.TrimPrefix(rawKey, "0x")
	key, err := crypto.HexToECDSA(rawKey)
	if err != nil {
		fail(fmt.Errorf("PRIVATE_KEY: %w", err))
	}
	signer, err := identity.NewPrivateKeySigner(key)
	if err != nil {
		fail(err)
	}
	fmt.Printf("signer=%s\n", signer.Address().Hex())

	params := identity.RegisterParams{AgentURI: os.Getenv("AGENT_URI")}
	gas, err := client.EstimateRegisterGas(ctx, signer.Address(), params)
	if err != nil {
		fail(err)
	}
	fmt.Printf("estimate_gas=%d\n", gas)

	result, err := client.Register(ctx, signer, params)
	if err != nil {
		fail(err)
	}
	fmt.Printf("tx=%s\n", result.TxHash)
	fmt.Printf("ref=(%d, %s, %s)\n", result.Ref.ChainID, result.Ref.RegistryAddress, result.Ref.AgentID)

	owner, err := client.OwnerOf(ctx, result.Ref.AgentID)
	if err != nil {
		fail(err)
	}
	fmt.Printf("ownerOf=%s\n", owner.Hex())
	if owner != signer.Address() {
		fail(fmt.Errorf("owner mismatch: chain=%s signer=%s", owner.Hex(), signer.Address().Hex()))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
