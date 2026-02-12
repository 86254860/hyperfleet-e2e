# Feature: Concurrent Processing

## Table of Contents

1. [Concurrent cluster creation - no message loss](#test-title-concurrent-cluster-creation---no-message-loss)

---

## Test Title: Concurrent cluster creation - no message loss

### Description

This test validates that the system can handle multiple cluster creation requests submitted simultaneously without message loss, resource conflicts, or processing failures. It ensures that the message broker, Sentinel, and adapters can correctly process concurrent events and that all clusters reach their expected final state.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier1 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |


---

### Preconditions
1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API, Sentinel, and Adapter services are deployed and running successfully
3. The adapters defined in testdata/adapter-configs are all deployed successfully

---

### Test Steps

#### Step 1: Submit 5 cluster creation requests simultaneously
**Action:**
- Submit 5 POST requests in parallel using background processes:
```bash
for i in $(seq 1 5); do
  curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
    -H "Content-Type: application/json" \
    -d "{
      \"kind\": \"Cluster\",
      \"name\": \"concurrent-test-${i}\",
      \"spec\": {\"platform\": {\"type\": \"gcp\", \"gcp\": {\"projectID\": \"test\", \"region\": \"us-central1\"}}}
    }" &
done
wait
```

**Expected Result:**
- All 5 requests return successful responses (HTTP 200/201)
- Each response contains a unique cluster ID
- No request is rejected or fails due to concurrency

#### Step 2: Wait for all clusters to be processed
**Action:**
- Poll each cluster's status until all reach Ready state or a timeout is reached:
```bash
for CLUSTER_ID in ${CLUSTER_IDS[@]}; do
  curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID} | jq '.conditions'
done
```

**Expected Result:**
- All 5 clusters eventually reach Ready=True and Available=True
- No cluster is stuck in a pending or processing state indefinitely

#### Step 3: Verify Kubernetes resources for all clusters
**Action:**
- Check that each cluster has its own namespace and expected resources:
```bash
for CLUSTER_ID in ${CLUSTER_IDS[@]}; do
  kubectl get namespace ${CLUSTER_ID}
  kubectl get jobs -n ${CLUSTER_ID}
done
```

**Expected Result:**
- 5 separate namespaces exist (one per cluster)
- Each namespace contains the expected jobs/resources created by adapters
- No cross-contamination between clusters (resources are isolated)

#### Step 4: Verify adapter statuses for all clusters
**Action:**
- Check that each cluster has complete adapter status reports:
```bash
for CLUSTER_ID in ${CLUSTER_IDS[@]}; do
  curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.items | length'
done
```

**Expected Result:**
- Each cluster has the expected number of adapter status entries
- All adapters report Applied=True, Available=True, Health=True for each cluster
- No missing status reports (no message was lost)

#### Step 5: Cleanup
**Action:**
- Delete all 5 test clusters:
```bash
for CLUSTER_ID in ${CLUSTER_IDS[@]}; do
  curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}
done
```

**Expected Result:**
- All clusters are deleted successfully

---
