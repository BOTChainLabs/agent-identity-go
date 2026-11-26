package chains

import "testing"

func TestByChainID(t *testing.T) {
	t.Parallel()

	mainnet, err := ByChainID(MainnetChainID)
	if err != nil {
		t.Fatalf("mainnet: %v", err)
	}
	if mainnet.IdentityRegistryAddress != Mainnet.IdentityRegistryAddress {
		t.Fatalf("mainnet registry = %s", mainnet.IdentityRegistryAddress)
	}
	if mainnet.RPCURL == "" {
		t.Fatal("mainnet RPC empty")
	}

	testnet, err := ByChainID(TestnetChainID)
	if err != nil {
		t.Fatalf("testnet: %v", err)
	}
	if testnet.IdentityRegistryAddress != "0xec8fFbC3c9A34AdDCbB3A14F91db2bd26A8b99c0" {
		t.Fatalf("testnet registry = %s", testnet.IdentityRegistryAddress)
	}

	if _, err := ByChainID(1); err == nil {
		t.Fatal("expected error for unsupported chain")
	}
}
