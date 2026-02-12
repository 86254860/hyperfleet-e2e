# Feature: NodePool Lifecycle - Negative Tests

## Table of Contents

1. [Create nodepool with non-existent cluster returns 404](#test-title-create-nodepool-with-non-existent-cluster-returns-404)

---

## Test Title: Create nodepool with non-existent cluster returns 404

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
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet API is deployed and running successfully
2. No cluster exists with the test cluster ID

---

### Test Steps

#### Step 1: Attempt to create nodepool with non-existent cluster ID
**Action:**
```bash
FAKE_CLUSTER_ID="non-existent-cluster-12345"

curl -X POST ${API_URL}/api/hyperfleet/v1/clusters/${FAKE_CLUSTER_ID}/nodepools \
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
  "title": "Not Found",
  "status": 404,
  "detail": "Cluster not found"
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
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/${FAKE_CLUSTER_ID}/nodepools
```

**Expected Result:**
- API returns HTTP 404 (cluster doesn't exist, so nodepools list is not accessible)
- OR returns empty list if API allows listing nodepools for non-existent clusters

---

#### Step 4: Verify no Sentinel events triggered
**Action:**
- Check that no namespace was created for the fake cluster
```bash
kubectl get namespace ${FAKE_CLUSTER_ID}
```

**Expected Result:**
- Namespace does not exist
- No adapter processing occurred (validation-level rejection)

---
