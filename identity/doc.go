// Package identity is the BOT Chain ERC-8004 Direct SDK.
//
// It talks to IdentityRegistry on-chain. It does not call identity-service HTTP
// APIs, does not use project API keys, and does not use the platform Relayer.
//
// Identity locator returned by this SDK is always:
//
//	(chain_id, registry_address, agent_id)
package identity
