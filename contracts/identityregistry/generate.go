// Package identityregistry contains generated bindings for the deployed
// ERC-8004 IdentityRegistryUpgradeable contract.
package identityregistry

// 生成前先更新 erc-8004-contracts/abis/IdentityRegistry.json；禁止手改
// identity_registry.go。绑定版本必须随合约产物一起评审和提交。
//
// Path is relative to this package (same sibling layout as identity-service):
//
//	go generate ./contracts/identityregistry
//
// Use the real contracts-repo ABI (with Solidity struct names). A JSON stub
// stripped of internalType struct names will emit Struct0 and break callers.
//
//go:generate go run github.com/ethereum/go-ethereum/cmd/abigen --abi ../../../erc-8004-contracts/abis/IdentityRegistry.json --pkg identityregistry --type IdentityRegistry --out identity_registry.go
