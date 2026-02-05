# Test Case: E2E-004 - Full NodePool Creation Flow

## Test Information

- **Test ID**: E2E-004
- **Test Title**: Full NodePool Creation Flow
- **Priority**: Critical
- **Test Type**: End-to-End (Happy Path)
- **Component**: NodePool Lifecycle
- **Reference**: [hyperfleet-api-e2e-scenario.md#e2e-004](../../user-journeys/hyperfleet-api-e2e-scenario.md#e2e-004)

## Test Objective

Validate end-to-end nodepool creation from API request to Ready state for an existing cluster, ensuring all adapters complete successfully and nodes are provisioned.

## Prerequisites

- HyperFleet API is accessible
- Valid API credentials
- GCP project with necessary permissions
- Parent cluster already created and in Ready state

## Test Steps

### Step 1: Create Prerequisite Cluster

**Action**: Create a GCP cluster via HyperFleet API

**API Request**:
```bash
POST /api/hyperfleet/v1/clusters
Content-Type: application/json
```

**Request Body**:
```json
{
  "kind": "Cluster",
  "name": "hp-gcp-cluster-1",
  "labels": {
    "environment": "test",
    "team": "platform"
  },
  "spec": {
    "platform": {
      "type": "gcp",
      "gcp": {
        "projectID": "my-gcp-project",
        "region": "us-central1",
        "zone": "us-central1-a",
        "network": "default",
        "subnet": "default-subnet"
      }
    },
    "release": {
      "image": "registry.redhat.io/openshift4/ose-cluster-version-operator:v4.14.0",
      "version": "4.14.0"
    },
    "networking": {
      "clusterNetwork": [
        {
          "cidr": "10.10.0.0/16",
          "hostPrefix": 24
        }
      ],
      "serviceNetwork": ["10.96.0.0/12"]
    },
    "dns": {
      "baseDomain": "example.com"
    }
  }
}
```

**Expected Response**: HTTP 201 Created

**Wait for**: Cluster status.phase = "Ready"

---

### Step 2: Submit NodePool Creation Request

**Action**: Create a nodepool for the cluster

**API Request**:
```bash
POST /api/hyperfleet/v1/clusters/{cluster_id}/nodepools
Content-Type: application/json
```

**Request Body**:
```json
{
  "kind": "NodePool",
  "name": "gpu-nodepool",
  "labels": {
    "workload": "gpu",
    "tier": "compute",
    "environment": "test"
  },
  "spec": {
    "replicas": 2,
    "machineType": "n1-standard-8",
    "labels": {
      "node-role": "worker",
      "gpu-enabled": "true"
    }
  }
}
```

**Expected Response**: HTTP 201 Created

**Response Body** (example):
```json
{
  "id": "np-550e8400-e29b-41d4-a716-446655440001",
  "href": "/api/hyperfleet/v1/clusters/cluster-123/nodepools/np-550e8400-e29b-41d4-a716-446655440001",
  "kind": "NodePool",
  "name": "gpu-nodepool",
  "labels": {
    "workload": "gpu",
    "tier": "compute",
    "environment": "test"
  },
  "created_by": "system-admin",
  "updated_by": "system-admin",
  "created_time": "2024-01-15T10:40:00Z",
  "updated_time": "2024-01-15T10:40:00Z",
  "generation": 1,
  "spec": {
    "replicas": 2,
    "machineType": "n1-standard-8",
    "labels": {
      "node-role": "worker",
      "gpu-enabled": "true"
    }
  },
  "status": {
    "phase": "NotReady",
    "message": "NodePool creation in progress"
  }
}
```

**Validations**:
- Response status code is 201 Created
- NodePool ID is generated (not empty)
- `status.phase` is "NotReady" initially
- `kind` is "NodePool"
- `name` matches request ("gpu-nodepool")
- `spec.replicas` is 2
- `spec.machineType` is "n1-standard-8"
- `generation` is 1

---

### Step 3: Verify NodePool in List

**Action**: List all nodepools for the cluster

**API Request**:
```bash
GET /api/hyperfleet/v1/clusters/{cluster_id}/nodepools
```

**Expected Response**: HTTP 200 OK

**Response Body** (example):
```json
{
  "kind": "NodePoolList",
  "page": 1,
  "size": 1,
  "total": 1,
  "items": [
    {
      "id": "np-550e8400-e29b-41d4-a716-446655440001",
      "name": "gpu-nodepool",
      "labels": {
        "workload": "gpu",
        "tier": "compute",
        "environment": "test"
      },
      "spec": {
        "replicas": 2,
        "machineType": "n1-standard-8"
      },
      "status": {
        "phase": "NotReady"
      }
    }
  ]
}
```

**Validations**:
- Created nodepool appears in list
- `total` >= 1
- NodePool ID matches the one created in Step 2

---

### Step 4: Monitor NodePool Status

**Action**: Poll nodepool status until Ready

**API Request**:
```bash
GET /api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}
```

**Expected Response**: HTTP 200 OK

**Validations**:
- `status.phase` remains "NotReady" during provisioning
- Eventually transitions to "Ready"
- No errors in `status.message`

---

### Step 5: Monitor Adapter Statuses

**Action**: Check detailed adapter status for nodepool

**API Request**:
```bash
GET /api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}/statuses
```

**Expected Response**: HTTP 200 OK

**Response Body** (example):
```json
{
  "kind": "NodePoolStatusList",
  "items": [
    {
      "adapter": "validation-adapter",
      "conditions": [
        {
          "type": "Applied",
          "status": "True",
          "reason": "JobLaunched",
          "message": "Validation Job created successfully",
          "last_transition_time": "2024-01-15T10:40:15Z"
        },
        {
          "type": "Available",
          "status": "True",
          "reason": "ValidationComplete",
          "message": "NodePool validation completed successfully",
          "last_transition_time": "2024-01-15T10:40:45Z"
        },
        {
          "type": "Health",
          "status": "True",
          "reason": "NoErrors",
          "message": "Adapter executed without errors",
          "last_transition_time": "2024-01-15T10:40:45Z"
        }
      ]
    },
    {
      "adapter": "nodepool-adapter",
      "conditions": [
        {
          "type": "Applied",
          "status": "True",
          "reason": "JobLaunched",
          "message": "NodePool provisioning Job created successfully",
          "last_transition_time": "2024-01-15T10:41:00Z"
        },
        {
          "type": "Available",
          "status": "True",
          "reason": "NodesProvisioned",
          "message": "All 2 nodes provisioned and joined cluster",
          "last_transition_time": "2024-01-15T10:45:30Z"
        },
        {
          "type": "Health",
          "status": "True",
          "reason": "NoErrors",
          "message": "Adapter executed without errors",
          "last_transition_time": "2024-01-15T10:45:30Z"
        }
      ]
    }
  ]
}
```

**Validations**:
- NodePoolStatusList contains adapter statuses
- Each adapter has three condition types: `Applied`, `Available`, `Health`
- All conditions transition to:
  - **Applied**: `True` (Job/resources created)
  - **Available**: `True` (work completed successfully)
  - **Health**: `True` (no unexpected errors)

**Expected Adapter Flow**:
1. **Validation Adapter**:
   - Applied: True → Job created
   - Available: False (validating) → True (validation complete)
   - Health: True throughout

2. **NodePool Adapter**:
   - Applied: True → Job created
   - Available: False (provisioning) → True (nodes ready)
   - Health: True throughout

---

### Step 6: Verify Final State

**Action**: Verify nodepool reached Ready state and all adapters successful

**API Request**:
```bash
GET /api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}
```

**Expected Response**: HTTP 200 OK

**Final Response Body** (example):
```json
{
  "id": "np-550e8400-e29b-41d4-a716-446655440001",
  "href": "/api/hyperfleet/v1/clusters/cluster-123/nodepools/np-550e8400-e29b-41d4-a716-446655440001",
  "kind": "NodePool",
  "name": "gpu-nodepool",
  "labels": {
    "workload": "gpu",
    "tier": "compute",
    "environment": "test"
  },
  "created_by": "system-admin",
  "updated_by": "system-admin",
  "created_time": "2024-01-15T10:40:00Z",
  "updated_time": "2024-01-15T10:46:00Z",
  "generation": 1,
  "spec": {
    "replicas": 2,
    "machineType": "n1-standard-8",
    "labels": {
      "node-role": "worker",
      "gpu-enabled": "true"
    }
  },
  "status": {
    "phase": "Ready",
    "message": "NodePool is ready, all nodes provisioned and healthy"
  }
}
```

**Validations**:
- `status.phase` is "Ready"
- `status.message` indicates successful completion
- All adapter conditions remain `True`
- Nodes are running and joined to the cluster (can verify via cluster API)

---

## Success Criteria

1. **NodePool Creation**:
   - HTTP 201 response received
   - NodePool ID generated
   - Initial phase is "NotReady"
   - All spec fields match request

2. **Phase Transition**:
   - NodePool transitions from "NotReady" to "Ready"
   - Transition completes within expected timeout (e.g., 30 minutes)

3. **Adapter Execution**:
   - All adapters complete successfully
   - All three condition types report `True`:
     - Applied: True (resources created)
     - Available: True (work completed)
     - Health: True (no errors)
   - No errors in adapter logs

4. **Node Provisioning**:
   - Specified number of nodes (2) are created
   - Nodes are in Ready state
   - Nodes have joined the parent cluster
   - Nodes have correct labels and machine type

5. **API Consistency**:
   - NodePool appears in list endpoint
   - Status endpoint returns detailed adapter information
   - No data corruption or inconsistency

## Expected Duration

- **Average**: 15-30 minutes (depends on cloud provider provisioning time)
- **Maximum**: 45 minutes

## Cleanup

After test completion:
1. Delete the nodepool: `DELETE /api/hyperfleet/v1/clusters/{cluster_id}/nodepools/{nodepool_id}`
2. Delete the parent cluster: `DELETE /api/hyperfleet/v1/clusters/{cluster_id}`
3. Verify all resources are cleaned up (no orphaned nodes or cloud resources)

## Notes

- This test requires a parent cluster to be created first (prerequisite)
- NodePool creation depends on cluster being in Ready state
- Node provisioning time varies by cloud provider and region
- Test validates both NodePool lifecycle and node provisioning
- Different from cluster creation (E2E-001) which focuses on cluster-level resources

## Related Test Cases

- **E2E-001**: Full Cluster Creation Flow - prerequisite for this test
- **E2E-005** (Post-MVP): NodePool Configuration Update
- **E2E-006** (Post-MVP): NodePool Deletion
- **E2E-FAIL-002**: NodePool API Request Body Validation Failures
- **E2E-FAIL-004**: Adapter Failures (applies to nodepool adapters too)

## Test Implementation

- **File**: `e2e/nodepool/creation.go`
- **Labels**: `lifecycle`, `critical`, `happy-path`
- **Payload**: `testdata/payloads/nodepools/gcp.json`
