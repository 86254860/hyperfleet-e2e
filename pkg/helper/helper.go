package helper

import (
    "context"
    "fmt"
    "time"

    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/config"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/wait"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

// Helper provides utility functions for e2e tests
type Helper struct {
    Cfg       *config.Config
    Client    *client.HyperFleetClient
    K8sClient kubernetes.Interface
}

// initK8sClient initializes a Kubernetes clientset from kubeconfig
func initK8sClient() (kubernetes.Interface, error) {
    // Build config from KUBECONFIG env var or default ~/.kube/config
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
    }

    // Set user agent for observability in API server logs
    config.UserAgent = "hyperfleet-e2e-tests"

    // Create clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
    }

    return clientset, nil
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
//
//    return h.Client.DeleteCluster(ctx, clusterID)
//
// Temporary workaround: delete the Kubernetes namespace using client-go (may temporarily hardcode a timeout duration).
func (h *Helper) CleanupTestCluster(ctx context.Context, clusterID string) error {
    logger.Info("deleting cluster namespace (workaround)", "cluster_id", clusterID, "namespace", clusterID)

    // Delete namespace using client-go
    err := h.K8sClient.CoreV1().Namespaces().Delete(ctx, clusterID, metav1.DeleteOptions{})
    if err != nil && !apierrors.IsNotFound(err) {
        logger.Error("failed to delete cluster namespace", "cluster_id", clusterID, "error", err)
        return fmt.Errorf("failed to delete namespace %s: %w", clusterID, err)
    }

    // Wait for namespace to be fully deleted (garbage collection finalization)
    logger.Info("waiting for namespace deletion to complete", "cluster_id", clusterID)
    backoff := wait.Backoff{
        Duration: 500 * time.Millisecond,
        Factor:   1.5,
        Jitter:   0.1,
        Steps:    20, // ~2 minutes with exponential backoff
    }
    err = wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
        _, err := h.K8sClient.CoreV1().Namespaces().Get(ctx, clusterID, metav1.GetOptions{})
        if apierrors.IsNotFound(err) {
            return true, nil // Namespace fully deleted
        }
        if err != nil {
            return false, err // Unexpected error
        }
        return false, nil // Still exists, keep polling
    })
    if err != nil {
        logger.Error("timeout waiting for namespace deletion", "cluster_id", clusterID, "error", err)
        return fmt.Errorf("timeout waiting for namespace %s deletion: %w", clusterID, err)
    }

    logger.Info("successfully deleted cluster namespace", "cluster_id", clusterID)
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
