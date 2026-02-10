# Feature: Clusters Resource Type Lifecycle Management

## Table of Contents

1. [Clusters Resource Type - Workflow Validation](#test-title-clusters-resource-type---workflow-validation)
2. [Clusters Resource Type - K8s Resources Check Aligned with Preinstalled Clusters Related Adapters Specified](#test-title-clusters-resource-type---k8s-resources-check-aligned-with-preinstalled-clusters-related-adapters-specified)
3. [Clusters Resource Type - Adapter Dependency Relationships Workflow Validation for Preinstalled Clusters Related Dependent Adapters](#test-title-clusters-resource-type---adapter-dependency-relationships-workflow-validation-for-preinstalled-clusters-related-dependent-adapters)
---

## Test Title: Clusters Resource Type - Workflow Validation

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
  - `clusters-namespace`
  - `clusters-job`
  - `clusters-deployment`
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

## Test Title: Clusters Resource Type - K8s Resources Check Aligned with Preinstalled Clusters Related Adapters Specified

### Description

This test verifies that the Kubernetes resources (namespace and job) can be successfully created, aligned with the preinstalled adapters specified when submitting a clusters resource request.  

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Draft         |
| **Automation** | Not Automated |
| **Version** | MVP           |
| **Created** | 2026-01-29    |
| **Updated** | 2026-02-04    |


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

#### Step 2: Verify Kubernetes Resources Management
**Action:**
- Verify resource created

**Expected Result:**
- Resource is created successfully with templated values rendered

#### Step 3: Cleanup resources

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

## Test Title: Clusters Resource Type - Adapter Dependency Relationships Workflow Validation for Preinstalled Clusters Related Dependent Adapters

### Description

This test validates that CLM correctly executes workflows with preinstalled dependent adapters when submitting a clusters resource request. The test ensures adapter dependency relationships are honored, adapters execute in the correct order based on preconditions, and dependent adapters wait for prerequisite adapters to complete successfully. It also validates intermediate workflow states using job-based delay simulation.

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Draft         |
| **Automation** | Not Automated |
| **Version** | MVP           |
| **Created** | 2026-01-29    |
| **Updated** | 2026-02-04    |


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

#### Step 2: Verify execution order and state
**Action:**
- Send GET request to retrieve the specific cluster:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
At the intermediate workflow state (when adapter 2 is executing):
- The first adapter (namespace-adapter) has completed successfully
- The second adapter (job-adapter) is in progress
- The third adapter (job-dependency-adapter) waits for preconditions to be met
- Adapter status information shows:
    - Adapter 1 "message": "namespace is active and ready" and 'status' is 'true'
    - Adapter 2 "message": "job is progress" and 'status' is 'unknown'
    - Adapter 3 "message": "job-dependency is pending" and 'status' is 'false'

#### Step 3: Verify sequential execution and resource creation
**Action:**
- Verify Kubernetes resources: namespace is created
- Send GET request to retrieve the specific cluster:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```
- Monitor adapter execution progress

**Expected Result:**
- Adapter 1 creates Namespace
- After adapter 2 executes completely, adapter 3 is in progress.
- Adapters execute in correct order based on preconditions:
    1. Adapter 1 (namespace-creator) completes first
    2. Adapter 2 (workload) executes after adapter 1 completes
    3. Adapter 3 (another workload) executes after adapter 2 completes
- All three adapters report Ready status to API


#### Step 4: Verify final cluster state
**Action:**
- Send GET request to retrieve cluster statuses:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```

**Expected Result:**
- Final cluster conditions have `status: True` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

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
