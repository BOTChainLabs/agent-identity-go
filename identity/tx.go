package identity

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const defaultGasBufferBPS uint64 = 12000

// TxParams are shared gas overrides for write methods.
type TxParams struct {
	GasLimit     uint64
	GasBufferBPS uint64
}

func (c *Client) transactOpts(ctx context.Context, signer Signer, gasLimit uint64, estimate func(context.Context, common.Address) (uint64, error)) (*bind.TransactOpts, error) {
	if signer == nil {
		return nil, fmt.Errorf("agent-identity-go: signer is nil")
	}
	from := signer.Address()
	nonce, err := c.eth.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, wrapTx("pending nonce", err)
	}

	chainID := big.NewInt(c.chain.ChainID)
	opts := &bind.TransactOpts{
		From:  from,
		Nonce: new(big.Int).SetUint64(nonce),
		Signer: func(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if addr != from {
				return nil, fmt.Errorf("agent-identity-go: signer address mismatch")
			}
			signed, err := signer.SignTx(tx, chainID)
			if err != nil {
				return nil, wrapTx("sign tx", err)
			}
			return signed, nil
		},
		Context: ctx,
	}

	if gasLimit > 0 {
		opts.GasLimit = gasLimit
	} else if estimate != nil {
		estimated, err := estimate(ctx, from)
		if err != nil {
			return nil, err
		}
		opts.GasLimit = estimated
	}

	head, err := c.eth.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, wrapTx("latest header", err)
	}
	if head.BaseFee != nil && head.BaseFee.Sign() > 0 {
		tip, err := c.eth.SuggestGasTipCap(ctx)
		if err != nil {
			return nil, wrapTx("suggest tip", err)
		}
		opts.GasTipCap = tip
		opts.GasFeeCap = new(big.Int).Add(new(big.Int).Mul(head.BaseFee, big.NewInt(2)), tip)
	} else {
		price, err := c.eth.SuggestGasPrice(ctx)
		if err != nil {
			return nil, wrapTx("suggest gas price", err)
		}
		opts.GasPrice = price
	}
	return opts, nil
}

func (c *Client) estimateGas(ctx context.Context, from common.Address, data []byte, bufferBPS uint64) (uint64, error) {
	to := common.HexToAddress(c.chain.IdentityRegistryAddress)
	gas, err := c.eth.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &to, Data: data})
	if err != nil {
		return 0, wrapTx("estimate gas", err)
	}
	return applyGasBuffer(gas, bufferBPS), nil
}

func (c *Client) waitSuccess(ctx context.Context, tx *types.Transaction, op string) (*types.Receipt, error) {
	hash := tx.Hash().Hex()
	receipt, err := bind.WaitMined(ctx, c.eth, tx)
	if err != nil {
		return nil, wrapTxHash("wait receipt", hash, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, &TxError{
			Kind:   TxErrorRevert,
			Op:     op,
			TxHash: hash,
			Err:    fmt.Errorf("transaction reverted: %s", hash),
		}
	}
	return receipt, nil
}

// WaitMined waits until txHash is mined and returns the receipt (any status).
// Use after a write method returned a TxError / partial result with TxHash —
// do not blindly retry Register, which would mint another identity.
func (c *Client) WaitMined(ctx context.Context, txHash string) (*types.Receipt, error) {
	hash, err := parseExactHash(txHash)
	if err != nil {
		return nil, err
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		receipt, err := c.eth.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, wrapTxHash("wait receipt", hash.Hex(), err)
		}
		select {
		case <-ctx.Done():
			return nil, wrapTxHash("wait receipt", hash.Hex(), ctx.Err())
		case <-ticker.C:
		}
	}
}

// parseExactHash requires a full 32-byte hash (with or without 0x). Unlike
// common.HexToHash, short / overlong inputs are rejected instead of padded or truncated.
func parseExactHash(txHash string) (common.Hash, error) {
	raw := strings.TrimSpace(txHash)
	raw = strings.TrimPrefix(strings.TrimPrefix(raw, "0x"), "0X")
	if len(raw) != 64 {
		return common.Hash{}, fmt.Errorf("agent-identity-go: invalid tx hash %q: want 32 bytes (64 hex chars)", txHash)
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		return common.Hash{}, fmt.Errorf("agent-identity-go: invalid tx hash %q: %w", txHash, err)
	}
	var hash common.Hash
	copy(hash[:], b)
	return hash, nil
}

func (c *Client) requireOwner(ctx context.Context, agentID *big.Int, signer Signer) error {
	if agentID == nil {
		return fmt.Errorf("agent-identity-go: agentID is nil")
	}
	owner, err := c.OwnerOf(ctx, agentID)
	if err != nil {
		return err
	}
	if owner != signer.Address() {
		return &TxError{
			Kind: TxErrorRevert,
			Op:   "owner check",
			Err:  fmt.Errorf("signer %s is not NFT owner %s for agentId %s", signer.Address().Hex(), owner.Hex(), agentID.String()),
		}
	}
	return nil
}

func applyGasBuffer(gas, bps uint64) uint64 {
	if bps == 0 {
		bps = defaultGasBufferBPS
	}
	return (gas*bps + 9999) / 10000
}
