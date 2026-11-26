package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/BOTChainLabs/agent-identity-go/chains"
	"github.com/BOTChainLabs/agent-identity-go/contracts/identityregistry"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client is the Direct SDK surface for BOT Chain IdentityRegistry.
type Client struct {
	chain    chains.Config
	eth      *ethclient.Client
	registry *identityregistry.IdentityRegistry
	ownsEth  bool
}

type options struct {
	rpcURL string
	eth    *ethclient.Client
}

// Option configures NewClient.
type Option func(*options)

// WithRPCURL overrides the chain preset RPC endpoint.
func WithRPCURL(rpcURL string) Option {
	return func(o *options) {
		o.rpcURL = rpcURL
	}
}

// WithEthClient injects an already-dialed ethclient. When set, the Client does
// not close it on Close. Useful for tests and shared connection pools.
func WithEthClient(eth *ethclient.Client) Option {
	return func(o *options) {
		o.eth = eth
	}
}

// NewClient builds a client for a supported chain preset (677 or 968).
// By default it dials the preset RPC URL; use WithRPCURL or WithEthClient to
// customize connectivity. The RPC endpoint's eth_chainId must match chainID.
func NewClient(chainID int64, opts ...Option) (*Client, error) {
	cfg, err := chains.ByChainID(chainID)
	if err != nil {
		return nil, &ErrUnsupportedChain{ChainID: chainID}
	}

	o := options{rpcURL: cfg.RPCURL}
	for _, opt := range opts {
		opt(&o)
	}

	var (
		eth     *ethclient.Client
		ownsEth bool
	)
	switch {
	case o.eth != nil:
		eth = o.eth
	default:
		if o.rpcURL == "" {
			return nil, fmt.Errorf("agent-identity-go: empty RPC URL for chain_id %d", chainID)
		}
		eth, err = ethclient.Dial(o.rpcURL)
		if err != nil {
			return nil, fmt.Errorf("agent-identity-go: dial RPC: %w", err)
		}
		ownsEth = true
	}

	if err := verifyChainID(eth, chainID); err != nil {
		if ownsEth {
			eth.Close()
		}
		return nil, err
	}

	addr := common.HexToAddress(cfg.IdentityRegistryAddress)
	registry, err := identityregistry.NewIdentityRegistry(addr, eth)
	if err != nil {
		if ownsEth {
			eth.Close()
		}
		return nil, fmt.Errorf("agent-identity-go: bind IdentityRegistry: %w", err)
	}

	return &Client{
		chain:    cfg,
		eth:      eth,
		registry: registry,
		ownsEth:  ownsEth,
	}, nil
}

func verifyChainID(eth *ethclient.Client, want int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	got, err := eth.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("agent-identity-go: eth_chainId: %w", err)
	}
	if got.Int64() != want {
		return fmt.Errorf("agent-identity-go: RPC chain_id %s does not match requested %d", got.String(), want)
	}
	return nil
}

// Close releases the underlying RPC connection when this Client dialed it.
func (c *Client) Close() {
	if c == nil || !c.ownsEth || c.eth == nil {
		return
	}
	c.eth.Close()
	c.ownsEth = false
}

// Chain returns the active chain preset.
func (c *Client) Chain() chains.Config {
	return c.chain
}

// RegistryAddress is a convenience accessor for the IdentityRegistry proxy.
func (c *Client) RegistryAddress() string {
	return c.chain.IdentityRegistryAddress
}

// EthClient exposes the underlying ethclient for advanced callers.
func (c *Client) EthClient() *ethclient.Client {
	return c.eth
}
