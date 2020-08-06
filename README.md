# Kotal Operator

Kotal operator is **cloud agnostic blockchain deployer** that make it easy to deploy highly available, self-managing, self-healing blockchain infrastructure (networks, nodes, storage clusters ...) on any cloud.

## What can I do with Kotal Operator ?
* Create ipfs storage cluster and swarms
* Deploy highly available blokchain networks and nodes
* Join public networks like Ethereum or Bitcoin Mainnet
* Join test networks like Rinkeby, Ropsten or Goerli
* Launch your own validator node
* Launch your own baking node
* Create and join private consortium networks

## Kubernetes Custom Resources

Kotal extended kubernetes with custom resources in different API groups.

| Resource | Description | API Group | Status |
| -------- | ------ | ----------- | --- |
| **Ethereum**| Create and join private and public ethereum networks | ethereum.kotal.io/v1alpha1 | alpha |
| **Ethereum 2.0**  | Create validator and beacon chain nodes | ethereum2.kotal.io/v1alpha1 | coming soon :rocket:  |
| **IPFS**  | Create and join ipfs storage cluster and swarms | ipfs.kotal.io/v1alpha1 | alpha  |

## Example

Check `config/samples` directory for more examples

Network of a single node that will join and sync rinkeby blockchain:

```yaml
apiVersion: ethereum.kotal.io/v1alpha1
kind: Network
metadata:
  name: my-network
spec:
  join: rinkeby
  nodes:
    - name: node-1
      rpc: true
      rpcAPI:
        - web3
        - net
        - eth
```

## Requiremenets

* Access to kubernetes cluster v1.11+
* cert-manager v0.15+

## Quick Start

For development purposes, we recommend [KinD](https://kind.sigs.k8s.io/) (Kubernetes in Docker) to create kubernetes clusters and tear down kubernetes clusters in seconds

```
kind create cluster
```

after the cluster is up and running, [install](https://cert-manager.io/docs/installation/kubernetes/) cert-manager

```
kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.16.0/cert-manager.yaml
```

Install kotal to the kind cluster.

```
make kind IMG=kotal/operator:0.1
```

This command will run the tests, build the docker image, load it into kind and deploy the operator.

Apply any of the examples in `config/samples` directory and watch magic happens :tophat:

```
kubectl apply -f config/samples/ipfs/ipfs_v1alpha1_swarm.yaml
```

this sample will create an ipfs storage swarm of 3 nodes.

Finally, tear down the cluster

```
kind delete cluster
```

## Documentation

TODO

## Contact

* Slack kotal.slack.com
* website kotal.co
* twitter [@kotalco](https://twitter.com/kotalco)
* github github.com/kotalco
* email mostafa@kotal.co

## Contriubuting

TODO

## Licensing

TODO
