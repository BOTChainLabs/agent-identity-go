package identity

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// Ref assembles the on-chain identity locator for agentID. It does not call RPC.
func (c *Client) Ref(agentID *big.Int) IdentityRef {
	var id *big.Int
	if agentID != nil {
		id = new(big.Int).Set(agentID)
	}
	return IdentityRef{
		ChainID:         c.chain.ChainID,
		RegistryAddress: c.chain.IdentityRegistryAddress,
		AgentID:         id,
	}
}

// OwnerOf returns the NFT owner of agentID (ERC-721 ownerOf).
// A nonexistent agentId reverts on-chain; treat that as "not found".
func (c *Client) OwnerOf(ctx context.Context, agentID *big.Int) (common.Address, error) {
	if agentID == nil {
		return common.Address{}, fmt.Errorf("agent-identity-go: agentID is nil")
	}
	owner, err := c.registry.OwnerOf(&bind.CallOpts{Context: ctx}, agentID)
	if err != nil {
		return common.Address{}, fmt.Errorf("agent-identity-go: ownerOf(%s): %w", agentID.String(), err)
	}
	return owner, nil
}

// TokenURI returns the registration file URI for agentID.
func (c *Client) TokenURI(ctx context.Context, agentID *big.Int) (string, error) {
	if agentID == nil {
		return "", fmt.Errorf("agent-identity-go: agentID is nil")
	}
	uri, err := c.registry.TokenURI(&bind.CallOpts{Context: ctx}, agentID)
	if err != nil {
		return "", fmt.Errorf("agent-identity-go: tokenURI(%s): %w", agentID.String(), err)
	}
	return uri, nil
}

// GetVersion returns the IdentityRegistry contract version string (e.g. "2.0.0").
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	version, err := c.registry.GetVersion(&bind.CallOpts{Context: ctx})
	if err != nil {
		return "", fmt.Errorf("agent-identity-go: getVersion: %w", err)
	}
	return version, nil
}

// GetIdentity verifies agentID exists via ownerOf and returns its IdentityRef.
func (c *Client) GetIdentity(ctx context.Context, agentID *big.Int) (IdentityRef, error) {
	if _, err := c.OwnerOf(ctx, agentID); err != nil {
		return IdentityRef{}, err
	}
	return c.Ref(agentID), nil
}

// GetMetadata returns on-chain metadata bytes for agentID + key.
func (c *Client) GetMetadata(ctx context.Context, agentID *big.Int, key string) ([]byte, error) {
	if agentID == nil {
		return nil, fmt.Errorf("agent-identity-go: agentID is nil")
	}
	value, err := c.registry.GetMetadata(&bind.CallOpts{Context: ctx}, agentID, key)
	if err != nil {
		return nil, fmt.Errorf("agent-identity-go: getMetadata(%s, %q): %w", agentID.String(), key, err)
	}
	return value, nil
}
