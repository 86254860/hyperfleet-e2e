package helper

import (
	"context"
	"fmt"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/config"
)

// Helper provides utility functions for e2e tests
type Helper struct {
	Cfg    *config.Config
	Client *client.HyperFleetClient
}

// GetTestCluster creates a new temporary test cluster
func (h *Helper) GetTestCluster(ctx context.Context, payloadPath string) (string, error) {
	cluster, err := h.Client.CreateClusterFromPayload(ctx, payloadPath)
	if err != nil {
		return "", err
	}
	if cluster == nil {
		return "", fmt.Errorf("CreateClusterFromPayload returned nil")
	}
	if cluster.Id == nil {
		return "", fmt.Errorf("created cluster has no ID")
	}
	return *cluster.Id, nil
}

// CleanupTestCluster deletes the temporary test cluster
func (h *Helper) CleanupTestCluster(ctx context.Context, clusterID string) error {
	return h.Client.DeleteCluster(ctx, clusterID)
}

// GetTestNodePool creates a nodepool on the specified cluster from a payload file
func (h *Helper) GetTestNodePool(ctx context.Context, clusterID, payloadPath string) (*openapi.NodePool, error) {
	return h.Client.CreateNodePoolFromPayload(ctx, clusterID, payloadPath)
}

// CleanupTestNodePool cleans up test nodepool
func (h *Helper) CleanupTestNodePool(ctx context.Context, clusterID, nodepoolID string) error {
	return h.Client.DeleteNodePool(ctx, clusterID, nodepoolID)
}
