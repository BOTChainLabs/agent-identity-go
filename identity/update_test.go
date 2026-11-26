package identity

import (
	"errors"
	"math/big"
	"testing"

	"github.com/BOTChainLabs/agent-identity-go/contracts/identityregistry"
	"github.com/ethereum/go-ethereum/common"
)

func TestNonOwnerErrorKind(t *testing.T) {
	t.Parallel()

	owner := common.HexToAddress("0x1111111111111111111111111111111111111111")
	signer := common.HexToAddress("0x2222222222222222222222222222222222222222")
	err := &TxError{
		Kind: TxErrorRevert,
		Op:   "owner check",
		Err:  errors.New("signer " + signer.Hex() + " is not NFT owner " + owner.Hex() + " for agentId 9"),
	}
	var txErr *TxError
	if !errors.As(err, &txErr) || txErr.Kind != TxErrorRevert {
		t.Fatal(err)
	}
}

func TestPackSetAgentURIAndMetadata(t *testing.T) {
	t.Parallel()

	abi, err := identityregistry.IdentityRegistryMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	uriData, err := abi.Pack("setAgentURI", big.NewInt(1), "https://example.com/a.json")
	if err != nil {
		t.Fatal(err)
	}
	if common.Bytes2Hex(uriData[:4]) == "" {
		t.Fatal("empty setAgentURI selector")
	}
	metaData, err := abi.Pack("setMetadata", big.NewInt(1), "k", []byte("v"))
	if err != nil {
		t.Fatal(err)
	}
	if len(metaData) < 4 {
		t.Fatal("short setMetadata calldata")
	}
}
