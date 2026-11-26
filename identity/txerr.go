package identity

import (
	"errors"
	"fmt"
	"strings"
)

// TxErrorKind classifies submit / estimate failures for callers.
type TxErrorKind int

const (
	// TxErrorRPC is a transport / node / encoding failure before or after send.
	TxErrorRPC TxErrorKind = iota
	// TxErrorRejected means the signer or wallet refused to sign / send.
	TxErrorRejected
	// TxErrorRevert means the chain estimated or mined a revert.
	TxErrorRevert
)

// TxError wraps a transaction-path failure with a coarse kind.
// When a transaction was already broadcast, TxHash is set so callers can
// recover with WaitMined / TransactionReceipt instead of blindly retrying
// Register (which would mint another identity).
type TxError struct {
	Kind   TxErrorKind
	Op     string
	TxHash string
	Err    error
}

func (e *TxError) Error() string {
	if e.TxHash != "" {
		return fmt.Sprintf("agent-identity-go: %s: %v (tx=%s; query receipt before retrying Register)", e.Op, e.Err, e.TxHash)
	}
	return fmt.Sprintf("agent-identity-go: %s: %v", e.Op, e.Err)
}

func (e *TxError) Unwrap() error { return e.Err }

func wrapTx(op string, err error) error {
	return wrapTxHash(op, "", err)
}

func wrapTxHash(op, txHash string, err error) error {
	if err == nil {
		return nil
	}
	var existing *TxError
	if errors.As(err, &existing) {
		if existing.TxHash == "" && txHash != "" {
			existing.TxHash = txHash
		}
		if existing.Op == "" {
			existing.Op = op
		}
		return existing
	}
	return &TxError{Kind: classifyTxErr(err), Op: op, TxHash: txHash, Err: err}
}

func classifyTxErr(err error) TxErrorKind {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "user rejected"),
		strings.Contains(msg, "user denied"),
		strings.Contains(msg, "rejected by user"),
		strings.Contains(msg, "request rejected"),
		strings.Contains(msg, "cancelled"),
		strings.Contains(msg, "canceled"):
		return TxErrorRejected
	case strings.Contains(msg, "execution reverted"),
		strings.Contains(msg, "invalid opcode"),
		strings.Contains(msg, "revert"):
		return TxErrorRevert
	default:
		return TxErrorRPC
	}
}
