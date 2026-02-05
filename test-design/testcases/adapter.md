# Feature: Adapter Framework - Customization

## Table of Contents

1. [Adapter framework can detect and report invalid adapter configuration failures in deploying progress](#test-title-adapter-framework-can-detect-and-report-invalid-adapter-configuration-failures-in-deploying-progress)
2. [Adapter framework can detect and report failures to cluster API endpoints](#test-title-adapter-framework-can-detect-and-report-failures-to-cluster-api-endpoints)
3. [Adapter framework can detect and handle resource timeouts](#test-title-adapter-framework-can-detect-and-handle-resource-timeouts)
4. [K8s object will be created correctly on targeted cluster with Maestro enabled](#test-title-k8s-object-will-be-created-correctly-on-targeted-cluster-with-maestro-enabled)

---
## Test Title: Adapter framework can detect and report invalid adapter configuration failures in deploying progress

### Description

This test validates that the adapter framework correctly detects and reports failures caused by invalid adapter configuration, including template rendering errors, malformed CloudEvents, invalid AdapterConfig structure, and CEL evaluation errors. It ensures proper error handling and status reporting for configuration-related issues.

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
- Configure AdapterConfig with invalid AdapterConfig (missing required fields, invalid field types)
- Deploy the test adapter

**Expected Result:**
- Adapter detects template rendering error
- Log reports failure with clear error message

#### Step 2: Test malformed CloudEvents
**Action:**
- Configure AdapterConfig with malformed CloudEvent (missing required fields, invalid JSON structure)
- Deploy the test adapter

**Expected Result:**
- Adapter rejects malformed CloudEvent
- Log reports CloudEvent validation error


#### Step 3: Test CEL evaluation errors
**Action:**
- Configure AdapterConfig with invalid CEL expressions in payload construction
- Create cluster to trigger CEL evaluation

**Expected Result:**
- Adapter detects CEL evaluation error
- Error message includes the failing expression

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
     | jq -r '.items[] | select(.adapter=="<adapter_name>") | .conditions[] | select(.type=="Applied")'

   # Example:
   # {
   #   "type": "Available",
   #   "status": "False",
   #   "reason": "`invalid k8s object` resource is invalid",
   #   "message": "Invalid Kubernetes object"
   # }
```

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
     | jq -r '.items[] | select(.adapter=="<adapter_name>") | .conditions[] | select(.type=="Applied")'

   # Example:
   # {
   #   "type": "Applied",
   #   "status": "False",
   #   "reason": "JobTimeout",
   #   "message": "Validation job did not complete within 30 seconds"
   # }
```
-----

## Test Title: K8s object will be created correctly on targeted cluster with Maestro enabled

### Description

This test validates that the adapter framework creates ManifestWork resources correctly when Maestro is enabled, allowing resources to be propagated to targeted managed clusters. It verifies ManifestWork generation, structure, and successful delivery to target clusters.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier0 |
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

#### Step 1: Configure adapter with Maestro enabled
**Action:**
- Configure AdapterConfig with Maestro enable. See full example: [adapter-business-logic-template-MVP.yaml](https://github.com/openshift-hyperfleet/architecture/blob/main/hyperfleet/components/adapter/framework/configs/adapter-business-logic-template-MVP.yaml)
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

#### Step 3: Verify ManifestWork
**Action:**
- Verify ManifestWork

**Expected Result:**
- ManifestWork is created successfully with templated values rendered

#### Step 4: Verify Kubernetes Resources Management
**Action:**
- Verify configured resources created

**Expected Result:**
- Resource is created successfully with templated values rendered

#### Step 5: Verify adapter status

**Action:**
- Retrieve adapter statuses information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
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

#### Step 6: Verify cluster final status

**Action:**
- Retrieve cluster status information:
```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}
```
**Expected Result:**
- Final cluster conditions have `status: True` for both condition `{"type": "Ready"}` and `{"type": "Available"}`
---
