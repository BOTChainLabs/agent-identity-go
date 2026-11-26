package identity

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer injects transaction signing without putting a raw private key on the
// Client API. Examples may wrap a keystore or env-loaded key behind this.
type Signer interface {
	Address() common.Address
	SignTx(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
}

// PrivateKeySigner signs Legacy / DynamicFee transactions with an ECDSA key.
// Construct it outside the Client (e.g. from a keystore); do not pass hex keys
// into Client methods.
type PrivateKeySigner struct {
	key  *ecdsa.PrivateKey
	addr common.Address
}

// NewPrivateKeySigner wraps key. The caller retains ownership of the key material.
func NewPrivateKeySigner(key *ecdsa.PrivateKey) (*PrivateKeySigner, error) {
	if key == nil {
		return nil, fmt.Errorf("agent-identity-go: private key is nil")
	}
	return &PrivateKeySigner{
		key:  key,
		addr: crypto.PubkeyToAddress(key.PublicKey),
	}, nil
}

// Address returns the EOA that will own newly registered agents.
func (s *PrivateKeySigner) Address() common.Address {
	return s.addr
}

// SignTx signs tx for chainID (EIP-155).
func (s *PrivateKeySigner) SignTx(tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	if chainID == nil {
		return nil, fmt.Errorf("agent-identity-go: chainID is nil")
	}
	signer := types.LatestSignerForChainID(chainID)
	return types.SignTx(tx, signer, s.key)
}
