# Kotal Operator

Kotal operator is a **cloud agnostic blockchain deployer** that makes it super easy to deploy highly-available, self-managing, self-healing blockchain infrastructure (networks, nodes, storage clusters ...) on any cloud.

## What can I do with Kotal Operator ?

- Deploy ipfs peers and cluster peers
- Deploy ipfs swarms
- Deploy Ethereum transaction and mining nodes
- Deploy Ethereum 2 beacon and validation nodes
- Deploy private Ethereum networks
- Deploy Filecoin nodes
- Deploy Filecoin backed pinning services (FPS)

## Kubernetes Custom Resources

Kotal extended kubernetes with custom resources in different API groups.

| Group            | Description                                  | API Group                   | Status               |
| ---------------- | -------------------------------------------- | --------------------------- | -------------------- |
| **Ethereum**     | Deploy ethereum nodes                        | ethereum.kotal.io/v1alpha1  | alpha                |
| **IPFS**         | Deploy IPFS peers, cluster peers, and swarms | ipfs.kotal.io/v1alpha1      | alpha                |
| **Filecoin**     | Deploy Filecoin nodes                        | filecoin.kotal.io/v1alpha1  | alpha                |
| **Ethereum 2.0** | Deploy validator and beacon chain nodes      | ethereum2.kotal.io/v1alpha1 | alpha                |
| **Algorand**     | Deploy relay and participation nodes         | algorand.kotal.io/v1alpha1  | coming soon :rocket: |

## Client support

For each protocol, kotal supports at least 1 client (reference client), client can be changed by updating `client: ...` specification parameter.

| Protocol         | Client(s)                                    |
| ---------------- | -------------------------------------------- |
| **Ethereum**     | Hyperledger Besu, Go-Ethereum, Open Ethereum |
| **Ethereum 2.0** | Teku, Prysm, Lighthouse, Nimbus              |
| **Filecoin**     | Lotus                                        |
| **IFPS**         | go-ipfs                                      |

## Install Kotal

Kotal requires access to Kubernetes cluster with cert-manager installed.

For development purposes, we recommend [KinD](https://kind.sigs.k8s.io/) (Kubernetes in Docker) to create kubernetes clusters and tear down kubernetes clusters in seconds:

```bash
kind create cluster
```

After the cluster is up and running, [install](https://cert-manager.io/docs/installation/kubernetes/) cert-manager:

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.2.0/cert-manager.yaml
```

Install kotal custom resources and controllers:

```bash
kubectl apply -f https://github.com/kotalco/kotal/releases/download/v0.1-alpha.4/kotal.yaml
```

## Example

Ethereum node using Hyperleger Besu client, joining rinkeby network, and enabling RPC HTTP server:

```yaml
# ethereum-node.yaml
apiVersion: ethereum.kotal.io/v1alpha1
kind: Node
metadata:
  name: ethereum-node
spec:
  client: besu
  network: rinkeby
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

TODO
