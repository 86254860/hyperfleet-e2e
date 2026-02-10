# Feature: Nodepools Resource Type Lifecycle Management

## Table of Contents

1. [Nodepools Resource Type - Workflow Validation](#test-title-nodepools-resource-type---workflow-validation)
2. [Nodepools Resource Type - K8s Resource Check Aligned with Preinstalled NodePool Related Adapters Specified](#test-title-nodepools-resource-type---k8s-resource-check-aligned-with-preinstalled-nodepool-related-adapters-specified)

---

## Test Title: Nodepools Resource Type - Workflow Validation

### Description

This test validates that the workflow can work correctly for nodepools resource type. It verifies that when a nodepool resource is created via the HyperFleet API, the system correctly processes the resource through its lifecycle, configured adapters execute successfully, and accurately reports status transitions back to the API. The test ensures the complete workflow of CLM can successfully handle nodepools resource type requests end-to-end.

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Draft         |
| **Automation** | Not Automated |
| **Version** | MVP           |
| **Created** | 2026-02-04    |
| **Updated** | 2026-02-05    |


---

### Preconditions

1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully
3. The adapters defined in testdata/adapter-configs are all deployed successfully
4. A cluster resource has been created and its cluster_id is available
    - **Cleanup**: Cluster resource cleanup should be handled in test suite teardown where cluster was created

---

### Test Steps

#### Step 1: Submit an API request to create a NodePool resource
**Action:**
- Submit a POST request to create a NodePool resource (with cluster_id in the payload):
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/nodepools \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/nodepools/nodepool-request.json
```

**Expected Result:**
- Response includes the created nodepool ID and initial metadata
- Initial nodepool conditions have `status: False` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

#### Step 2: Verify adapter status

**Action:**
- Retrieve adapter statuses information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/nodepools/{nodepool_id}/statuses
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
- `observed_generation`: Set to `1` for the initial nodepool generation

#### Step 3: Verify nodepool final status

**Action:**
- Retrieve nodepool status information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/nodepools/{nodepool_id}
```
**Expected Result:**
- Final nodepool conditions have `status: True` for both condition `{"type": "Ready"}` and `{"type": "Available"}`

#### Step 4: Cleanup resources

**Action:**
- Delete nodepool-specific Kubernetes resources:
```bash
kubectl delete -n {cluster_id} <nodepool-resources>
```

**Expected Result:**
- Nodepool-specific resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "nodepools" resource type, this step should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/nodepools/{nodepool_id}
```

---

## Test Title: Nodepools Resource Type - K8s Resource Check Aligned with Preinstalled NodePool Related Adapters Specified

### Description

This test verifies that the Kubernetes resources of different types (e.g., configmap) can be successfully created, aligned with the preinstalled adapters specified when submitting a nodepools resource request.

---

| **Field** | **Value**     |
|-----------|---------------|
| **Pos/Neg** | Positive      |
| **Priority** | Tier0         |
| **Status** | Draft         |
| **Automation** | Not Automated |
| **Version** | MVP           |
| **Created** | 2026-02-04    |
| **Updated** | 2026-02-05    |


---

### Preconditions

1. Environment is prepared using [hyperfleet-infra](https://github.com/openshift-hyperfleet/hyperfleet-infra) with all required platform resources
2. HyperFleet API and HyperFleet Sentinel services are deployed and running successfully
3. The adapters defined in testdata/adapter-configs are all deployed successfully
4. A cluster resource has been created and its cluster_id is available
    - **Cleanup**: Cluster resource cleanup should be handled in test suite teardown where cluster was created

---

### Test Steps

#### Step 1: Submit an API request to create a NodePool resource
**Action:**
- Submit a POST request to create a NodePool resource (with cluster_id in the payload):
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/nodepools \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/nodepools/nodepool-request.json
```

**Expected Result:**
- API returns successful response

#### Step 2: Verify Kubernetes Resources Management
**Action:**
- Verify resources created for different resource types (e.g., configmap)

**Expected Result:**
- Resources are created successfully with templated values rendered
- Multiple Kubernetes resource types are properly managed (e.g., configmap)

#### Step 3: Cleanup resources

**Action:**
- Delete nodepool-specific Kubernetes resources:
```bash
kubectl delete -n {cluster_id} <nodepool-resources>
```

**Expected Result:**
- Nodepool-specific resources are deleted successfully

**Note:** This is a workaround cleanup method. Once CLM supports DELETE operations for "nodepools" resource type, this step should be replaced with:
```bash
curl -X DELETE ${API_URL}/api/hyperfleet/v1/nodepools/{nodepool_id}
```

---
