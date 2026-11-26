// Package chains provides BOT Chain presets for the ERC-8004 IdentityRegistry.
package chains

import "fmt"

// ChainID values used by this SDK.
const (
	MainnetChainID  int64 = 677
	TestnetChainID  int64 = 968
)

// Config is the on-chain identity locator for one network.
type Config struct {
	ChainID                 int64
	Name                    string
	RPCURL                  string
	ExplorerURL             string
	IdentityRegistryAddress string
}

var (
	// Mainnet is BOT Chain mainnet (677).
	Mainnet = Config{
		ChainID:                 MainnetChainID,
		Name:                    "BOT Chain",
		RPCURL:                  "https://rpc.botchain.ai",
		ExplorerURL:             "https://scan.botchain.ai",
		IdentityRegistryAddress: "0xB43Edfb9C7609cF645e932B2fF20f26F0d4488dE",
	}

	// Testnet is Bohr testnet (968).
	Testnet = Config{
		ChainID:                 TestnetChainID,
		Name:                    "Bohr Testnet",
		RPCURL:                  "https://rpc.bohr.life",
		ExplorerURL:             "https://scan.bohr.life",
		IdentityRegistryAddress: "0xec8fFbC3c9A34AdDCbB3A14F91db2bd26A8b99c0",
	}
)

// ByChainID returns a built-in preset or an error if unsupported.
func ByChainID(chainID int64) (Config, error) {
	switch chainID {
	case MainnetChainID:
		return Mainnet, nil
	case TestnetChainID:
		return Testnet, nil
	default:
		return Config{}, fmt.Errorf("unsupported chain_id %d: only 677 and 968 are configured", chainID)
	}
}
