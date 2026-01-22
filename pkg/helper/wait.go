package helper

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega" //nolint:staticcheck // dot import for test readability

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
)

// WaitForClusterPhase waits for a cluster to reach the expected phase
func (h *Helper) WaitForClusterPhase(ctx context.Context, clusterID string, expectedPhase openapi.ResourcePhase, timeout time.Duration) error {
	logger.Debug("waiting for cluster phase transition", "cluster_id", clusterID, "target_phase", expectedPhase, "timeout", timeout)

	Eventually(func(g Gomega) {
		cluster, err := h.Client.GetCluster(ctx, clusterID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
		g.Expect(cluster).NotTo(BeNil(), "cluster is nil")
		g.Expect(cluster.Status).NotTo(BeNil(), "cluster.Status is nil")
		g.Expect(cluster.Status.Phase).To(Equal(expectedPhase),
			fmt.Sprintf("cluster phase: got %s, want %s",
				cluster.Status.Phase, expectedPhase))
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	logger.Info("cluster reached target phase", "cluster_id", clusterID, "phase", expectedPhase)
	return nil
}

// WaitForAdapterCondition waits for a specific adapter condition to be in the expected status
func (h *Helper) WaitForAdapterCondition(ctx context.Context, clusterID, adapterName, condType string, expectedStatus openapi.ConditionStatus, timeout time.Duration) error {
	Eventually(func(g Gomega) {
		statuses, err := h.Client.GetClusterStatuses(ctx, clusterID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster statuses")

		// Find the specific adapter
		var found bool
		for _, status := range statuses.Items {
			if status.Adapter == adapterName {
				found = true
				hasCondition := h.HasCondition(status.Conditions, condType, expectedStatus)
				g.Expect(hasCondition).To(BeTrue(),
					fmt.Sprintf("adapter %s does not have condition %s=%s", adapterName, condType, expectedStatus))
				break
			}
		}
		g.Expect(found).To(BeTrue(), fmt.Sprintf("adapter %s not found", adapterName))
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	return nil
}

// WaitForAllAdapterConditions waits for all adapters to have the specified condition
func (h *Helper) WaitForAllAdapterConditions(ctx context.Context, clusterID, condType string, expectedStatus openapi.ConditionStatus, timeout time.Duration) error {
	Eventually(func(g Gomega) {
		statuses, err := h.Client.GetClusterStatuses(ctx, clusterID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster statuses")

		for _, adapterStatus := range statuses.Items {
			hasCondition := h.HasCondition(adapterStatus.Conditions, condType, expectedStatus)
			g.Expect(hasCondition).To(BeTrue(),
				fmt.Sprintf("adapter %s does not have condition %s=%s",
					adapterStatus.Adapter, condType, expectedStatus))
		}
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	return nil
}

// WaitForNodePoolPhase waits for a nodepool to reach the expected phase
func (h *Helper) WaitForNodePoolPhase(ctx context.Context, clusterID, nodepoolID string, expectedPhase openapi.ResourcePhase, timeout time.Duration) error {
	logger.Debug("waiting for nodepool phase transition", "cluster_id", clusterID, "nodepool_id", nodepoolID, "target_phase", expectedPhase, "timeout", timeout)

	Eventually(func(g Gomega) {
		nodepool, err := h.Client.GetNodePool(ctx, clusterID, nodepoolID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get nodepool")
		g.Expect(nodepool).NotTo(BeNil(), "nodepool is nil")
		g.Expect(nodepool.Status).NotTo(BeNil(), "nodepool.Status is nil")
		g.Expect(nodepool.Status.Phase).To(Equal(expectedPhase),
			fmt.Sprintf("nodepool phase: got %s, want %s",
				nodepool.Status.Phase, expectedPhase))
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	logger.Info("nodepool reached target phase", "cluster_id", clusterID, "nodepool_id", nodepoolID, "phase", expectedPhase)
	return nil
}
