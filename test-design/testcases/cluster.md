# Feature: Clusters Resource Type Lifecycle Management

## Table of Contents

1. [Clusters Resource Type - Workflow Validation](#test-title-clusters-resource-type---workflow-validation)
2. [Clusters Resource Type - K8s Resources Check Aligned with Preinstalled Clusters Related Adapters Specified](#test-title-clusters-resource-type---k8s-resources-check-aligned-with-preinstalled-clusters-related-adapters-specified)
3. [Clusters Resource Type - Adapter Dependency Relationships Workflow Validation for Preinstalled Clusters Related Dependent Adapters](#test-title-clusters-resource-type---adapter-dependency-relationships-workflow-validation-for-preinstalled-clusters-related-dependent-adapters)
---

## Test Title: Clusters Resource Type - Workflow Validation

### Description

This test validates that the workflow can work correctly for clusters resource type. It verifies that when a cluster resource is created via the HyperFleet API, the system correctly processes the resource through its lifecycle, configured adapters execute successfully, and accurately reports status transitions back to the API. The test ensures the complete workflow of CLM can successfully handle clusters resource type requests end-to-end.

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

#### Step 1: Submit a "clusters" resource type request via API
**Action:**
- Submit a POST request for "clusters" resource type:
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/cluster-request.json
```

**Expected Result:**
- Response includes the created cluster ID and initial metadata
- Initial cluster conditions have `status: False` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

#### Step 2: Verify adapter status

**Action:**
- Retrieve adapter statuses information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- Response returns HTTP 200 (OK) status code
- Adapter status payload contains the following:

**Condition Types:**
- All required condition types are present: `Applied`, `Available`, `Health`
- Each condition has `status: "True"` when successful
- `reason`: Human-readable summary of the condition state
- `message`: Detailed human-readable description
- `created_time`: Timestamp when the condition was first created
- `last_transition_time`: Timestamp of the last status change
- `last_updated_time`: Timestamp of the most recent update
- `observed_generation`: Set to `1` for the initial cluster generation

#### Step 3: Verify cluster final status

**Action:**
- Retrieve cluster status information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```
**Expected Result:**
- Final cluster conditions have `status: True` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

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

#### Step 1: Submit a "clusters" resource type request via API
**Action:**
- Execute cluster creation request:
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

#### Step 1: Submit a "clusters" resource type request via API
**Action:**
- Execute cluster creation request:
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
- Namespace is created by adapter 1
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
