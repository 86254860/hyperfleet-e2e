# Feature: API Input Validation

## Table of Contents

1. [API validates cluster name format (RFC 1123)](#test-title-api-validates-cluster-name-format-rfc-1123)
2. [API validates cluster name length](#test-title-api-validates-cluster-name-length)
3. [API validates nodepool name length (15 characters)](#test-title-api-validates-nodepool-name-length-15-characters)
4. [API validates required spec field](#test-title-api-validates-required-spec-field)
5. [API validates Kind field](#test-title-api-validates-kind-field)
6. [API validates JSON format](#test-title-api-validates-json-format)
7. [API handles database connection failure gracefully](#test-title-api-handles-database-connection-failure-gracefully)

---

## Test Title: API validates cluster name format (RFC 1123)

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
  "code": "HYPERFLEET-VAL-000",
  "detail": "name must match pattern ^[a-z0-9]([-a-z0-9]*[a-z0-9])?$",
  "status": 400,
  "title": "Validation Failed"
}
```

---

## Test Title: API validates cluster name length

### Description

This test validates that the HyperFleet API correctly rejects cluster creation requests with names exceeding the maximum allowed length.

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
- Response contains validation error about name length

---

## Test Title: API validates nodepool name length (15 characters)

### Description

This test validates that the HyperFleet API correctly rejects nodepool creation requests with names exceeding 15 characters.

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
2. A valid cluster exists

---

### Test Steps

#### Step 1: Send POST request with nodepool name exceeding 15 characters
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/nodepools \
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
  "code": "HYPERFLEET-VAL-000",
  "detail": "name must be at most 15 characters",
  "status": 400,
  "title": "Validation Failed"
}
```

---

## Test Title: API validates required spec field

### Description

This test validates that the HyperFleet API correctly rejects cluster creation requests with missing or null spec field, as defined in the OpenAPI specification.

**Note:** This test currently FAILS - see Issue #10.

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

---

### Test Steps

#### Step 1: Send POST request without spec field
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "test-no-spec"
  }'
```

**Expected Result:**
- API returns HTTP 400 Bad Request
- Response contains validation error:
```json
{
  "code": "HYPERFLEET-VAL-000",
  "detail": "spec is required",
  "status": 400,
  "title": "Validation Failed"
}
```

**Actual Result (Bug - Issue #10):**
- API returns HTTP 200 and creates cluster with `spec: null`

---

## Test Title: API validates Kind field

### Description

This test validates that the HyperFleet API correctly rejects requests with invalid Kind field values.

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

---

### Test Steps

#### Step 1: Send POST request with wrong Kind value
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "WrongKind",
    "name": "test-wrong-kind",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'
```

**Expected Result:**
- API returns HTTP 400 Bad Request
- Response contains validation error about invalid Kind

---

## Test Title: API validates JSON format

### Description

This test validates that the HyperFleet API correctly rejects requests with malformed JSON.

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

---

### Test Steps

#### Step 1: Send POST request with invalid JSON
**Action:**
```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{invalid json content'
```

**Expected Result:**
- API returns HTTP 400 Bad Request
- Response contains JSON parsing error

---

## Test Title: API handles database connection failure gracefully

### Description

This test validates that the API handles database connection failures gracefully for cluster and nodepool operations, ensuring proper error responses (HTTP 503), no data corruption, and automatic recovery when the connection is restored.

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
2. A cluster has been created and is in Ready state
3. Ability to simulate database connection failure (e.g., stop PostgreSQL pod)
4. Ability to restore database connection

---

### Test Steps

#### Step 1: Establish baseline - verify normal operations
**Action:**
- Verify API is functioning correctly before simulating failure:
```bash
# Verify cluster operations
curl -s ${API_URL}/api/hyperfleet/v1/clusters | jq '.total'

# Verify nodepool operations (using a known cluster ID)
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/nodepools | jq '.total'
```

**Expected Result:**
- API returns HTTP 200 with valid cluster list
- API returns HTTP 200 with valid nodepool list
- Record baseline data for comparison after recovery

#### Step 2: Simulate database connection failure
**Action:**
- Stop or disrupt the PostgreSQL pod:
```bash
kubectl scale deployment -n hyperfleet hyperfleet-hyperfleet-api-postgresql --replicas=0
```

**Expected Result:**
- PostgreSQL pod is terminated

#### Step 3: Verify API returns proper error during outage
**Action:**
- Attempt API operations during database outage:
```bash
# List clusters
curl -s -w "\n%{http_code}" ${API_URL}/api/hyperfleet/v1/clusters

# Create cluster
curl -s -w "\n%{http_code}" -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "db-failure-test",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'

# List nodepools (using a known cluster ID from baseline)
curl -s -w "\n%{http_code}" ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/nodepools

# Create nodepool
curl -s -w "\n%{http_code}" -X POST ${API_URL}/api/hyperfleet/v1/nodepools \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "NodePool",
    "name": "db-fail-np",
    "cluster_id": "'${CLUSTER_ID}'",
    "spec": {"nodeCount": 1, "platform": {"type": "gcp", "gcp": {"machineType": "n2-standard-4"}}}
  }'
```

**Expected Result:**
- All operations (cluster and nodepool) return HTTP 503 (Service Unavailable)
- API process does not crash
- No partial data is written
- API service remains running and responsive (returns errors, does not hang)

#### Step 4: Restore database connection
**Action:**
- Restore the PostgreSQL pod:
```bash
kubectl scale deployment -n hyperfleet hyperfleet-hyperfleet-api-postgresql --replicas=1
kubectl rollout status deployment/hyperfleet-hyperfleet-api-postgresql -n hyperfleet --timeout=60s
```
- Wait for API to reconnect

**Expected Result:**
- PostgreSQL pod starts successfully
- API reconnects to the database automatically

#### Step 5: Verify recovery and data integrity
**Action:**
- Verify normal operations resume:
```bash
# Verify cluster operations
curl -s ${API_URL}/api/hyperfleet/v1/clusters | jq '.total'

# Verify nodepool operations
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/nodepools | jq '.total'
```

**Expected Result:**
- API returns HTTP 200 with valid data for both cluster and nodepool operations
- All baseline clusters and nodepools are intact (no data loss or corruption)
- No orphaned records from failed operations during outage
- End-to-end workflow functions correctly (create new cluster and verify it reaches Ready state)

---
