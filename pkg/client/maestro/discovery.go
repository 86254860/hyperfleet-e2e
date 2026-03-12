package maestro

import (
	"context"
	"fmt"
	"time"

	k8sclient "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MaestroServiceName is the name of the Maestro service in Kubernetes
	MaestroServiceName = "maestro"

	// MaestroNamespace is the namespace where Maestro is deployed
	MaestroNamespace = "maestro"

	// MaestroHTTPPortName is the name of the HTTP port in the Maestro service
	MaestroHTTPPortName = "http"
)

// DiscoverMaestroURL attempts to discover the Maestro service URL from the Kubernetes cluster
// It looks for the maestro service in the maestro namespace and constructs the in-cluster URL
// Returns an error if the Kubernetes client cannot be created or the service is not found
func DiscoverMaestroURL() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	k8sClient, err := k8sclient.NewClient()
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Get the maestro service
	svc, err := k8sClient.CoreV1().Services(MaestroNamespace).Get(ctx, MaestroServiceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get maestro service: %w", err)
	}

	// Find the HTTP port
	var port int32
	for _, p := range svc.Spec.Ports {
		if p.Name == MaestroHTTPPortName {
			port = p.Port
			break
		}
	}

	if port == 0 {
		return "", fmt.Errorf("HTTP port not found in maestro service (looking for port named '%s')", MaestroHTTPPortName)
	}

	// Construct the in-cluster service URL
	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		svc.Name,
		svc.Namespace,
		port,
	)

	return url, nil
}
