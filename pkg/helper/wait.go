package helper

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/gomega" //nolint:staticcheck // dot import for test readability

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
)

// WaitForClusterCondition waits for a cluster to have a specific condition with the expected status
func (h *Helper) WaitForClusterCondition(ctx context.Context, clusterID string, conditionType string, expectedStatus openapi.ResourceConditionStatus, timeout time.Duration) error {
	logger.Debug("waiting for cluster condition", "cluster_id", clusterID, "condition_type", conditionType, "expected_status", expectedStatus, "timeout", timeout)

	Eventually(func(g Gomega) {
		cluster, err := h.Client.GetCluster(ctx, clusterID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster")
		g.Expect(cluster).NotTo(BeNil(), "cluster is nil")
		g.Expect(cluster.Status).NotTo(BeNil(), "cluster.Status is nil")

		// Check if the condition exists with the expected status
		found := false
		for _, cond := range cluster.Status.Conditions {
			if cond.Type == conditionType && cond.Status == expectedStatus {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(),
			fmt.Sprintf("cluster does not have condition %s=%s", conditionType, expectedStatus))
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	logger.Info("cluster reached target condition", "cluster_id", clusterID, "condition_type", conditionType, "status", expectedStatus)
	return nil
}

// WaitForAdapterCondition waits for a specific adapter condition to be in the expected status
func (h *Helper) WaitForAdapterCondition(ctx context.Context, clusterID, adapterName, condType string, expectedStatus openapi.AdapterConditionStatus, timeout time.Duration) error {
	Eventually(func(g Gomega) {
		statuses, err := h.Client.GetClusterStatuses(ctx, clusterID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster statuses")

		// Find the specific adapter
		var found bool
		for _, status := range statuses.Items {
			if status.Adapter == adapterName {
				found = true
				hasCondition := h.HasAdapterCondition(status.Conditions, condType, expectedStatus)
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
func (h *Helper) WaitForAllAdapterConditions(ctx context.Context, clusterID, condType string, expectedStatus openapi.AdapterConditionStatus, timeout time.Duration) error {
	Eventually(func(g Gomega) {
		statuses, err := h.Client.GetClusterStatuses(ctx, clusterID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster statuses")

		for _, adapterStatus := range statuses.Items {
			hasCondition := h.HasAdapterCondition(adapterStatus.Conditions, condType, expectedStatus)
			g.Expect(hasCondition).To(BeTrue(),
				fmt.Sprintf("adapter %s does not have condition %s=%s",
					adapterStatus.Adapter, condType, expectedStatus))
		}
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	return nil
}

// WaitForNodePoolCondition waits for a nodepool to have a specific condition with the expected status
func (h *Helper) WaitForNodePoolCondition(ctx context.Context, clusterID, nodepoolID string, conditionType string, expectedStatus openapi.ResourceConditionStatus, timeout time.Duration) error {
	logger.Debug("waiting for nodepool condition", "cluster_id", clusterID, "nodepool_id", nodepoolID, "condition_type", conditionType, "expected_status", expectedStatus, "timeout", timeout)

	Eventually(func(g Gomega) {
		nodepool, err := h.Client.GetNodePool(ctx, clusterID, nodepoolID)
		g.Expect(err).NotTo(HaveOccurred(), "failed to get nodepool")
		g.Expect(nodepool).NotTo(BeNil(), "nodepool is nil")
		g.Expect(nodepool.Status).NotTo(BeNil(), "nodepool.Status is nil")

		// Check if the condition exists with the expected status
		found := false
		for _, cond := range nodepool.Status.Conditions {
			if cond.Type == conditionType && cond.Status == expectedStatus {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(),
			fmt.Sprintf("nodepool does not have condition %s=%s", conditionType, expectedStatus))
	}, timeout, h.Cfg.Polling.Interval).Should(Succeed())

	logger.Info("nodepool reached target condition", "cluster_id", clusterID, "nodepool_id", nodepoolID, "condition_type", conditionType, "status", expectedStatus)
	return nil
}
