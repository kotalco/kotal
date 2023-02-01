# Kotal Operator

Kotal operator is a **cloud agnostic blockchain deployer** that makes it super easy to deploy highly-available, self-managing, self-healing blockchain infrastructure (networks, nodes, storage clusters ...) on any cloud.

## What can I do with Kotal Operator ?

- Deploy Bitcoin rpc nodes
- Deploy ipfs peers and cluster peers
- Deploy ipfs swarms
- Deploy Ethereum transaction and mining nodes
- Deploy Ethereum 2 beacon and validation nodes
- Deploy private Ethereum networks
- Deploy NEAR rpc, archive, and validator nodes
- Deploy Polkadot rpc and validator nodes
- Deploy Chainlink nodes
- Deploy Filecoin nodes
- Deploy Filecoin backed pinning services (FPS)
- Deploy Stacks rpc and api nodes
- Deploy Aptos full and validator nodes


## Kubernetes Custom Resources

Kotal extended kubernetes with custom resources in different API groups.

| Protocol         | Description                                      | API Group                   | Status |
| ---------------- | ------------------------------------------------ | --------------------------- | ------ |
| **Aptos**        | Deploy Aptos full and validator nodes            | aptos.kotal.io/v1alpha1     | alpha  |
| **Bitcoin**      | Deploy Bitcoin nodes                             | bitcoin.kotal.io/v1alpha1   | alpha  |
| **Chainlink**    | Deploy Chainlink nodes                           | chainlink.kotal.io/v1alpha1 | alpha  |
| **Ethereum**     | Deploy private and public network Ethereum nodes | ethereum.kotal.io/v1alpha1  | alpha  |
| **Ethereum 2.0** | Deploy validator and beacon chain nodes          | ethereum2.kotal.io/v1alpha1 | alpha  |
| **Filecoin**     | Deploy Filecoin nodes                            | filecoin.kotal.io/v1alpha1  | alpha  |
| **Graph**        | Deploy graph nodes                               | graph.kotal.io/v1alpha1     | alpha  |
| **IPFS**         | Deploy IPFS peers, cluster peers, and swarms     | ipfs.kotal.io/v1alpha1      | alpha  |
| **NEAR**         | Deploy NEAR rpc, archive and validator nodes     | near.kotal.io/v1alpha1      | alpha  |
| **Polkadot**     | Deploy Polkadot nodes and validator nodes        | polkadot.kotal.io/v1alpha1  | alpha  |
| **Stacks**       | Deploy Stacks rpc and api nodes                  | stacks.kotal.io/v1alpha1    | alpha  |

## Client support

For each protocol, kotal supports at least 1 client (reference client):

| Protocol         | Client(s)                                                                                                                                                                                        |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Aptos**        | [Aptos Core](https://github.com/aptos-labs/aptos-core)                                                                                                                                           |
| **Bitcoin**      | [Bitcoin Core](https://github.com/bitcoin/bitcoin)                                                                                                                                               |
| **Chainlink**    | [Chainlink](https://github.com/smartcontractkit/chainlink)                                                                                                                                       |
| **Ethereum**     | [Hyperledger Besu](https://github.com/hyperledger/besu), [Go-Ethereum](https://github.com/ethereum/go-ethereum), [Nethermind](https://github.com/NethermindEth/nethermind)                       |
| **Ethereum 2.0** | [Teku](https://github.com/ConsenSys/teku), [Prysm](https://github.com/prysmaticlabs/prysm), [Lighthouse](https://github.com/sigp/lighthouse), [Nimbus](https://github.com/status-im/nimbus-eth2) |
| **Filecoin**     | [Lotus](https://github.com/filecoin-project/lotus)                                                                                                                                               |
| **Graph**        | [graph-node](https://github.com/graphprotocol/graph-node)                                                                                                                                        |
| **IPFS**         | [kubo](https://github.com/ipfs/kubo), [ipfs-cluster-service](https://github.com/ipfs/ipfs-cluster)                                                                                         |
| **NEAR**         | [nearcore](https://github.com/near/nearcore)                                                                                                                                                     |
| **Polkadot**     | [Parity Polkadot](https://github.com/paritytech/polkadot)                                                                                                                                        |
| **Stacks**       | [Stacks Node](https://github.com/stacks-network/stacks-blockchain)                                                                                                                               |

## Install Kotal

Kotal requires access to Kubernetes cluster with cert-manager installed.

For development purposes, we recommend [KinD](https://kind.sigs.k8s.io/) (Kubernetes in Docker) to create kubernetes clusters and tear down kubernetes clusters in seconds:

```bash
kind create cluster
```

After the cluster is up and running, [install](https://cert-manager.io/docs/installation/kubernetes/) cert-manager:

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.5.3/cert-manager.yaml
```

Install kotal custom resources and controllers:

```bash
kubectl apply -f https://github.com/kotalco/kotal/releases/download/v0.1.0/kotal.yaml
```

## Example

Ethereum node using Hyperleger Besu client, joining goerli network, and enabling RPC HTTP server:

```yaml
# ethereum-node.yaml
apiVersion: ethereum.kotal.io/v1alpha1
kind: Node
metadata:
  name: ethereum-node
spec:
  client: besu
  network: goerli
  rpc: true
```

```bash
kubectl apply -f ethereum-node.yaml
```

## Documentation

Kotal documentation is available [here](https://docs.kotal.co)

## Get in touch

- [Discord](https://discord.com/invite/kTxy4SA)
- [website](https://kotal.co)
- [@kotalco](https://twitter.com/kotalco)
- [mostafa@kotal.co](mailto:mostafa@kotal.co)

## Contriubuting

TODO

## Licensing

Kotal Blockchain Kubernetes operator is free and open-source software licensed under the [Apache 2.0](LICENSE) License
