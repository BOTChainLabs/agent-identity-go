package identity

import "fmt"

// ErrUnsupportedChain is returned when chain_id is not 677 or 968.
type ErrUnsupportedChain struct {
	ChainID int64
}

func (e *ErrUnsupportedChain) Error() string {
	return fmt.Sprintf("unsupported chain_id %d: only 677 and 968 are configured", e.ChainID)
}
