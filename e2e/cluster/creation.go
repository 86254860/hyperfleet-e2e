package cluster

import (
    "context"

    "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega" //nolint:staticcheck // dot import for test readability

    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/helper"
    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/labels"
)

var lifecycleTestName = "[Suite: cluster][baseline] Clusters Resource Type - Workflow Validation"

var _ = ginkgo.Describe(lifecycleTestName,
    ginkgo.Label(labels.Tier0),
    func() {
        var h *helper.Helper
        var clusterID string

        ginkgo.BeforeEach(func() {
            h = helper.New()
        })

        // This test validates the end-to-end cluster lifecycle workflow:
        // 1. Cluster creation via API with initial condition validation
        // 2. Required adapter execution with comprehensive metadata validation
        // 3. Final cluster state verification (Ready and Available conditions)
        ginkgo.It("should validate complete workflow for clusters resource type from creation to Ready state",
            func(ctx context.Context) {
                ginkgo.By("Submit an API request to create a Cluster resource")
                cluster, err := h.Client.CreateClusterFromPayload(ctx, "testdata/payloads/clusters/cluster-request.json")
                Expect(err).NotTo(HaveOccurred(), "failed to create cluster")
                Expect(cluster.Id).NotTo(BeNil(), "cluster ID should be generated")
                clusterID = *cluster.Id
                ginkgo.GinkgoWriter.Printf("Created cluster ID: %s\n", clusterID)
                Expect(cluster.Status).NotTo(BeNil(), "cluster status should be present")

                ginkgo.By("Verify initial status of cluster")
                // Verify initial conditions are False, indicating workflow has not completed yet
                // This ensures the cluster starts in the correct initial state
                hasReadyFalse := h.HasResourceCondition(cluster.Status.Conditions,
                    client.ConditionTypeReady, openapi.ResourceConditionStatusFalse)
                Expect(hasReadyFalse).To(BeTrue(),
                    "initial cluster conditions should have Ready=False")

                hasAvailableFalse := h.HasResourceCondition(cluster.Status.Conditions,
                    client.ConditionTypeAvailable, openapi.ResourceConditionStatusFalse)
                Expect(hasAvailableFalse).To(BeTrue(),
                    "initial cluster conditions should have Available=False")

                ginkgo.By("Verify required adapter execution results")
                // Validate required adapters from config have completed successfully
                // If an adapter fails, we can identify which specific adapter failed
                Eventually(func(g Gomega) {
                    statuses, err := h.Client.GetClusterStatuses(ctx, clusterID)
                    g.Expect(err).NotTo(HaveOccurred(), "failed to get cluster statuses")
                    g.Expect(statuses.Items).NotTo(BeEmpty(), "at least one adapter should have executed")

                    // Build a map of adapter statuses for easy lookup
                    adapterMap := make(map[string]openapi.AdapterStatus)
                    for _, adapter := range statuses.Items {
                        adapterMap[adapter.Adapter] = adapter
                    }

                    // Validate each required adapter from config
                    for _, requiredAdapter := range h.Cfg.Adapters.Cluster {
                        adapter, exists := adapterMap[requiredAdapter]
                        g.Expect(exists).To(BeTrue(),
                            "required adapter %s should be present in adapter statuses", requiredAdapter)

                        // Validate adapter-level metadata
                        g.Expect(adapter.CreatedTime).NotTo(BeZero(),
                            "adapter %s should have valid created_time", adapter.Adapter)
                        g.Expect(adapter.LastReportTime).NotTo(BeZero(),
                            "adapter %s should have valid last_report_time", adapter.Adapter)
                        g.Expect(adapter.ObservedGeneration).To(Equal(int32(1)),
                            "adapter %s should have observed_generation=1 for new creation request", adapter.Adapter)

                        hasApplied := h.HasAdapterCondition(
                            adapter.Conditions,
                            client.ConditionTypeApplied,
                            openapi.AdapterConditionStatusTrue,
                        )
                        g.Expect(hasApplied).To(BeTrue(),
                            "adapter %s should have Applied=True", adapter.Adapter)

                        hasAvailable := h.HasAdapterCondition(
                            adapter.Conditions,
                            client.ConditionTypeAvailable,
                            openapi.AdapterConditionStatusTrue,
                        )
                        g.Expect(hasAvailable).To(BeTrue(),
                            "adapter %s should have Available=True", adapter.Adapter)

                        hasHealth := h.HasAdapterCondition(
                            adapter.Conditions,
                            client.ConditionTypeHealth,
                            openapi.AdapterConditionStatusTrue,
                        )
                        g.Expect(hasHealth).To(BeTrue(),
                            "adapter %s should have Health=True", adapter.Adapter)

                        // Validate condition metadata for each condition
                        for _, condition := range adapter.Conditions {
                            g.Expect(condition.Reason).NotTo(BeNil(),
                                "adapter %s condition %s should have non-nil reason", adapter.Adapter, condition.Type)
                            g.Expect(*condition.Reason).NotTo(BeEmpty(),
                                "adapter %s condition %s should have non-empty reason", adapter.Adapter, condition.Type)

                            g.Expect(condition.Message).NotTo(BeNil(),
                                "adapter %s condition %s should have non-nil message", adapter.Adapter, condition.Type)
                            g.Expect(*condition.Message).NotTo(BeEmpty(),
                                "adapter %s condition %s should have non-empty message", adapter.Adapter, condition.Type)

                            g.Expect(condition.LastTransitionTime).NotTo(BeZero(),
                                "adapter %s condition %s should have valid last_transition_time", adapter.Adapter, condition.Type)
                        }
                    }
                }, h.Cfg.Timeouts.Adapter.Processing, h.Cfg.Polling.Interval).Should(Succeed())

                ginkgo.By("Verify final cluster state")
                // Wait for cluster Ready condition and verify both Ready and Available conditions are True
                // This confirms the cluster has reached the desired end state
                err = h.WaitForClusterCondition(
                    ctx,
                    clusterID,
                    client.ConditionTypeReady,
                    openapi.ResourceConditionStatusTrue,
                    h.Cfg.Timeouts.Cluster.Ready,
                )
                Expect(err).NotTo(HaveOccurred(), "cluster Ready condition should transition to True")

                finalCluster, err := h.Client.GetCluster(ctx, clusterID)
                Expect(err).NotTo(HaveOccurred(), "failed to get final cluster state")
                Expect(finalCluster.Status).NotTo(BeNil(), "cluster status should be present")

                hasReady := h.HasResourceCondition(finalCluster.Status.Conditions,
                    client.ConditionTypeReady, openapi.ResourceConditionStatusTrue)
                Expect(hasReady).To(BeTrue(), "cluster should have Ready=True condition")

                hasAvailable := h.HasResourceCondition(finalCluster.Status.Conditions,
                    client.ConditionTypeAvailable, openapi.ResourceConditionStatusTrue)
                Expect(hasAvailable).To(BeTrue(), "cluster should have Available=True condition")

                // Validate observedGeneration for Ready and Available conditions
                for _, condition := range finalCluster.Status.Conditions {
                    if condition.Type == client.ConditionTypeReady || condition.Type == client.ConditionTypeAvailable {
                        Expect(condition.ObservedGeneration).To(Equal(int32(1)),
                            "cluster condition %s should have observed_generation=1 for new creation request", condition.Type)
                    }
                }
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
