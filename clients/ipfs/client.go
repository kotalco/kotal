package ipfs

import (
	"fmt"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/clients"
	"k8s.io/apimachinery/pkg/runtime"
)

// IPFSClient is IPFS peer client
type IPFSClient interface {
	clients.Interface
}

// NewClient creates a new client for ipfs peer or cluster peer
func NewClient(obj runtime.Object) (IPFSClient, error) {
	switch peer := obj.(type) {
	case *ipfsv1alpha1.Peer:
		return &KuboClient{peer}, nil
	case *ipfsv1alpha1.ClusterPeer:
		return &GoIPFSClusterClient{peer}, nil
	}
	return nil, fmt.Errorf("no client support for %s", obj)
}
