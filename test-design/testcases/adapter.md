# Feature: Adapter Framework - Customization

## Table of Contents

1. [Adapter framework can detect and report failures to cluster API endpoints](#test-title-adapter-framework-can-detect-and-report-failures-to-cluster-api-endpoints)
2. [Adapter Job failure correctly reports Health=False](#test-title-adapter-job-failure-correctly-reports-healthfalse)
3. [Adapter framework can detect and handle resource timeouts](#test-title-adapter-framework-can-detect-and-handle-resource-timeouts)
4. [Adapter crash recovery and message redelivery](#test-title-adapter-crash-recovery-and-message-redelivery)
5. [API handles incomplete adapter status reports gracefully](#test-title-api-handles-incomplete-adapter-status-reports-gracefully)

---

## Test Title: Adapter framework can detect and report failures to cluster API endpoints

### Description

This test validates that the adapter framework correctly detects and reports failures when attempting to create invalid Kubernetes resources on the target cluster. It ensures that when an adapter's configuration contains invalid K8s resource objects, the framework properly handles the API server rejection, logs meaningful error messages, and reports the failure status back to the HyperFleet API with appropriate condition states and error details.


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

---

### Test Steps

#### Step 1: Test template rendering errors
**Action:**
- Configure AdapterConfig with invalid AdapterConfig (invalid K8s resource object)
- Deploy the test adapter

**Expected Result:**
- Adapter detects template rendering error
- Log reports failure with clear error message

#### Step 2: Send POST request to create a new cluster
**Action:**
- Execute cluster creation request:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d <cluster_create_payload>
```

**Expected Result:**
- API returns successful response

#### Step 3: Wait for timeout and Verify Timeout Handling
**Action:**
- Wait for some minutes
- Verify adapter status

**Expected Result:**
```bash
   curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/<cluster_id>/statuses \
     | jq -r '.items[] | select(.adapter=="<adapter_name>") | .conditions[] | select(.type=="Available")'

   # Example:
   # {
   #   "type": "Available",
   #   "status": "False",
   #   "reason": "`invalid k8s object` resource is invalid",
   #   "message": "Invalid Kubernetes object"
   # }
```

---


## Test Title: Adapter Job failure correctly reports Health=False

### Description

This test validates that when an adapter Job runs but exits with a non-zero exit code (simulated via `SIMULATE_RESULT=failure`), the adapter framework correctly detects the job failure and reports `Health=False` back to the HyperFleet API. This is distinct from invalid K8s resource errors — this tests the scenario where the Job is created and scheduled successfully but the workload itself fails.

**Note:** There are two distinct adapter failure modes that should be differentiated:
- **Job execution failure** (this test): Job is created successfully (`Applied=True`) but the workload fails (`Health=False`). The `data` field should contain error details (exit code, error message).
- **Job creation failure** (covered by test 1 "Adapter framework can detect and report failures"): The Job or K8s resource cannot be created at all (`Applied=False`, `Health=False`).

Both failure modes should be verified for Cluster and NodePool resource types, as the adapter framework logic applies to both.

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
2. HyperFleet API, Sentinel, and Adapter services are deployed and running successfully
3. Example-adapter supports `SIMULATE_RESULT=failure` environment variable

---

### Test Steps

#### Step 1: Record original adapter configuration and set failure mode
**Action:**
- Record current `SIMULATE_RESULT` value for restoration
- Set adapter to failure mode:
```bash
kubectl set env deployment -n hyperfleet -l app.kubernetes.io/instance=example-adapter SIMULATE_RESULT=failure
kubectl rollout status deployment/example-adapter-hyperfleet-adapter -n hyperfleet --timeout=60s
```

**Expected Result:**
- Adapter deployment updated and new pod is Running

#### Step 2: Create a cluster to trigger adapter processing
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "failure-test",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'
```

**Expected Result:**
- API returns successful response with cluster ID

#### Step 3: Wait for job execution and verify job failure
**Action:**
- Wait for adapter to process the event and job to run
- Check job status:
```bash
kubectl get jobs -n ${CLUSTER_ID} -o json | jq '.items[0].status.conditions[]? | select(.type=="Failed")'
```

**Expected Result:**
- Job exists in the cluster namespace
- Job status shows `Failed=True`
- Job exit code is non-zero

#### Step 4: Verify adapter status report
**Action:**
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses \
  | jq '.items[0].conditions'
```

**Expected Result:**
- `Applied` condition has `status: "True"` (Job was created successfully before failing)
- `Available` condition has `status: "False"` (work not completed; currently returns empty string — see Issue #25)
- `Health` condition has `status: "False"` with reason indicating job failure
- `data` field contains error details:
  - Exit code (non-zero)
  - Error message describing the failure

#### Step 5: Verify top-level resource status reflects failure
**Action:**
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID} | jq '.status.conditions'
```

**Expected Result:**
- `Ready` condition remains `status: "False"` (cluster did not reach ready state)
- `Available` condition remains `status: "False"`
- Cluster does not transition to Ready while adapter reports failure

#### Step 6: Restore adapter to normal mode
**Action:**
```bash
kubectl set env deployment -n hyperfleet -l app.kubernetes.io/instance=example-adapter SIMULATE_RESULT=success
kubectl rollout status deployment/example-adapter-hyperfleet-adapter -n hyperfleet --timeout=60s
```

**Expected Result:**
- Adapter restored to normal operation

---

## Test Title: Adapter framework can detect and handle resource timeouts

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

---

### Test Steps

#### Step 1: Configure adapter with timeout setting
**Action:**
- Configure AdapterConfig with non-existed conditions that can't meet the precondition
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
- Deploy the test adapter

**Expected Result:**
- Adapter loads configuration successfully
- Adapter pods are running successfully
- Adapter logs show successful initialization

#### Step 2: Send POST request to create a new cluster
**Action:**
- Execute cluster creation request:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d <cluster_create_payload>
```

**Expected Result:**
- API returns successful response

#### Step 3: Wait for timeout and Verify Timeout Handling
**Action:**
- Wait for some minutes
- Verify adapter status

**Expected Result:**
```bash
   curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/<cluster_id>/statuses \
     | jq -r '.items[] | select(.adapter=="<adapter_name>") | .conditions[] | select(.type=="Available")'

   # Example:
   # {
   #   "type": "Available",
   #   "status": "False",
   #   "reason": "JobTimeout",
   #   "message": "Validation job did not complete within 30 seconds"
   # }
```

---


## Test Title: Adapter crash recovery and message redelivery

### Description

This test validates that when an adapter crashes during event processing, the message broker correctly redelivers the message after the adapter recovers. This ensures that no events are lost due to adapter failures and the system maintains eventual consistency.

**Note:** This test requires SIMULATE_RESULT=crash mode in the example-adapter, which may not be implemented.

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
1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra)
2. HyperFleet API, Sentinel, and Adapter services are deployed
3. Google Pub/Sub is configured with appropriate acknowledgment deadline and retry policy

---

### Test Steps

#### Step 1: Configure adapter to crash on event receipt
**Action:**
- Set adapter environment variable to crash mode:
```bash
kubectl set env deployment -n hyperfleet -l app.kubernetes.io/instance=example-adapter SIMULATE_RESULT=crash
kubectl rollout status deployment/example-adapter-hyperfleet-adapter -n hyperfleet
```

**Expected Result:**
- Adapter deployment updated
- New adapter pod starts

#### Step 2: Create a cluster to trigger event
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "crash-test",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'
```

**Expected Result:**
- API returns successful response with cluster ID
- Sentinel publishes event to broker

#### Step 3: Verify adapter crashes on event receipt
**Action:**
- Monitor adapter pod status:
```bash
kubectl get pods -n hyperfleet -l app.kubernetes.io/instance=example-adapter -w
```

**Expected Result:**
- Adapter pod crashes (CrashLoopBackOff or Error state)
- Pod restarts automatically

#### Step 4: Restore adapter to normal mode
**Action:**
```bash
kubectl set env deployment -n hyperfleet -l app.kubernetes.io/instance=example-adapter SIMULATE_RESULT=success
kubectl rollout status deployment/example-adapter-hyperfleet-adapter -n hyperfleet
```

**Expected Result:**
- Adapter deployment updated
- New adapter pod starts and remains Running

#### Step 5: Verify message redelivery and processing
**Action:**
- Wait for adapter to process redelivered message
- Check cluster status:
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.'
```

**Expected Result:**
- Adapter eventually processes the cluster event
- Status is reported (may show Applied=True, Available progressing)
- No event is lost

---

### Notes

- Message redelivery behavior depends on Google Pub/Sub configuration:
  - Uses acknowledgment deadline and retry policy
  - If adapter crashes before acknowledging message, Pub/Sub will redeliver after the acknowledgment deadline expires
- Sentinel may also republish events during its polling cycle if generation > observed_generation

---

## Test Title: API handles incomplete adapter status reports gracefully

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
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses \
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
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses \
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
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses \
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
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID} | jq '.conditions'
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.'
```

**Expected Result:**
- Cluster conditions remain consistent and valid
- Previously reported statuses from other adapters are not affected
- Incomplete status entries are stored but do not interfere with cluster Ready/Available evaluation

---
