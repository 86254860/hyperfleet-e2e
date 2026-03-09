# Feature: Clusters Resource Type Lifecycle Management

## Table of Contents

1. [Cluster can complete end-to-end workflow with all required adapters](#test-title-cluster-can-complete-end-to-end-workflow-with-all-required-adapters)
2. [Cluster adapters can create K8s resources with correct state and metadata](#test-title-cluster-adapters-can-create-k8s-resources-with-correct-state-and-metadata)
3. [Cluster adapters can enforce dependency order during workflow](#test-title-cluster-adapters-can-enforce-dependency-order-during-workflow)
4. [Cluster can reflect adapter failure in top-level status](#test-title-cluster-can-reflect-adapter-failure-in-top-level-status)
5. [API can reject cluster with invalid name format (RFC 1123)](#test-title-api-can-reject-cluster-with-invalid-name-format-rfc-1123)
6. [API can reject cluster with name exceeding max length](#test-title-api-can-reject-cluster-with-name-exceeding-max-length)

---

## Test Variables

```bash
# Required adapters from configs/config.yaml under adapters.cluster
export ADAPTER_NAMESPACE='cl-namespace'
export ADAPTER_JOB='cl-job'
export ADAPTER_DEPLOYMENT='cl-deployment'
```

---

## Test Title: Cluster can complete end-to-end workflow with all required adapters

### Description

This test validates that the workflow can work correctly for clusters resource type. It verifies that when a cluster resource is created via the HyperFleet API, the system correctly processes the resource through its lifecycle, required adapters (configured in the test config) execute successfully, and accurately reports status transitions back to the API. The test validates required adapters first to identify specific failures, then confirms the cluster reaches the final Ready and Available state. This approach ensures the complete workflow of CLM can successfully handle clusters resource type requests end-to-end.

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Automated     |
| **Automation** | Automated     |
| **Version** | MVP           |
| **Created** | 2026-01-29    |
| **Updated** | 2026-02-09    |


---

### Preconditions

1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully
3. The adapters defined in testdata/adapter-configs are all deployed successfully 

---

### Test Steps

#### Step 1: Submit an API request to create a Cluster resource

**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- Response includes the created cluster ID and initial metadata
- Initial cluster conditions have `status: False` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

#### Step 2: Verify initial status of cluster
**Action:**
- Poll cluster status for initial response
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

**Expected Result:**
- Cluster `Ready` condition `status: False`
- Cluster `Available` condition `status: False`

#### Step 3: Verify required adapter execution results

**Action:**
- Retrieve adapter statuses information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- Response returns HTTP 200 (OK) status code
- All required adapters from config are present in the response:
  - `${ADAPTER_NAMESPACE}`
  - `${ADAPTER_JOB}`
  - `${ADAPTER_DEPLOYMENT}`
- Each required adapter has all required condition types: `Applied`, `Available`, `Health`
- Each condition has `status: "True"` indicating successful execution
- **Adapter condition metadata validation** (for each condition in adapter.conditions):
  - `reason`: Non-empty string providing human-readable summary of the condition state
  - `message`: Non-empty string with detailed human-readable description
  - `last_transition_time`: Valid RFC3339 timestamp of the last status change
- **Adapter status metadata validation** (for each required adapter):
  - `created_time`: Valid RFC3339 timestamp when the adapter status was first created
  - `last_report_time`: Valid RFC3339 timestamp when the adapter last reported its status
  - `observed_generation`: Non-nil integer value equal to 1 for new creation requests

**Note:** Required adapters are configurable via:
- Config file: `configs/config.yaml` under `adapters.cluster`
- Environment variable: `HYPERFLEET_ADAPTERS_CLUSTER` (comma-separated list)

#### Step 4: Verify final cluster state

**Action:**
- Wait for cluster Ready condition to transition to True
- Retrieve final cluster status information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

**Expected Result:**
- Cluster `Ready` condition transitions from `status: False` to `status: True`
- Final cluster conditions have `status: True` for both condition `{"type": "Ready"}` and `{"type": "Available"}`
- Validate that the observedGeneration for the Ready and Available conditions is 1 for a new creation request
- This confirms the cluster has reached the desired end state

#### Step 5: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```

**Expected Result:**
- Namespace and all associated resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, this step should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: Cluster adapters can create K8s resources with correct state and metadata

### Description

This test verifies that Kubernetes resources are successfully created with correct templated values for all required cluster adapters. The test dynamically reads the list of required adapters from config, waits for each adapter to complete execution, then validates that corresponding Kubernetes resources (Namespace, Job, Deployment) exist with properly rendered metadata (labels, annotations) matching the cluster request payload. This ensures adapter Kubernetes resource management and templating work correctly across all configured adapters.

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Automated     |
| **Automation** | Automated     |
| **Version** | MVP           |
| **Created** | 2026-01-29    |
| **Updated** | 2026-02-11    |


---

### Preconditions

1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully 
3. The adapters defined in testdata/adapter-configs are all deployed successfully

---

### Test Steps

#### Step 1: Submit an API request to create a Cluster resource

**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- Response includes the created cluster ID and initial metadata
- Initial cluster conditions have `status: False` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

#### Step 2: Wait for all required adapters to complete

**Action:**
- Poll adapter statuses until all required adapters complete execution:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- All required adapters from config (${ADAPTER_NAMESPACE}, ${ADAPTER_JOB}, ${ADAPTER_DEPLOYMENT}) are present
- Each adapter has all three conditions (`Applied`, `Available`, `Health`) with `status: True`

**Note:** Required adapters are configurable via `configs/config.yaml` under `adapters.cluster`

#### Step 3: Verify Kubernetes resources for each adapter with correct metadata

**Action:**
- For each required adapter, retrieve and validate corresponding Kubernetes resources:

**For ${ADAPTER_NAMESPACE} adapter:**
```bash
kubectl get namespace {cluster_id} -o yaml
```

**Expected Result:**
- Namespace exists with name matching the cluster ID
- Namespace status phase is `Active`
- Required annotations:
  - `hyperfleet.io/generation`: Equals "1" for new creation request

**For ${ADAPTER_JOB} adapter:**
```bash
kubectl get job -n {cluster_id} -l hyperfleet.io/cluster-id={cluster_id},hyperfleet.io/resource-type=job -o yaml
```

**Expected Result:**
- Job exists in the cluster namespace, identified by the label selector
- Job has completed successfully (status.succeeded > 0 or status.conditions contains type=Complete with status=True)
- Required annotations:
  - `hyperfleet.io/generation`: Equals "1" for new creation request

**For ${ADAPTER_DEPLOYMENT} adapter:**
```bash
kubectl get deployment -n {cluster_id} -l hyperfleet.io/cluster-id={cluster_id},hyperfleet.io/resource-type=deployment -o yaml
```

**Expected Result:**
- Deployment exists in the cluster namespace, identified by the label selector
- Deployment is available (status.availableReplicas > 0 and status.conditions contains type=Available with status=True)
- Required annotations:
  - `hyperfleet.io/generation`: Equals "1" for new creation request

#### Step 4: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```

**Expected Result:**
- Namespace and all associated resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, this step should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: Cluster adapters can enforce dependency order during workflow

### Description

This test validates that CLM correctly handles adapter dependency relationships when processing a clusters resource request. Specifically, it verifies the dependency relationship where the deployment adapter (${ADAPTER_DEPLOYMENT}) depends on the job adapter (${ADAPTER_JOB}) completion. The test continuously polls and validates throughout the workflow period to ensure: (1) ${ADAPTER_DEPLOYMENT}'s Applied condition remains False until ${ADAPTER_JOB}'s Available condition reaches True, enforcing the dependency precondition; (2) during ${ADAPTER_JOB} execution, ${ADAPTER_DEPLOYMENT}'s Available condition stays Unknown (never False), confirming the adapter waits correctly without attempting execution; (3) successful completion with ${ADAPTER_DEPLOYMENT}'s Available eventually transitioning to True. This validation demonstrates that the workflow engine properly enforces adapter dependencies and ensures dependent adapters wait for prerequisites before executing.

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Automated     |
| **Automation** | Automated     |
| **Version** | MVP           |
| **Created** | 2026-01-29    |
| **Updated** | 2026-02-11    |


---

### Preconditions

1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully 
3. The adapters defined in testdata/adapter-configs are all deployed successfully

---

### Test Steps

#### Step 1: Submit an API request to create a Cluster resource
**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- API returns successful response

#### Step 2: Verify ${ADAPTER_DEPLOYMENT} initial state and dependency waiting behavior

**Action:**
- Poll adapter statuses to capture ${ADAPTER_DEPLOYMENT}'s initial waiting state:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
At the initial state (when ${ADAPTER_DEPLOYMENT} first appears in statuses):
- Response returns HTTP 200 (OK) status code
- The `${ADAPTER_DEPLOYMENT}` adapter is present with initial waiting state:
  - `Applied` condition has `status: "False"` (deployment hasn't been applied yet, waiting for ${ADAPTER_JOB} dependency)
  - `Available` condition has `status: "Unknown"` (deployment hasn't been applied yet)
  - `Health` condition has `status: "True"` (adapter itself is healthy, just waiting)

#### Step 3: Verify dependency relationship and condition transitions throughout entire workflow

**Action:**
- Continuously poll adapter statuses from the initial state until ${ADAPTER_DEPLOYMENT} completes:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
Throughout the entire period (from initial state until ${ADAPTER_DEPLOYMENT} completes), validate the following on each poll:

**Validation 1 - Dependency enforcement (during ${ADAPTER_JOB} execution):**
- While `${ADAPTER_JOB}` adapter's `Available` condition has NOT reached `status: "True"`:
  - The `${ADAPTER_DEPLOYMENT}` adapter's `Applied` condition must remain `status: "False"`
  - The `${ADAPTER_DEPLOYMENT}` adapter's `Available` condition must remain `status: "Unknown"` (never `status: "False"`)
  - This validates that ${ADAPTER_DEPLOYMENT} waits for ${ADAPTER_JOB} to complete without attempting to apply resources

**Validation 2 - Success condition:**
- Once `${ADAPTER_JOB}` adapter's `Available` reaches `status: "True"`, ${ADAPTER_DEPLOYMENT} can proceed with execution
- Once `${ADAPTER_DEPLOYMENT}` completes execution, its `Available` condition eventually becomes `status: "True"`
- This confirms the complete dependency workflow succeeded

**Note:** After ${ADAPTER_JOB} completes, ${ADAPTER_DEPLOYMENT}'s `Available` condition may temporarily be `False` (e.g., `MinimumReplicasUnavailable` during deployment startup) before becoming `True`, which is expected behavior and not validated.

#### Step 4: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```

**Expected Result:**
- Namespace and all associated resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, this step should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: Cluster can reflect adapter failure in top-level status

### Description

This test validates that the end-to-end workflow correctly handles adapter failure scenarios. When an adapter reports a failure status (e.g., `Health=False`), the cluster's top-level conditions (`Ready`, `Available`) should remain `False`, accurately reflecting that the cluster has not reached a healthy state.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | Tier1 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-03-04 |


---

### Preconditions

1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully
3. A dedicated failure-adapter is deployed via Helm with pre-configured failure behavior (e.g., `SIMULATE_RESULT=failure`), separate from the normal adapters used in other tests

---

### Test Steps

#### Step 1: Submit an API request to create a Cluster resource

**Action:**
- Submit a POST request to create a Cluster resource:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- API returns successful response with cluster ID

#### Step 2: Verify adapter failure is reported via status API

**Action:**
- Poll adapter statuses until the failure-adapter reports its status:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- The failure-adapter is present in the statuses response
- The failure-adapter reports `Health` condition with `status: "False"`, with reason and message indicating the failure
- The failure-adapter reports `Available` condition with `status: "False"`

#### Step 3: Verify cluster top-level status reflects adapter failure

**Action:**
- Retrieve cluster status:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

**Expected Result:**
- Cluster `Ready` condition remains `status: "False"`
- Cluster `Available` condition remains `status: "False"`
- Cluster does not transition to Ready state while any adapter reports failure

#### Step 4: Cleanup resources

**Action:**
- Delete the namespace created for this cluster:
```bash
kubectl delete namespace {cluster_id}
```
- Uninstall the failure-adapter Helm release

**Expected Result:**
- Namespace and all associated resources are deleted successfully
- Failure-adapter deployment is removed

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "clusters" resource type, the namespace deletion should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

---

## Test Title: API can reject cluster with invalid name format (RFC 1123)

### Description

This test validates that the HyperFleet API correctly rejects cluster creation requests with invalid name formats that don't comply with RFC 1123 DNS label naming conventions.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | Tier1 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet API is deployed and running successfully
2. API is accessible via port-forward or ingress

---

### Test Steps

#### Step 1: Send POST request with invalid name format
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "Invalid_Name_With_Underscore",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'
```

**Expected Result:**
- API returns HTTP 400 Bad Request
- Response contains validation error:
```json
{
  "type": "https://api.hyperfleet.io/errors/validation-error",
  "code": "HYPERFLEET-VAL-000",
  "title": "Validation Failed",
  "status": 400,
  "detail": "name must start and end with lowercase letter or number, and contain only lowercase letters, numbers, and hyphens"
}
```

---

## Test Title: API can reject cluster with name exceeding max length

### Description

This test validates that the HyperFleet API correctly rejects cluster creation requests with names exceeding the maximum allowed length.

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
1. HyperFleet API is deployed and running successfully

---

### Test Steps

#### Step 1: Send POST request with name exceeding max length
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "this-is-a-very-long-cluster-name-that-exceeds-the-maximum-allowed-length-for-cluster-names",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'
```

**Expected Result:**
- API returns HTTP 400 Bad Request
- Response contains validation error:
```json
{
  "type": "https://api.hyperfleet.io/errors/validation-error",
  "code": "HYPERFLEET-VAL-000",
  "title": "Validation Failed",
  "status": 400,
  "detail": "name must be at most 53 characters"
}
```

---
