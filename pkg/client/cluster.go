package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
)

// CreateCluster creates a new cluster and returns the created cluster object.
func (c *HyperFleetClient) CreateCluster(ctx context.Context, req openapi.ClusterCreateRequest) (*openapi.Cluster, error) {
	logger.Info("creating cluster", "name", req.Name)

	resp, err := c.PostCluster(ctx, req)
	if err != nil {
		logger.Error("failed to create cluster", "name", req.Name, "error", err)
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	cluster, err := handleHTTPResponse[openapi.Cluster](resp, http.StatusCreated, "create cluster")
	if err != nil {
		return nil, err
	}

	logger.Info("cluster created", "cluster_id", *cluster.Id, "name", req.Name)
	return cluster, nil
}

// GetCluster retrieves a cluster by ID.
func (c *HyperFleetClient) GetCluster(ctx context.Context, clusterID string) (*openapi.Cluster, error) {
	resp, err := c.GetClusterById(ctx, clusterID, &openapi.GetClusterByIdParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}
	return handleHTTPResponse[openapi.Cluster](resp, http.StatusOK, "get cluster")
}

// ListClusters retrieves all clusters.
func (c *HyperFleetClient) ListClusters(ctx context.Context) (*openapi.ClusterList, error) {
	resp, err := c.GetClusters(ctx, &openapi.GetClustersParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	return handleHTTPResponse[openapi.ClusterList](resp, http.StatusOK, "list clusters")
}

// GetClusterStatuses retrieves all adapter statuses for a cluster.
func (c *HyperFleetClient) GetClusterStatuses(ctx context.Context, clusterID string) (*openapi.AdapterStatusList, error) {
	resp, err := c.Client.GetClusterStatuses(ctx, clusterID, &openapi.GetClusterStatusesParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster statuses: %w", err)
	}
	return handleHTTPResponse[openapi.AdapterStatusList](resp, http.StatusOK, "get cluster statuses")
}

// CreateClusterFromPayload creates a cluster from a JSON payload file.
// The payload file should contain a ClusterCreateRequest in JSON format.
func (c *HyperFleetClient) CreateClusterFromPayload(ctx context.Context, payloadPath string) (*openapi.Cluster, error) {
	logger.Debug("loading cluster payload", "payload_path", payloadPath)

	req, err := loadPayloadFromFile[openapi.ClusterCreateRequest](payloadPath)
	if err != nil {
		logger.Error("failed to load payload", "payload_path", payloadPath, "error", err)
		return nil, err
	}

	return c.CreateCluster(ctx, *req)
}

// DeleteCluster deletes a cluster by ID.
// TODO(API): Implement cluster deletion once HyperFleet API supports DELETE operations.
// Currently this is a no-op as the API does not support cluster deletion yet.
// Resources will remain in the system until manually cleaned up.
func (c *HyperFleetClient) DeleteCluster(ctx context.Context, clusterID string) error {
	// HyperFleet API does not yet support cluster deletion
	// Log this as info (not error) since it's expected behavior
	logger.Debug("cluster deletion not supported by API - skipping", "cluster_id", clusterID)
	return nil
}
