package nodepool

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot import for test readability

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/helper"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/labels"
)

var lifecycleTestName = "[Suite: nodepool] Full NodePool Creation Flow"

var _ = ginkgo.Describe(lifecycleTestName,
	ginkgo.Label(labels.Tier0, labels.Stable, labels.HappyPath, labels.Lifecycle),
	func() {
		var h *helper.Helper
		var clusterID string
		var nodepoolID string

		ginkgo.BeforeEach(func() {
			h = helper.New()
		})

		ginkgo.It("should create nodepool on existing cluster and transition to Ready state", func(ctx context.Context) {
			ginkgo.By("getting test cluster for nodepool creation")
			var err error
			clusterID, err = h.GetTestCluster(ctx, "testdata/payloads/clusters/gcp.json")
			Expect(err).NotTo(HaveOccurred(), "failed to get test cluster")
			ginkgo.GinkgoWriter.Printf("Using cluster ID: %s\n", clusterID)

			ginkgo.By("waiting for cluster to become Ready")
			err = h.WaitForClusterPhase(ctx, clusterID, openapi.Ready, h.Cfg.Timeouts.Cluster.Ready)
			Expect(err).NotTo(HaveOccurred(), "cluster should be in Ready phase")

			ginkgo.By("submitting nodepool creation request via POST /api/hyperfleet/v1/clusters/{id}/nodepools")
			nodepool, err := h.Client.CreateNodePoolFromPayload(ctx, clusterID, "testdata/payloads/nodepools/gcp.json")
			Expect(err).NotTo(HaveOccurred(), "failed to create nodepool")

			ginkgo.By("verifying API response (HTTP 201 Created)")
			Expect(nodepool.Id).NotTo(BeNil(), "nodepool ID should be generated")
			nodepoolID = *nodepool.Id
			ginkgo.GinkgoWriter.Printf("Created nodepool ID: %s\n", nodepoolID)

			Expect(nodepool.Status).NotTo(BeNil(), "nodepool status should be present")
			Expect(nodepool.Status.Phase).To(Equal(openapi.NotReady), "nodepool should be in NotReady phase initially")

			ginkgo.By("monitoring nodepool status - waiting for phase transition to Ready")
			err = h.WaitForNodePoolPhase(ctx, clusterID, nodepoolID, openapi.Ready, h.Cfg.Timeouts.NodePool.Ready)
			Expect(err).NotTo(HaveOccurred(), "nodepool should reach Ready phase")

			ginkgo.By("verifying all nodepool adapter conditions")
			const expectedAdapterCount = 1 // GCP nodepool expects 1 adapter
			Eventually(func(g Gomega) {
				statuses, err := h.Client.GetNodePoolStatuses(ctx, clusterID, nodepoolID)
				g.Expect(err).NotTo(HaveOccurred(), "failed to get nodepool statuses")
				g.Expect(statuses.Items).To(HaveLen(expectedAdapterCount),
					"expected %d adapter(s), got %d", expectedAdapterCount, len(statuses.Items))

				for _, adapter := range statuses.Items {
					hasApplied := h.HasCondition(adapter.Conditions, client.ConditionTypeApplied, openapi.True)
					g.Expect(hasApplied).To(BeTrue(),
						"adapter %s should have Applied=True", adapter.Adapter)

					hasAvailable := h.HasCondition(adapter.Conditions, client.ConditionTypeAvailable, openapi.True)
					g.Expect(hasAvailable).To(BeTrue(),
						"adapter %s should have Available=True", adapter.Adapter)

					hasHealth := h.HasCondition(adapter.Conditions, client.ConditionTypeHealth, openapi.True)
					g.Expect(hasHealth).To(BeTrue(),
						"adapter %s should have Health=True", adapter.Adapter)
				}
			}, h.Cfg.Timeouts.Adapter.Processing, h.Cfg.Polling.Interval).Should(Succeed())

			ginkgo.By("verifying final nodepool state")
			finalNodePool, err := h.Client.GetNodePool(ctx, clusterID, nodepoolID)
			Expect(err).NotTo(HaveOccurred(), "failed to get final nodepool state")
			Expect(finalNodePool.Status).NotTo(BeNil(), "nodepool status should be present")
			Expect(finalNodePool.Status.Phase).To(Equal(openapi.Ready), "nodepool phase should be Ready")
		})

		ginkgo.AfterEach(func(ctx context.Context) {
			// Skip cleanup if helper not initialized or no cluster created
			// Note: Deleting cluster will cascade delete nodepool automatically
			if h == nil || clusterID == "" {
				return
			}

			ginkgo.By("cleaning up test cluster " + clusterID)
			if err := h.CleanupTestCluster(ctx, clusterID); err != nil {
				ginkgo.GinkgoWriter.Printf("Warning: failed to cleanup cluster %s: %v\n", clusterID, err)
			}
		})
	},
)
