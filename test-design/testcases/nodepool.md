# Feature: Nodepools Resource Type Lifecycle Management

## Table of Contents

1. [Nodepool can complete end-to-end workflow with required adapters](#test-title-nodepool-can-complete-end-to-end-workflow-with-required-adapters)
2. [Nodepool adapters can create K8s resources with correct metadata](#test-title-nodepool-adapters-can-create-k8s-resources-with-correct-metadata)
3. [API can reject nodepool creation with non-existent cluster](#test-title-api-can-reject-nodepool-creation-with-non-existent-cluster)
4. [API can reject nodepool with name exceeding 15 characters](#test-title-api-can-reject-nodepool-with-name-exceeding-15-characters)

---

## Test Title: Nodepool can complete end-to-end workflow with required adapters

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
| **Updated** | 2026-03-04    |


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
  -d @testdata/payloads/nodepools/gcp.json
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

#### Step 4: Verify nodepool appears in list by cluster

**Action:**
- Retrieve all nodepools belonging to the cluster:
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/nodepools | jq '.'
```

**Expected Result:**
- Response returns HTTP 200 (OK) status code
- Response contains an array of nodepools
- The created nodepool appears in the list with matching id, name, and cluster_id

#### Step 5: Cleanup resources

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

## Test Title: Nodepool adapters can create K8s resources with correct metadata

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
  -d @testdata/payloads/nodepools/gcp.json
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

## Test Title: API can reject nodepool creation with non-existent cluster

### Description

This test validates that the HyperFleet API correctly validates the existence of the parent cluster resource when creating a nodepool, returning HTTP 404 Not Found for non-existent clusters. This ensures proper resource hierarchy validation and prevents orphaned nodepool records.

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
1. HyperFleet API is deployed and running successfully
2. No cluster exists with the test cluster ID

---

### Test Steps

#### Step 1: Attempt to create nodepool with non-existent cluster ID
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/{fake_cluster_id}/nodepools \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "NodePool",
    "name": "np-test",
    "spec": {
      "nodeCount": 1,
      "platform": {
        "type": "gcp",
        "gcp": {
          "machineType": "n2-standard-4"
        }
      }
    }
  }'
```

**Expected Result:**
- API returns HTTP 404 Not Found
- Response follows RFC 9457 Problem Details format:
```json
{
  "type": "https://api.hyperfleet.io/errors/not-found",
  "code": "HYPERFLEET-NTF-001",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "Cluster with id='{fake_cluster_id}' not found"
}
```

---

#### Step 2: Verify error response format (RFC 9457)
**Action:**
- Parse the error response and verify required fields

**Expected Result:**
- Response contains `type` field
- Response contains `title` field
- Response contains `status` field with value 404
- Optional: Response contains `detail` field with descriptive message

---

#### Step 3: Verify no nodepool was created
**Action:**
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{fake_cluster_id}/nodepools
```

**Expected Result:**
- API returns HTTP 404 (cluster doesn't exist, so nodepools list is not accessible)
- OR returns empty list if API allows listing nodepools for non-existent clusters

---

## Test Title: API can reject nodepool with name exceeding 15 characters

### Description

This test validates that the HyperFleet API correctly rejects nodepool creation requests with names exceeding 15 characters.

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
2. A valid cluster exists

---

### Test Steps

#### Step 1: Send POST request with nodepool name exceeding 15 characters
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/nodepools \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "NodePool",
    "name": "this-is-too-long-nodepool-name",
    "spec": {"nodeCount": 1, "platform": {"type": "gcp", "gcp": {"machineType": "n2-standard-4"}}}
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
  "detail": "name must be at most 15 characters"
}
```

---
