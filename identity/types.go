package identity

import "math/big"

// IdentityRef is the full on-chain identity locator.
type IdentityRef struct {
	ChainID         int64    `json:"chain_id"`
	RegistryAddress string   `json:"registry_address"`
	AgentID         *big.Int `json:"agent_id"`
}

// RegisterResult is returned after a successful register + receipt parse.
type RegisterResult struct {
	Ref    IdentityRef `json:"ref"`
	TxHash string      `json:"tx_hash"`
}
