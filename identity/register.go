package identity

import (
	"context"
	"fmt"

	"github.com/BOTChainLabs/agent-identity-go/contracts/identityregistry"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MetadataEntry is on-chain metadata passed to register(uri, metadata).
type MetadataEntry struct {
	Key   string
	Value []byte
}

// RegisterParams configures an IdentityRegistry.register call.
type RegisterParams struct {
	// AgentURI is the initial tokenURI. Empty uses register() with no URI.
	AgentURI string
	// Metadata is optional; when non-empty, uses register(uri, metadata).
	Metadata []MetadataEntry
	TxParams
}

// Register submits register(...), waits for the receipt, and parses agentId
// from the Registered event. The NFT owner is signer.Address().
//
// If the transaction was broadcast but waiting / parsing failed, the returned
// *RegisterResult still carries TxHash (Ref may be empty) together with a
// non-nil error. Callers must WaitMined / ParseRegisteredReceipt before
// retrying Register.
func (c *Client) Register(ctx context.Context, signer Signer, params RegisterParams) (*RegisterResult, error) {
	opts, err := c.transactOpts(ctx, signer, params.GasLimit, func(ctx context.Context, from common.Address) (uint64, error) {
		return c.EstimateRegisterGas(ctx, from, params)
	})
	if err != nil {
		return nil, err
	}

	tx, err := c.sendRegister(opts, params)
	if err != nil {
		return nil, wrapTx("register", err)
	}
	pending := &RegisterResult{TxHash: tx.Hash().Hex()}

	receipt, err := c.waitSuccess(ctx, tx, "register")
	if err != nil {
		return pending, err
	}

	ev, err := c.parseRegistered(receipt)
	if err != nil {
		return pending, wrapTxHash("parse Registered", pending.TxHash, err)
	}

	return &RegisterResult{
		Ref:    c.Ref(ev.AgentId),
		TxHash: pending.TxHash,
	}, nil
}

// ParseRegisteredReceipt extracts IdentityRef from a mined register receipt.
func (c *Client) ParseRegisteredReceipt(receipt *types.Receipt) (IdentityRef, error) {
	ev, err := c.parseRegistered(receipt)
	if err != nil {
		return IdentityRef{}, err
	}
	return c.Ref(ev.AgentId), nil
}

// EstimateRegisterGas estimates gas for register(...) from the given EOA.
func (c *Client) EstimateRegisterGas(ctx context.Context, from common.Address, params RegisterParams) (uint64, error) {
	data, err := c.packRegister(params)
	if err != nil {
		return 0, err
	}
	return c.estimateGas(ctx, from, data, params.GasBufferBPS)
}

func (c *Client) sendRegister(opts *bind.TransactOpts, params RegisterParams) (*types.Transaction, error) {
	switch {
	case len(params.Metadata) > 0:
		entries := make([]identityregistry.IdentityRegistryUpgradeableMetadataEntry, len(params.Metadata))
		for i, m := range params.Metadata {
			entries[i] = identityregistry.IdentityRegistryUpgradeableMetadataEntry{
				MetadataKey:   m.Key,
				MetadataValue: m.Value,
			}
		}
		return c.registry.Register0(opts, params.AgentURI, entries)
	case params.AgentURI != "":
		return c.registry.Register1(opts, params.AgentURI)
	default:
		return c.registry.Register(opts)
	}
}

func (c *Client) packRegister(params RegisterParams) ([]byte, error) {
	abi, err := identityregistry.IdentityRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	// Overloaded register methods are disambiguated as register / register0 / register1
	// after abigen (same mapping as the generated Transactor methods).
	switch {
	case len(params.Metadata) > 0:
		entries := make([]identityregistry.IdentityRegistryUpgradeableMetadataEntry, len(params.Metadata))
		for i, m := range params.Metadata {
			entries[i] = identityregistry.IdentityRegistryUpgradeableMetadataEntry{
				MetadataKey:   m.Key,
				MetadataValue: m.Value,
			}
		}
		return abi.Pack("register0", params.AgentURI, entries)
	case params.AgentURI != "":
		return abi.Pack("register1", params.AgentURI)
	default:
		return abi.Pack("register")
	}
}

func (c *Client) parseRegistered(receipt *types.Receipt) (*identityregistry.IdentityRegistryRegistered, error) {
	if receipt == nil {
		return nil, fmt.Errorf("agent-identity-go: receipt is nil")
	}
	registry := common.HexToAddress(c.chain.IdentityRegistryAddress)
	for _, log := range receipt.Logs {
		if log == nil {
			continue
		}
		if log.Address != registry {
			continue
		}
		ev, err := c.registry.ParseRegistered(*log)
		if err == nil {
			return ev, nil
		}
	}
	return nil, fmt.Errorf("agent-identity-go: Registered event missing from receipt %s for registry %s", receipt.TxHash.Hex(), registry.Hex())
}
