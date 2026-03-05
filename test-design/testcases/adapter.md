# Feature: Adapter Framework - Customization

## Table of Contents

1. [Adapter can detect and report invalid K8s resource failures](#test-title-adapter-can-detect-and-report-invalid-k8s-resource-failures)
2. [Adapter can detect and handle precondition timeouts](#test-title-adapter-can-detect-and-handle-precondition-timeouts)
3. [Adapter can recover from crash and process redelivered events](#test-title-adapter-can-recover-from-crash-and-process-redelivered-events)
4. [Adapter can process pending events after restart](#test-title-adapter-can-process-pending-events-after-restart)
5. [Adapter can handle duplicate events for same cluster idempotently](#test-title-adapter-can-handle-duplicate-events-for-same-cluster-idempotently)
6. [API can handle incomplete adapter status reports gracefully](#test-title-api-can-handle-incomplete-adapter-status-reports-gracefully)

---

## Test Title: Adapter can detect and report invalid K8s resource failures

### Description

This test validates that the adapter framework correctly detects and reports failures when attempting to create invalid Kubernetes resources on the target cluster. It ensures that when an adapter's configuration contains invalid K8s resource objects, the framework properly handles the API server rejection and reports the failure status back to the HyperFleet API with appropriate condition states and error details.


---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-01-30 |
| **Updated** | 2026-01-30 |


---

### Preconditions
1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully
3. A dedicated test adapter is deployed via Helm with AdapterConfig containing invalid K8s resource objects

---

### Test Steps

#### Step 1: Send POST request to create a new cluster
**Action:**
- Execute cluster creation request:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- API returns successful response

#### Step 2: Verify adapter status reports failure
**Action:**
- Poll adapter statuses:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- The test adapter reports `Available` condition with `status: "False"`, with reason indicating invalid K8s resource

#### Step 3: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```
- Uninstall the test adapter Helm release

**Expected Result:**
- Namespace and all associated resources are deleted successfully
- Test adapter deployment is removed

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: Adapter can detect and handle precondition timeouts

### Description

This test validates that the adapter framework correctly detects and handles resource timeouts when adapter Jobs exceed configured timeout limits.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-01-30 |
| **Updated** | 2026-01-30 |


---

### Preconditions
1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully
3. A dedicated timeout-adapter is deployed via Helm with AdapterConfig containing preconditions that cannot be met, for example:
```yaml
preconditions:
  - name: "clusterStatus"
    apiCall:
      method: "GET"
      url: "{{ .hyperfleetApiBaseUrl }}/api/hyperfleet/{{ .hyperfleetApiVersion }}/clusters/{{ .clusterId }}"
      timeout: 10s
      retryAttempts: 3
      retryBackoff: "exponential"
    capture:
      - name: "clusterName"
        field: "name"
      - name: "clusterPhase"
        field: "status.phase"
      - name: "generationId"
        field: "generation"
    conditions:
      - field: "clusterPhase"
        operator: "in"
        values: ["NotReady", "Ready"]
```

---

### Test Steps

#### Step 1: Send POST request to create a new cluster
**Action:**
- Execute cluster creation request:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- API returns successful response

#### Step 2: Verify adapter status reports timeout
**Action:**
- Poll adapter statuses:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- The timeout-adapter reports `Available` condition with `status: "False"`, with reason indicating timeout (e.g., `"reason": "JobTimeout"`)

#### Step 3: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```
- Uninstall the timeout-adapter Helm release

**Expected Result:**
- Namespace and all associated resources are deleted successfully
- Timeout-adapter deployment is removed

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---


## Test Title: Adapter can recover from crash and process redelivered events

### Description

This test validates that when an adapter crashes during event processing, the system ensures that pending events are eventually processed after the adapter recovers. This ensures that no events are lost due to adapter failures and the system maintains eventual consistency.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-03-04 |


---

### Preconditions
1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra)
2. HyperFleet API, Sentinel, and Adapter services are deployed
3. A dedicated crash-adapter is deployed via Helm with pre-configured crash behavior (`SIMULATE_RESULT=crash`), separate from the normal adapters used in other tests
4. Message broker is configured with appropriate acknowledgment deadline and retry policy

---

### Test Steps

#### Step 1: Create a cluster to trigger event
**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- API returns successful response with cluster ID

**Note:** After the cluster is created, Sentinel will detect the new cluster during its polling cycle and publish an event to the broker, which triggers the crash-adapter to receive and process the event.

#### Step 2: Verify crash-adapter crashes on event receipt
**Action:**
- Monitor crash-adapter pod status:
```bash
kubectl get pods -n hyperfleet -l app.kubernetes.io/instance=crash-adapter -w
```

**Expected Result:**
- crash-adapter pod crashes (CrashLoopBackOff or Error state)

**Note:** The unacknowledged message will be redelivered by the broker, which is verified in Step 4.

#### Step 3: Restore crash-adapter to normal mode
**Action:**
- Upgrade crash-adapter Helm release with `SIMULATE_RESULT=success`

**Expected Result:**
- crash-adapter pod starts and remains Running

#### Step 4: Verify message redelivery and processing
**Action:**
- Wait for crash-adapter to process redelivered message
- Check cluster status:
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses | jq '.'
```

**Expected Result:**
- crash-adapter eventually processes the cluster event via redelivered message
- Status is reported with appropriate conditions
- No event is lost

#### Step 5: Cleanup resources
**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```
- Uninstall the crash-adapter Helm release

**Expected Result:**
- Namespace and all associated resources are deleted successfully
- crash-adapter deployment is removed

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

### Notes

- Message redelivery behavior depends on the message broker configuration:
  - For Google Pub/Sub: Uses acknowledgment deadline and retry policy
  - If adapter crashes before acknowledging message, broker will redeliver after the acknowledgment deadline expires
- Sentinel may also republish events during its polling cycle if generation > observed_generation

---

## Test Title: Adapter can process pending events after restart

### Description

This test validates that adapters can recover from restarts and continue processing pending events. Pending events should be eventually processed after the adapter restarts.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-03-04 |


---

### Preconditions
1. HyperFleet system is deployed and running
2. Adapter is running normally

---

### Test Steps

#### Step 1: Create cluster and verify initial processing
**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```
- Wait for adapter to process and verify status:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- Cluster created and processed
- Adapter status shows Available=True

---

#### Step 2: Restart adapter pod
**Action:**
- Delete an adapter pod to trigger restart:
```bash
kubectl delete pod -n hyperfleet -l app.kubernetes.io/instance=<adapter-release-name>
```

**Expected Result:**
- Adapter pod is terminated
- New adapter pod starts up automatically

---

#### Step 3: Create new cluster while adapter is restarting
**Action:**
- Create another cluster via API during adapter restart window:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- Cluster created successfully (API is independent of adapter)

---

#### Step 4: Verify adapter processes pending events after restart
**Action:**
- Wait for adapter to fully restart
- Check status of the new cluster:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- Both clusters have adapter statuses with Applied=True, Available=True

#### Step 5: Cleanup resources
**Action:**
- Delete the namespaces created for both clusters:
```bash
kubectl delete namespace {cluster_id_1}
kubectl delete namespace {cluster_id_2}
```

**Expected Result:**
- Namespaces and all associated resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: Adapter can handle duplicate events for same cluster idempotently

### Description

This test validates that when multiple events are published for the same cluster (e.g., due to retries or rapid updates), the adapter handles them idempotently without creating duplicate resources.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier1 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-03-04 |


---

### Preconditions
1. HyperFleet system is deployed and running
2. Adapter uses discovery pattern to check existing resources

---

### Test Steps

#### Step 1: Create cluster and wait for initial processing
**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```
- Wait for adapter to process and verify status:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- Cluster created
- Adapter creates namespace and other resources
- Status reported

---

#### Step 2: Trigger multiple events for same cluster
**Action:**
- Restart Sentinel to trigger re-polling:
```bash
kubectl delete pod -n hyperfleet -l app.kubernetes.io/name=sentinel
```

**Expected Result:**
- Multiple events may be published for the same cluster

---

#### Step 3: Verify no duplicate resources created
**Action:**
```bash
# Check namespace count (should be exactly 1)
kubectl get namespace {cluster_id}

# Check job count (should be 1 or managed correctly)
kubectl get jobs -n {cluster_id}
```

**Expected Result:**
- Only one namespace exists for the cluster
- No duplicate jobs or resources

---

#### Step 4: Verify status is consistent
**Action:**
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses | jq '.items | length'
```

**Expected Result:**
- Only one status entry per adapter
- Status reflects latest state (not duplicated)

#### Step 5: Cleanup resources
**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```

**Expected Result:**
- Namespace and all associated resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: API can handle incomplete adapter status reports gracefully

### Description

This test validates that the HyperFleet API can gracefully handle adapter status reports that are missing expected fields. When an adapter reports status with incomplete or missing condition fields (e.g., missing `reason`, `message`, or `observed_generation`), the API should accept the report without crashing and store what is available, rather than rejecting the entire status update.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |


---

### Preconditions
1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API is deployed and running successfully
3. A cluster resource has been created and its cluster_id is available

---

### Test Steps

#### Step 1: Submit a status report with missing optional fields
**Action:**
- Send a status report with minimal fields (missing `reason`, `message`):
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses \
  -H "Content-Type: application/json" \
  -d '{
    "adapter": "test-incomplete-adapter",
    "conditions": [
      {
        "type": "Applied",
        "status": "True"
      }
    ]
  }'
```

**Expected Result:**
- API accepts the status report (HTTP 200/201)
- API does not return an error or crash
- Status is stored with available fields; missing fields default to empty or null

#### Step 2: Submit a status report with missing observed_generation
**Action:**
- Send a status report without `observed_generation`:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses \
  -H "Content-Type: application/json" \
  -d '{
    "adapter": "test-no-generation-adapter",
    "conditions": [
      {
        "type": "Applied",
        "status": "True",
        "reason": "ResourceApplied",
        "message": "Resource applied successfully"
      }
    ]
  }'
```

**Expected Result:**
- API accepts the status report
- `observed_generation` defaults to 0 or null
- Cluster conditions are not corrupted

#### Step 3: Submit a status report with empty conditions array
**Action:**
- Send a status report with an empty conditions list:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses \
  -H "Content-Type: application/json" \
  -d '{
    "adapter": "test-empty-conditions-adapter",
    "conditions": []
  }'
```

**Expected Result:**
- API either accepts the report with empty conditions, or returns a clear validation error (HTTP 400)
- API does not crash or return HTTP 500

#### Step 4: Verify cluster state is not corrupted
**Action:**
- Retrieve the cluster and its statuses:
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id} | jq '.conditions'
curl -s ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses | jq '.'
```

**Expected Result:**
- Cluster conditions remain consistent and valid
- Previously reported statuses from other adapters are not affected
- Incomplete status entries are stored but do not interfere with cluster Ready/Available evaluation

#### Step 5: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```

**Expected Result:**
- Namespace and all associated resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---
