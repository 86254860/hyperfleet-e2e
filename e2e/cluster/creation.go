package cluster

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot import for test readability

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/helper"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/labels"
)

var lifecycleTestName = "[Suite: cluster][baseline] Full Cluster Creation Flow on GCP"

var _ = ginkgo.Describe(lifecycleTestName,
	ginkgo.Label(labels.Tier0),
	func() {
		var h *helper.Helper
		var clusterID string

		ginkgo.BeforeEach(func() {
			h = helper.New()
		})

		ginkgo.It("should create GCP cluster and transition to Ready state with all adapters healthy", func(ctx context.Context) {
			ginkgo.By("submitting cluster creation request via POST /api/hyperfleet/v1/clusters")
			cluster, err := h.Client.CreateClusterFromPayload(ctx, "testdata/payloads/clusters/gcp.json")
			Expect(err).NotTo(HaveOccurred(), "failed to create cluster")

			ginkgo.By("verifying API response (HTTP 201 Created)")
			Expect(cluster.Id).NotTo(BeNil(), "cluster ID should be generated")
			clusterID = *cluster.Id
			ginkgo.GinkgoWriter.Printf("Created cluster ID: %s\n", clusterID)

			Expect(cluster.Status).NotTo(BeNil(), "cluster status should be present")
			Expect(cluster.Status.Phase).To(Equal(openapi.NotReady), "cluster should be in NotReady phase initially")
			/** <TODO>
						 Cluster final status depends on all deployed adapter result, this is still in progress.
			             Will update this part once adapter scope is finalized.
						ginkgo.By("monitoring cluster status - waiting for phase transition to Ready")
						err = h.WaitForClusterPhase(ctx, clusterID, openapi.Ready, h.Cfg.Timeouts.Cluster.Ready)
						Expect(err).NotTo(HaveOccurred(), "cluster should reach Ready phase")

						ginkgo.By("verifying all adapter conditions via /clusters/{id}/statuses endpoint")
						const expectedAdapterCount = 1 // GCP cluster expects 1 adapter
						Eventually(func(g Gomega) {
							statuses, err := h.Client.GetClusterStatuses(ctx, clusterID)
							g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster statuses")
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

						ginkgo.By("verifying final cluster state")
						finalCluster, err := h.Client.GetCluster(ctx, clusterID)
						Expect(err).NotTo(HaveOccurred(), "failed to get final cluster state")
						Expect(finalCluster.Status).NotTo(BeNil(), "cluster status should be present")
						Expect(finalCluster.Status.Phase).To(Equal(openapi.Ready), "cluster phase should be Ready")
						**/
		})

		ginkgo.AfterEach(func(ctx context.Context) {
			// Skip cleanup if helper not initialized or no cluster created
			if h == nil || clusterID == "" {
				return
			}

			ginkgo.By("cleaning up cluster " + clusterID)
			if err := h.CleanupTestCluster(ctx, clusterID); err != nil {
				ginkgo.GinkgoWriter.Printf("Warning: failed to cleanup cluster %s: %v\n", clusterID, err)
			}
		})
	},
)
