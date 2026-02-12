# Feature: Asynchronous Processing and Generation Tracking

## Table of Contents

1. [Generation update triggers new adapter processing](#test-title-generation-update-triggers-new-adapter-processing)
2. [Sentinel detects and publishes generation changes](#test-title-sentinel-detects-and-publishes-generation-changes)
3. [Adapter tracks observed_generation correctly](#test-title-adapter-tracks-observed_generation-correctly)
4. [Adapter resilience - processes events after restart](#test-title-adapter-resilience---processes-events-after-restart)
5. [Multiple events for same cluster are idempotent](#test-title-multiple-events-for-same-cluster-are-idempotent)

---

## Test Title: Generation update triggers new adapter processing

### Description

This test validates that when a cluster's spec is updated, the generation field is incremented, and Sentinel detects this change and triggers a new adapter processing cycle. This ensures that spec changes are propagated through the system correctly.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier1 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet API, Sentinel, and Adapter are deployed and running
2. A cluster exists in Ready state

---

### Test Steps

#### Step 1: Create a cluster and wait for Ready state
**Action:**
```bash
# Create cluster
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "kind": "Cluster",
    "name": "generation-test",
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-central1"}}}
  }'

# Wait for Ready state
```

**Expected Result:**
- Cluster created with generation=1
- Cluster reaches Ready state

---

#### Step 2: Record initial generation and observed_generation
**Action:**
```bash
# Get cluster generation
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID} | jq '.generation'

# Get adapter observed_generation
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.items[0].observed_generation'
```

**Expected Result:**
- Cluster generation = 1
- Adapter observed_generation = 1

---

#### Step 3: Update cluster spec to trigger generation increment
**Action:**
```bash
curl -X PATCH ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID} \
  -H "Content-Type: application/json" \
  -d '{
    "spec": {"platform": {"type": "gcp", "gcp": {"projectID": "test", "region": "us-west1"}}}
  }'
```

**Expected Result:**
- API returns updated cluster with generation=2
- Spec change is persisted

---

#### Step 4: Wait for Sentinel to detect and publish event
**Action:**
- Wait for Sentinel polling interval (30-60 seconds)
- Check Sentinel logs for generation change detection

**Expected Result:**
- Sentinel detects generation changed from 1 to 2
- Sentinel publishes new event to broker

---

#### Step 5: Verify adapter processed new generation
**Action:**
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.items[0].observed_generation'
```

**Expected Result:**
- Adapter observed_generation = 2
- Adapter has processed the updated spec

---

## Test Title: Sentinel detects and publishes generation changes

### Description

This test validates that Sentinel correctly detects generation changes during its polling cycle and publishes events only for resources that need processing.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet API and Sentinel are deployed and running
2. Access to Sentinel logs

---

### Test Steps

#### Step 1: Create cluster and wait for initial processing
**Action:**
- Create cluster via API
- Wait for adapter to process and report status

**Expected Result:**
- Cluster created
- Adapter status reported with observed_generation=1

---

#### Step 2: Monitor Sentinel logs during polling
**Action:**
```bash
kubectl logs -n hyperfleet -l app.kubernetes.io/name=sentinel -f
```

**Expected Result:**
- Sentinel logs show polling activity
- No events published for clusters where generation == observed_generation

---

#### Step 3: Update cluster spec
**Action:**
- Update cluster spec via API
- Monitor Sentinel logs

**Expected Result:**
- Sentinel detects generation > observed_generation
- Sentinel publishes event for the cluster
- Log shows "generation changed" or similar message

---

## Test Title: Adapter tracks observed_generation correctly

### Description

This test validates that adapters correctly track and report observed_generation, which indicates the last generation they successfully processed.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier1 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet system is deployed and running
2. Adapter is processing events correctly

---

### Test Steps

#### Step 1: Create cluster and verify initial observed_generation
**Action:**
```bash
# Create cluster
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{"kind": "Cluster", "name": "obs-gen-test", "spec": {...}}'

# Wait and check status
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.items[0].observed_generation'
```

**Expected Result:**
- observed_generation = 1 (matches cluster generation)

---

#### Step 2: Update cluster multiple times
**Action:**
- Update cluster spec 3 times
- Check observed_generation after each update

**Expected Result:**
- After update 1: generation=2, observed_generation eventually=2
- After update 2: generation=3, observed_generation eventually=3
- After update 3: generation=4, observed_generation eventually=4

---

## Test Title: Adapter resilience - processes events after restart

### Description

This test validates that adapters can recover from restarts and continue processing pending events. Events that were not acknowledged before restart should be redelivered and processed.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet system is deployed and running
2. Adapter is running normally

---

### Test Steps

#### Step 1: Create cluster and verify initial processing
**Action:**
- Create cluster via API
- Verify adapter processes it and reports status

**Expected Result:**
- Cluster created and processed
- Adapter status shows Available=True or processing in progress

---

#### Step 2: Restart adapter pod
**Action:**
```bash
kubectl delete pod -n hyperfleet -l app.kubernetes.io/name=hyperfleet-adapter
```

**Expected Result:**
- Adapter pod is terminated
- New adapter pod starts up

---

#### Step 3: Create new cluster while adapter is restarting
**Action:**
- Create another cluster via API during adapter restart window

**Expected Result:**
- Cluster created successfully (API is independent of adapter)

---

#### Step 4: Verify adapter processes pending events after restart
**Action:**
- Wait for adapter to fully restart
- Check status of the new cluster

**Expected Result:**
- Adapter recovers and connects to broker
- Adapter processes pending events
- Both clusters eventually show adapter status

---

## Test Title: Multiple events for same cluster are idempotent

### Description

This test validates that when multiple events are published for the same cluster (e.g., due to retries or rapid updates), the adapter handles them idempotently without creating duplicate resources.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Tier2 |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-02-11 |
| **Updated** | 2026-02-11 |

---

### Preconditions
1. HyperFleet system is deployed and running
2. Adapter uses discovery pattern to check existing resources

---

### Test Steps

#### Step 1: Create cluster and wait for initial processing
**Action:**
- Create cluster via API
- Wait for adapter to create resources

**Expected Result:**
- Cluster created
- Adapter creates namespace and other resources
- Status reported

---

#### Step 2: Trigger multiple events for same cluster
**Action:**
- Update cluster spec multiple times in rapid succession
- Or restart Sentinel to trigger re-polling

**Expected Result:**
- Multiple events may be published for the same cluster

---

#### Step 3: Verify no duplicate resources created
**Action:**
```bash
# Check namespace count (should be exactly 1)
kubectl get namespace ${CLUSTER_ID}

# Check job count (should be 1 or managed correctly)
kubectl get jobs -n ${CLUSTER_ID}
```

**Expected Result:**
- Only one namespace exists for the cluster
- No duplicate jobs or resources
- Adapter discovers existing resources and evaluates postconditions

---

#### Step 4: Verify status is consistent
**Action:**
```bash
curl -s ${API_URL}/api/hyperfleet/v1/clusters/${CLUSTER_ID}/statuses | jq '.items | length'
```

**Expected Result:**
- Only one status entry per adapter
- Status reflects latest state (not duplicated)

---
