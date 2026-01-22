package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
)

// CreateNodePool creates a new nodepool for the specified cluster.
func (c *HyperFleetClient) CreateNodePool(ctx context.Context, clusterID string, req openapi.NodePoolCreateRequest) (*openapi.NodePool, error) {
	logger.Info("creating nodepool", "cluster_id", clusterID, "name", req.Name)

	resp, err := c.Client.CreateNodePool(ctx, clusterID, req)
	if err != nil {
		logger.Error("failed to create nodepool", "cluster_id", clusterID, "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create nodepool: %w", err)
	}

	nodepool, err := handleHTTPResponse[openapi.NodePool](resp, http.StatusCreated, "create nodepool")
	if err != nil {
		return nil, err
	}

	logger.Info("nodepool created", "cluster_id", clusterID, "nodepool_id", *nodepool.Id, "name", req.Name)
	return nodepool, nil
}

// GetNodePool retrieves a nodepool by ID.
func (c *HyperFleetClient) GetNodePool(ctx context.Context, clusterID, nodepoolID string) (*openapi.NodePool, error) {
	resp, err := c.GetNodePoolById(ctx, clusterID, nodepoolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodepool: %w", err)
	}
	return handleHTTPResponse[openapi.NodePool](resp, http.StatusOK, "get nodepool")
}

// ListNodePools retrieves all nodepools for a cluster.
func (c *HyperFleetClient) ListNodePools(ctx context.Context, clusterID string) (*openapi.NodePoolList, error) {
	resp, err := c.GetNodePoolsByClusterId(ctx, clusterID, &openapi.GetNodePoolsByClusterIdParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodepools: %w", err)
	}
	return handleHTTPResponse[openapi.NodePoolList](resp, http.StatusOK, "list nodepools")
}

// GetNodePoolStatuses retrieves all adapter statuses for a nodepool.
func (c *HyperFleetClient) GetNodePoolStatuses(ctx context.Context, clusterID, nodepoolID string) (*openapi.AdapterStatusList, error) {
	resp, err := c.GetNodePoolsStatuses(ctx, clusterID, nodepoolID, &openapi.GetNodePoolsStatusesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodepool statuses: %w", err)
	}
	return handleHTTPResponse[openapi.AdapterStatusList](resp, http.StatusOK, "get nodepool statuses")
}

// CreateNodePoolFromPayload creates a nodepool from a JSON payload file.
// The payload file should contain a NodePoolCreateRequest in JSON format.
func (c *HyperFleetClient) CreateNodePoolFromPayload(ctx context.Context, clusterID, payloadPath string) (*openapi.NodePool, error) {
	logger.Debug("loading nodepool payload", "cluster_id", clusterID, "payload_path", payloadPath)

	req, err := loadPayloadFromFile[openapi.NodePoolCreateRequest](payloadPath)
	if err != nil {
		logger.Error("failed to load payload", "cluster_id", clusterID, "payload_path", payloadPath, "error", err)
		return nil, err
	}

	return c.CreateNodePool(ctx, clusterID, *req)
}

// DeleteNodePool deletes a nodepool by ID.
// TODO(API): Implement nodepool deletion once HyperFleet API supports DELETE operations.
// Currently this is a no-op as the API does not support nodepool deletion yet.
// Resources will remain in the system until manually cleaned up.
func (c *HyperFleetClient) DeleteNodePool(ctx context.Context, clusterID, nodepoolID string) error {
	// HyperFleet API does not yet support nodepool deletion
	// Log this as info (not error) since it's expected behavior
	logger.Debug("nodepool deletion not supported by API - skipping", "cluster_id", clusterID, "nodepool_id", nodepoolID)
	return nil
}
