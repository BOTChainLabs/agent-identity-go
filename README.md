# agent-identity-go

BOT Chain **ERC-8004 Direct SDK**（Go）。开发者用自有钱包直连 `IdentityRegistry` 注册／查询／更新 Agent 身份，**不经过** `identity-service` HTTP，不需要项目 API Key，也不使用平台 Relayer。

身份完整定位：`(chain_id, registry_address, agent_id)`。

| 链 | chain ID | IdentityRegistry（proxy） |
| --- | --- | --- |
| Bohr 测试网 | `968` | `0xec8fFbC3c9A34AdDCbB3A14F91db2bd26A8b99c0` |
| BOT 主网 | `677` | `0xB43Edfb9C7609cF645e932B2fF20f26F0d4488dE` |

## 安装

```bash
go get github.com/BOTChainLabs/agent-identity-go@v0.1.0
```

```go
import "github.com/BOTChainLabs/agent-identity-go/identity"
```

## 用法

```go
client, err := identity.NewClient(968) // 或 identity.WithRPCURL("https://...")
if err != nil { /* ... */ }
defer client.Close()

version, err := client.GetVersion(ctx)
owner, err := client.OwnerOf(ctx, agentID)
uri, err := client.TokenURI(ctx, agentID)

signer, err := identity.NewPrivateKeySigner(key) // 调用方持有的 *ecdsa.PrivateKey；Client 方法只收 Signer
result, err := client.Register(ctx, signer, identity.RegisterParams{
    AgentURI: "https://example.com/agent.json", // 可空
})
// 若 err != nil 且 result != nil && result.TxHash != ""：先 WaitMined，勿盲重试 Register
// result.Ref = (chain_id, registry, agent_id); result.TxHash

_, err = client.UpdateURI(ctx, signer, identity.UpdateURIParams{
    AgentID: result.Ref.AgentID,
    NewURI:  "https://example.com/agent-v2.json",
})
```

`Client` 交易路径通过 `Signer` 注入签名；`NewPrivateKeySigner` 只是本地辅助构造器。示例可用环境变量，**不要把真实密钥提交进仓库或写进文档正文**。

## 重新生成 binding

需本机已有兄弟目录 `../erc-8004-contracts/abis/IdentityRegistry.json`（须为带 Solidity struct `internalType` 的完整 ABI，与 `identity-service` 同源）：

```bash
go generate ./contracts/identityregistry
```

生成文件禁止手改。

## 快速检查

```bash
# 复制模板后在 .env 填测试网 PRIVATE_KEY（不要提交）
cp .env.example .env

GOPROXY=off GOMODCACHE="$HOME/go/pkg/mod" go test ./...
set -a && source .env && set +a
go run ./examples/register
# 只读：AGENT_IDENTITY_AGENT_ID=1 go run ./examples/register
# 注册+更新：NEW_AGENT_URI=https://example.com/agent-v2.json go run ./examples/update
```
