package kubernetes

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps kubernetes.Interface and provides business logic methods
type Client struct {
	kubernetes.Interface
}

// NewClient initializes a Kubernetes clientset from kubeconfig
func NewClient() (*Client, error) {
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

	return &Client{Interface: clientset}, nil
}

// DeleteNamespaceAndWait deletes a namespace and waits for it to be fully removed
func (c *Client) DeleteNamespaceAndWait(ctx context.Context, namespace string) error {
	// Delete namespace
	err := c.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete namespace %s: %w", namespace, err)
	}

	// Wait for namespace to be fully deleted (garbage collection finalization)
	backoff := wait.Backoff{
		Duration: 500 * time.Millisecond,
		Factor:   1.5,
		Jitter:   0.1,
		Steps:    20,
		Cap:      10 * time.Second, // Cap individual retries at 10s for ~2.4 min total timeout
	}
	err = wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		_, err := c.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil // Namespace fully deleted
		}
		if err != nil {
			return false, err // Unexpected error
		}
		return false, nil // Still exists, keep polling
	})
	if err != nil {
		return fmt.Errorf("timeout waiting for namespace %s deletion: %w", namespace, err)
	}

	return nil
}
