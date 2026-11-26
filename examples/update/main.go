// Example: register, update URI, and set metadata on Bohr testnet (968).
//
// Env:
//
//	PRIVATE_KEY     required (hex, optional 0x prefix)
//	AGENT_URI       initial URI for register (optional)
//	NEW_AGENT_URI   URI after update (required)
//	METADATA_KEY    optional (default "sdk.smoke")
//	METADATA_VALUE  optional (default "ok")
//	RPC_URL         optional
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BOTChainLabs/agent-identity-go/identity"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	rawKey := strings.TrimSpace(os.Getenv("PRIVATE_KEY"))
	if rawKey == "" {
		fail(fmt.Errorf("PRIVATE_KEY is required"))
	}
	newURI := os.Getenv("NEW_AGENT_URI")
	if newURI == "" {
		fail(fmt.Errorf("NEW_AGENT_URI is required"))
	}
	metaKey := envOr("METADATA_KEY", "sdk.smoke")
	metaValue := []byte(envOr("METADATA_VALUE", "ok"))

	var opts []identity.Option
	if rpc := os.Getenv("RPC_URL"); rpc != "" {
		opts = append(opts, identity.WithRPCURL(rpc))
	}
	client, err := identity.NewClient(968, opts...)
	if err != nil {
		fail(err)
	}
	defer client.Close()

	key, err := crypto.HexToECDSA(strings.TrimPrefix(rawKey, "0x"))
	if err != nil {
		fail(err)
	}
	signer, err := identity.NewPrivateKeySigner(key)
	if err != nil {
		fail(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	reg, err := client.Register(ctx, signer, identity.RegisterParams{AgentURI: os.Getenv("AGENT_URI")})
	if err != nil {
		if reg != nil && reg.TxHash != "" {
			fmt.Fprintf(os.Stderr, "register pending tx=%s: %v\n", reg.TxHash, err)
		}
		fail(err)
	}
	fmt.Printf("registered agent_id=%s tx=%s\n", reg.Ref.AgentID, reg.TxHash)

	upd, err := client.UpdateURI(ctx, signer, identity.UpdateURIParams{
		AgentID: reg.Ref.AgentID,
		NewURI:  newURI,
	})
	if err != nil {
		if upd != nil && upd.TxHash != "" {
			fmt.Fprintf(os.Stderr, "updateURI pending tx=%s: %v\n", upd.TxHash, err)
		}
		fail(err)
	}
	fmt.Printf("updated uri=%s tx=%s\n", upd.NewURI, upd.TxHash)

	got, err := client.TokenURI(ctx, reg.Ref.AgentID)
	if err != nil {
		fail(err)
	}
	fmt.Printf("tokenURI=%s\n", got)
	if got != newURI {
		fail(fmt.Errorf("tokenURI mismatch: got %q want %q", got, newURI))
	}

	meta, err := client.SetMetadata(ctx, signer, identity.SetMetadataParams{
		AgentID: reg.Ref.AgentID,
		Key:     metaKey,
		Value:   metaValue,
	})
	if err != nil {
		if meta != nil && meta.TxHash != "" {
			fmt.Fprintf(os.Stderr, "setMetadata pending tx=%s: %v\n", meta.TxHash, err)
		}
		fail(err)
	}
	fmt.Printf("setMetadata key=%s tx=%s\n", meta.Key, meta.TxHash)

	value, err := client.GetMetadata(ctx, reg.Ref.AgentID, metaKey)
	if err != nil {
		fail(err)
	}
	fmt.Printf("getMetadata=%s\n", string(value))
	if string(value) != string(metaValue) {
		fail(fmt.Errorf("getMetadata mismatch: got %q want %q", string(value), string(metaValue)))
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
