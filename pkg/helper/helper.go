package helper

import (
    "context"
    "fmt"
    "os/exec"
    "time"

    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/config"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
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
// TODO: Replace this workaround with API DELETE once HyperFleet API supports
// DELETE operations for clusters resource type:
//   return h.Client.DeleteCluster(ctx, clusterID)
// Current workaround: Delete the Kubernetes namespace using kubectl
func (h *Helper) CleanupTestCluster(ctx context.Context, clusterID string) error {
    logger.Info("deleting cluster namespace (workaround)", "cluster_id", clusterID, "namespace", clusterID)

    // Create context with timeout for kubectl command
    cmdCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
    defer cancel()

    // Execute kubectl delete namespace command
    cmd := exec.CommandContext(cmdCtx, "kubectl", "delete", "namespace", clusterID)
    output, err := cmd.CombinedOutput()
    if err != nil {
        logger.Error("failed to delete cluster namespace", "cluster_id", clusterID, "error", err, "output", string(output))
        return fmt.Errorf("failed to delete namespace %s: %w (output: %s)", clusterID, err, string(output))
    }

    logger.Info("successfully deleted cluster namespace", "cluster_id", clusterID, "output", string(output))
    return nil
}

// GetTestNodePool creates a nodepool on the specified cluster from a payload file
func (h *Helper) GetTestNodePool(ctx context.Context, clusterID, payloadPath string) (*openapi.NodePool, error) {
    return h.Client.CreateNodePoolFromPayload(ctx, clusterID, payloadPath)
}

// CleanupTestNodePool cleans up test nodepool
func (h *Helper) CleanupTestNodePool(ctx context.Context, clusterID, nodepoolID string) error {
    return h.Client.DeleteNodePool(ctx, clusterID, nodepoolID)
}
