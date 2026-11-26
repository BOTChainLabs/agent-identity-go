package identity

import (
	"context"
	"fmt"
	"math/big"

	"github.com/BOTChainLabs/agent-identity-go/contracts/identityregistry"
	"github.com/ethereum/go-ethereum/common"
)

// UpdateURIParams configures setAgentURI.
type UpdateURIParams struct {
	AgentID *big.Int
	NewURI  string
	TxParams
}

// UpdateURIResult is returned after a successful setAgentURI.
// On wait failure after broadcast, TxHash is still populated with a non-nil error.
type UpdateURIResult struct {
	Ref    IdentityRef `json:"ref"`
	TxHash string      `json:"tx_hash"`
	NewURI string      `json:"new_uri"`
}

// SetMetadataParams configures setMetadata.
type SetMetadataParams struct {
	AgentID *big.Int
	Key     string
	Value   []byte
	TxParams
}

// SetMetadataResult is returned after a successful setMetadata.
// On wait failure after broadcast, TxHash is still populated with a non-nil error.
type SetMetadataResult struct {
	Ref    IdentityRef `json:"ref"`
	TxHash string      `json:"tx_hash"`
	Key    string      `json:"key"`
}

// UpdateURI calls setAgentURI. The signer must be the current NFT owner;
// otherwise a clear TxError is returned before broadcast when possible.
func (c *Client) UpdateURI(ctx context.Context, signer Signer, params UpdateURIParams) (*UpdateURIResult, error) {
	if err := c.requireOwner(ctx, params.AgentID, signer); err != nil {
		return nil, err
	}

	opts, err := c.transactOpts(ctx, signer, params.GasLimit, func(ctx context.Context, from common.Address) (uint64, error) {
		return c.EstimateUpdateURIGas(ctx, from, params)
	})
	if err != nil {
		return nil, err
	}

	tx, err := c.registry.SetAgentURI(opts, params.AgentID, params.NewURI)
	if err != nil {
		return nil, wrapTx("setAgentURI", err)
	}
	pending := &UpdateURIResult{
		Ref:    c.Ref(params.AgentID),
		TxHash: tx.Hash().Hex(),
		NewURI: params.NewURI,
	}
	if _, err := c.waitSuccess(ctx, tx, "setAgentURI"); err != nil {
		return pending, err
	}
	return pending, nil
}

// EstimateUpdateURIGas estimates gas for setAgentURI.
func (c *Client) EstimateUpdateURIGas(ctx context.Context, from common.Address, params UpdateURIParams) (uint64, error) {
	if params.AgentID == nil {
		return 0, fmt.Errorf("agent-identity-go: agentID is nil")
	}
	abi, err := identityregistry.IdentityRegistryMetaData.GetAbi()
	if err != nil {
		return 0, err
	}
	data, err := abi.Pack("setAgentURI", params.AgentID, params.NewURI)
	if err != nil {
		return 0, err
	}
	return c.estimateGas(ctx, from, data, params.GasBufferBPS)
}

// SetMetadata calls setMetadata. The signer must be the current NFT owner.
func (c *Client) SetMetadata(ctx context.Context, signer Signer, params SetMetadataParams) (*SetMetadataResult, error) {
	if err := c.requireOwner(ctx, params.AgentID, signer); err != nil {
		return nil, err
	}

	opts, err := c.transactOpts(ctx, signer, params.GasLimit, func(ctx context.Context, from common.Address) (uint64, error) {
		return c.EstimateSetMetadataGas(ctx, from, params)
	})
	if err != nil {
		return nil, err
	}

	tx, err := c.registry.SetMetadata(opts, params.AgentID, params.Key, params.Value)
	if err != nil {
		return nil, wrapTx("setMetadata", err)
	}
	pending := &SetMetadataResult{
		Ref:    c.Ref(params.AgentID),
		TxHash: tx.Hash().Hex(),
		Key:    params.Key,
	}
	if _, err := c.waitSuccess(ctx, tx, "setMetadata"); err != nil {
		return pending, err
	}
	return pending, nil
}

// EstimateSetMetadataGas estimates gas for setMetadata.
func (c *Client) EstimateSetMetadataGas(ctx context.Context, from common.Address, params SetMetadataParams) (uint64, error) {
	if params.AgentID == nil {
		return 0, fmt.Errorf("agent-identity-go: agentID is nil")
	}
	abi, err := identityregistry.IdentityRegistryMetaData.GetAbi()
	if err != nil {
		return 0, err
	}
	data, err := abi.Pack("setMetadata", params.AgentID, params.Key, params.Value)
	if err != nil {
		return 0, err
	}
	return c.estimateGas(ctx, from, data, params.GasBufferBPS)
}
