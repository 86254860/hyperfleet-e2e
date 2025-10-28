# HyperFleet API - Customer Critical Journey

## Overview

This document maps the critical user journeys for internal users interacting with the HyperFleet API. Each journey includes user actions, system responses, and the architectural components involved in processing the request.

> **Note:** All API request/response payloads shown in this document are still in design progress and not the final version. They will be updated once the API specification is published.

## Persona

**Internal Platform Engineer**
- Manages cluster/nodepool provisioning through HyperFleet API
- Monitors cluster/nodepool status

---

# Cluster Journeys

## Journey 1: Create a New Cluster

**Note on Cluster Phase Values:**
The HyperFleet API uses the following cluster phase values in `status.phase`:
- **`Pending`** - Cluster created, waiting for initial adapter processing
- **`Not Ready`** - One or more adapters are in Pending, Running, or Failed phase
- **`Ready`** - All required adapters have phase "Complete"
- **`Terminating`** - Cluster deletion initiated, cleanup in progress

Adapter phase values (in `statuses` table):
- **`Pending`** - Adapter hasn't started processing yet
- **`Running`** - Adapter job is actively executing
- **`Complete`** - Adapter successfully finished
- **`Failed`** - Adapter encountered an error

### Step 1: Submit Cluster Creation Request

**User Action:**
```bash
POST /api/hyperfleet/v1/clusters
Content-Type: application/json

{
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  }
}
```

**System Response / User Sees:**
```json
HTTP 201 Created
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "status": {
    "phase": "Pending",
    "lastTransitionTime": "2025-10-28T12:00:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T12:00:00Z"
}
```

---

### Step 2: Poll for Cluster Status (Initial Check)

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "status": {
    "phase": "Not Ready",
    "lastTransitionTime": "2025-10-28T12:00:15Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T12:00:15Z"
}
```

---

### Step 3: Monitor Validation Progress

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/statuses
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "statuses": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "validation",
      "phase": "Complete",
      "message": "Validation completed successfully. All checks passed: route53ZoneFound, s3BucketAccessible, quotaSufficient",
      "last_transition_time": "2025-10-28T12:02:00Z",
      "created_at": "2025-10-28T12:00:15Z"
    },
    {
      "id": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "dns",
      "phase": "Running",
      "message": "DNS configuration in progress",
      "last_transition_time": "2025-10-28T12:01:30Z",
      "created_at": "2025-10-28T12:01:00Z"
    },
    {
      "id": "c3d4e5f6-a7b8-9012-cdef-123456789012",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "infrastructure",
      "phase": "Pending",
      "message": "Waiting for validation and DNS to complete",
      "last_transition_time": "2025-10-28T12:00:15Z",
      "created_at": "2025-10-28T12:00:15Z"
    }
  ]
}
```

---

### Step 4: Continue Monitoring Until Ready

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

**System Response / User Sees (After All Adapters Complete):**
```json
HTTP 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "status": {
    "phase": "Ready",
    "lastTransitionTime": "2025-10-28T12:10:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T12:10:00Z"
}
```

**Success Criteria:**
- All required adapters have phase "Complete" (check via /statuses endpoint)
- Cluster status.phase is "Ready"
- status.lastTransitionTime reflects the latest adapter update

---

## Journey 2: Monitor Cluster Status and Troubleshoot Issues

### Step 1: Check High-Level Status

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "status": {
    "phase": "Not Ready",
    "lastTransitionTime": "2025-10-28T12:05:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T12:05:00Z"
}
```

**User Insight:**
- Cluster phase is "Not Ready"
- Last status update was at 12:05:00Z
- Need to check /statuses endpoint for detailed adapter information

---

### Step 2: Get Detailed Status for Failed Adapter

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/statuses
```

**System Response / User Sees (DNS Failure Example):**
```json
HTTP 200 OK
{
  "id": "status-cls-550e8400",
  "type": "clusterStatus",
  "href": "/api/hyperfleet/v1/clusters/cls-550e8400/statuses",
  "clusterId": "cls-550e8400",
  "adapterStatuses": [
    {
      "adapter": "validation",
      "observedGeneration": 1,
      "conditions": [
        {
          "type": "Available",
          "status": "True",
          "reason": "JobSucceeded",
          "message": "Job completed successfully after 115 seconds",
          "lastTransitionTime": "2025-10-17T12:02:00Z"
        },
        {
          "type": "Applied",
          "status": "True",
          "reason": "JobLaunched",
          "message": "Kubernetes Job created successfully",
          "lastTransitionTime": "2025-10-17T12:00:05Z"
        },
        {
          "type": "Health",
          "status": "True",
          "reason": "AllChecksPassed",
          "message": "All validation checks passed",
          "lastTransitionTime": "2025-10-17T12:02:00Z"
        }
      ],
      "data": {
        "validationResults": {
          "route53ZoneFound": true,
          "s3BucketAccessible": true,
          "quotaSufficient": true
        }
      },
      "metadata": {
        "jobName": "validation-cls-123-gen1",
        "executionTime": "115s"
      },
      "lastUpdated": "2025-10-17T12:02:00Z"
    },
    {
      "adapter": "dns",
      "observedGeneration": 1,
      "conditions": [
        {
          "type": "Available",
          "status": "True",
          "reason": "AllRecordsCreated",
          "message": "All DNS records created and verified",
          "lastTransitionTime": "2025-10-17T12:05:00Z"
        },
        {
          "type": "Applied",
          "status": "True",
          "reason": "JobLaunched",
          "message": "DNS Job created successfully",
          "lastTransitionTime": "2025-10-17T12:03:00Z"
        },
        {
          "type": "Health",
          "status": "True",
          "reason": "NoErrors",
          "message": "DNS adapter executed without errors",
          "lastTransitionTime": "2025-10-17T12:05:00Z"
        }
      ],
      "data": {
        "recordsCreated": ["api.my-cluster.example.com", "*.apps.my-cluster.example.com"]
      },
      "lastUpdated": "2025-10-17T12:05:00Z"
    }
  ],
  "lastUpdated": "2025-10-17T12:05:00Z"
}
```

**User Insight:**
- Validation and DNS adapter completed successfully

---

### Step 3: Wait for Automatic Recovery

**User Action:**
```bash
# Wait for Sentinel backoff period (10 seconds for Not Ready clusters)
# Then check status again
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

**System Response / User Sees (After Auto-Recovery):**
```json
HTTP 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "status": {
    "phase": "Ready",
    "lastTransitionTime": "2025-10-28T12:10:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T12:10:00Z"
}
```
**User Benefit:** No manual intervention required for transient failures. Sentinel Operator automatically retries based on backoff strategy.

---

## Journey 3: List and Filter Clusters

### Step 1: List All Clusters

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "my-test-cluster",
      "spec": {
        "provider": "gcp",
        "region": "us-east1",
        "nodeCount": 3
      },
      "status": {
        "phase": "Ready",
        "lastTransitionTime": "2025-10-28T12:10:00Z"
      },
      "labels": {
        "environment": "production",
        "team": "platform"
      },
      "created_at": "2025-10-28T12:00:00Z",
      "updated_at": "2025-10-28T12:10:00Z"
    },
    {
      "id": "660f9511-f3ac-52e5-b827-557766551111",
      "name": "my-staging-cluster",
      "spec": {
        "provider": "aws",
        "region": "us-west-2",
        "nodeCount": 2
      },
      "status": {
        "phase": "Not Ready",
        "lastTransitionTime": "2025-10-28T13:00:00Z"
      },
      "labels": {
        "environment": "staging",
        "team": "platform"
      },
      "created_at": "2025-10-28T13:00:00Z",
      "updated_at": "2025-10-28T13:00:00Z"
    }
  ],
  "total": 2
}
```

---

### Step 2: Filter Non-Ready Clusters

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters?phase=Not%20Ready
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "items": [
    {
      "id": "660f9511-f3ac-52e5-b827-557766551111",
      "name": "my-staging-cluster",
      "spec": {
        "provider": "aws",
        "region": "us-west-2",
        "nodeCount": 2
      },
      "status": {
        "phase": "Not Ready",
        "lastTransitionTime": "2025-10-28T13:00:00Z"
      },
      "labels": {
        "environment": "staging",
        "team": "platform"
      },
      "created_at": "2025-10-28T13:00:00Z",
      "updated_at": "2025-10-28T13:00:00Z"
    }
  ],
  "total": 1
}
```

**Use Case:** Monitor clusters that need attention

---

### Step 3: Filter Clusters with Customized Flags

**User Action:**
```bash
# Filter by custom labels (environment label)
GET /api/hyperfleet/v1/clusters?labels=environment:production

# Filter by region label
GET /api/hyperfleet/v1/clusters?labels=region:us-east1

# Filter by multiple labels
GET /api/hyperfleet/v1/clusters?labels=environment:production,team:platform

# Filter by provider
GET /api/hyperfleet/v1/clusters?provider=gcp

# Combine multiple filters (phase + labels)
GET /api/hyperfleet/v1/clusters?phase=Ready&labels=environment:production
```

**System Response / User Sees (Example: Filter by environment=production):**
```json
HTTP 200 OK
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "my-test-cluster",
      "spec": {
        "provider": "gcp",
        "region": "us-east1",
        "nodeCount": 3
      },
      "status": {
        "phase": "Ready",
        "lastTransitionTime": "2025-10-28T12:10:00Z"
      },
      "labels": {
        "environment": "production",
        "team": "platform"
      },
      "created_at": "2025-10-28T12:00:00Z",
      "updated_at": "2025-10-28T12:10:00Z"
    }
  ],
  "total": 1
}
```

**System Response / User Sees (Example: Combine filters - Ready + production):**
```json
HTTP 200 OK
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "my-test-cluster",
      "spec": {
        "provider": "gcp",
        "region": "us-east1",
        "nodeCount": 3
      },
      "status": {
        "phase": "Ready",
        "lastTransitionTime": "2025-10-28T12:10:00Z"
      },
      "labels": {
        "environment": "production",
        "team": "platform"
      },
      "created_at": "2025-10-28T12:00:00Z",
      "updated_at": "2025-10-28T12:10:00Z"
    }
  ],
  "total": 1
}
```

**Use Cases:**
- Filter clusters by environment (production, staging, development)
- Filter clusters by team or department
- Filter clusters by cloud provider (GCP, AWS)
- Filter clusters by region
- Combine multiple filters for precise cluster selection
- Support operational queries (e.g., "show all production clusters that are Ready")

**Supported Filter Parameters**:
- `labels`: Filter by custom label key-value pairs (format: `key:value` or `key1:value1,key2:value2`)
- `phase`: Filter by cluster phase (Pending, Not Ready, Ready, Terminating)
- `provider`: Filter by cloud provider (gcp, aws)
- `region`: Filter by cloud region (us-east1, us-west-2, etc.)
- `name`: Filter by cluster name (partial match or exact match)

**User Benefit:**
- Flexible filtering enables users to quickly locate specific clusters based on custom labels and metadata
- Supports complex operational scenarios like "find all production GCP clusters in us-east1 that are Ready"
- Enables team-based isolation and multi-tenancy use cases

---

## Journey 4: Update Cluster Configuration (Post-MVP)

### Step 1: Update Cluster Specification

**User Action:**
```bash
PATCH /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000  # Post-MVP
Content-Type: application/json

{
  "spec": {
    "provider": "gcp",
    "region": "us-west1",
    "nodeCount": 5
  }
}
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-west1",
    "nodeCount": 5
  },
  "status": {
    "phase": "Not Ready",
    "lastTransitionTime": "2025-10-28T13:00:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T13:00:00Z"
}
```

**What Happened:**
- Cluster spec updated (region changed to us-west1, nodeCount changed to 5)
- Phase automatically changed to "Not Ready" (adapters need to reconcile changes)
- Sentinel Operator will detect the spec change and publish reconciliation event

---

### Step 2: Monitor Adapter Reconciliation

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/statuses
```

**System Response / User Sees (During Reconciliation):**
```json
HTTP 200 OK
{
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "statuses": [
    {
      "id": "d4e5f6a7-b8c9-0123-defg-234567890123",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "validation",
      "phase": "Complete",
      "message": "Validation completed successfully for updated spec",
      "last_transition_time": "2025-10-28T13:01:00Z",
      "created_at": "2025-10-28T13:00:30Z"
    },
    {
      "id": "e5f6a7b8-c9d0-1234-efgh-345678901234",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "dns",
      "phase": "Running",
      "message": "DNS configuration in progress for new region us-west1",
      "last_transition_time": "2025-10-28T13:01:30Z",
      "created_at": "2025-10-28T13:01:00Z"
    },
    {
      "id": "f6a7b8c9-d0e1-2345-fghi-456789012345",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "infrastructure",
      "phase": "Pending",
      "message": "Waiting for validation and DNS to complete",
      "last_transition_time": "2025-10-28T13:00:30Z",
      "created_at": "2025-10-28T13:00:30Z"
    }
  ]
}
```

**User Insight:**
- Validation adapter has completed reconciliation for the updated spec
- DNS adapter is actively reconciling the region change
- Infrastructure adapter is pending (waiting for dependencies)

---

### Step 3: Verify Update Completion

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

**System Response / User Sees (After All Adapters Reconcile):**
```json
HTTP 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-west1",
    "nodeCount": 5
  },
  "status": {
    "phase": "Ready",
    "lastTransitionTime": "2025-10-28T13:05:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T13:05:00Z"
}
```

**Success Criteria:**
- All adapters have phase "Complete" (check via /statuses endpoint)
- Cluster status.phase is "Ready"
- Cluster updated with new spec successfully

---

## Journey 5: Delete Cluster (Deprovisioning) (Post-MVP)

### Step 1: Initiate Cluster Deletion

**User Action:**
```bash
DELETE /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000  # Post-MVP
```

**System Response / User Sees:**
```json
HTTP 202 Accepted
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-test-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-west1",
    "nodeCount": 5
  },
  "status": {
    "phase": "Terminating",
    "lastTransitionTime": "2025-10-28T14:00:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "created_at": "2025-10-28T12:00:00Z",
  "updated_at": "2025-10-28T14:00:00Z"
}
```

**What Happened:**
- Cluster status.phase changed to "Terminating"
- Sentinel Operator will detect the deletion and publish cleanup events
- Adapters will begin cleanup operations (typically in reverse order)

---

### Step 2: Monitor Deletion Progress

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/statuses
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "statuses": [
    {
      "id": "g7h8i9j0-k1l2-3456-mnop-567890123456",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "infrastructure",
      "phase": "Complete",
      "message": "Infrastructure resources successfully cleaned up",
      "last_transition_time": "2025-10-28T14:02:00Z",
      "created_at": "2025-10-28T14:00:30Z"
    },
    {
      "id": "h8i9j0k1-l2m3-4567-nopq-678901234567",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "adapter_name": "dns",
      "phase": "Running",
      "message": "Removing DNS records for us-west1",
      "last_transition_time": "2025-10-28T14:01:30Z",
      "created_at": "2025-10-28T14:01:00Z"
    }
  ]
}
```

---

### Step 3: Verify Deletion Completion

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000
```

**System Response / User Sees:**
```json
HTTP 404 Not Found
{
  "error": "ClusterNotFound",
  "message": "Cluster 550e8400-e29b-41d4-a716-446655440000 has been successfully deleted"
}
```

**Success Criteria:**
- Cluster no longer exists in the system (database record deleted)
- All adapter cleanup completed (all statuses show phase "Complete")
- Resources deprovisioned from cloud provider

---



---

## Journey 6: Cluster Access Control and Permissions(Out of MVP)

This journey demonstrates how the HyperFleet API enforces access control between different users.

### Scenario: User A Creates Cluster, User B Lacks Permission

**User A Action:**
```bash
POST /api/hyperfleet/v1/clusters
Authorization: Bearer <user-a-token>
Content-Type: application/json

{
  "name": "team-alpha-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "labels": {
    "environment": "production",
    "team": "alpha"
  }
}
```

**System Response:**
```json
HTTP 201 Created
{
  "id": "770e9622-g4bd-63f6-c938-668877662222",
  "name": "team-alpha-cluster",
  "spec": {
    "provider": "gcp",
    "region": "us-east1",
    "nodeCount": 3
  },
  "status": {
    "phase": "Pending",
    "lastTransitionTime": "2025-10-28T17:00:00Z"
  },
  "labels": {
    "environment": "production",
    "team": "alpha"
  },
  "created_at": "2025-10-28T17:00:00Z",
  "updated_at": "2025-10-28T17:00:00Z"
}
```

---

**User B Action (Attempting to Access User A's Cluster):**
```bash
GET /api/hyperfleet/v1/clusters/770e9622-g4bd-63f6-c938-668877662222
Authorization: Bearer <user-b-token>
```

**System Response:**
```json
HTTP 403 Forbidden
{
  "error": "PermissionDenied",
  "message": "User does not have permission to access this cluster"
}
```

**User Insight:**
- User A successfully created a cluster for team "alpha"
- User B (from a different team) cannot view or access User A's cluster
- The API enforces role-based access control (RBAC) based on labels and team membership
- Each user can only access clusters they have permissions for (e.g., their team's clusters)

---

# NodePool Journeys

## Journey 7: Create a New NodePool

**Note on NodePool Phase Values:**
NodePools use the same phase values as clusters:
- **`Pending`** - NodePool created, waiting for initial adapter processing
- **`Not Ready`** - One or more adapters are in Pending, Running, or Failed phase
- **`Ready`** - All required adapters have phase "Complete"
- **`Terminating`** - NodePool deletion initiated, cleanup in progress

### Step 1: Submit NodePool Creation Request

**User Action:**
```bash
POST /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/nodepools
Content-Type: application/json

{
  "name": "my-nodepool",
  "spec": {
    "nodeCount": 5,
    "machineType": "n1-standard-4"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  }
}
```

**System Response / User Sees:**
```json
HTTP 201 Created
{
  "id": "my-nodepool",
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "spec": {
    "nodeCount": 5,
    "machineType": "n1-standard-4"
  },
  "status": {
    "phase": "Pending",
    "lastTransitionTime": "2025-10-28T15:00:00Z"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  },
  "created_at": "2025-10-28T15:00:00Z",
  "updated_at": "2025-10-28T15:00:00Z"
}
```

---

### Step 2: Poll for NodePool Status

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/nodepools/my-nodepool
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "id": "my-nodepool",
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "spec": {
    "nodeCount": 5,
    "machineType": "n1-standard-4"
  },
  "status": {
    "phase": "Not Ready",
    "lastTransitionTime": "2025-10-28T15:00:15Z"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  },
  "created_at": "2025-10-28T15:00:00Z",
  "updated_at": "2025-10-28T15:00:15Z"
}
```

---


### Step 3: Verify NodePool Ready

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/nodepools/my-nodepool
```

**System Response / User Sees (After All Adapters Complete):**
```json
HTTP 200 OK
{
  "id": "my-nodepool",
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "spec": {
    "nodeCount": 5,
    "machineType": "n1-standard-4"
  },
  "status": {
    "phase": "Ready",
    "lastTransitionTime": "2025-10-28T15:05:00Z"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  },
  "created_at": "2025-10-28T15:00:00Z",
  "updated_at": "2025-10-28T15:05:00Z"
}
```

**Success Criteria:**
- All required adapters have phase "Complete"
- NodePool status.phase is "Ready"
- Nodes are provisioned and available

---

## Journey 8: List and Filter NodePools

### Step 1: List All NodePools

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/550e8400-e29b-41d4-a716-446655440000/nodepools
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "items": [
    {
      "id": "my-nodepool",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "spec": {
        "nodeCount": 5,
        "machineType": "n1-standard-4"
      },
      "status": {
        "phase": "Ready",
        "lastTransitionTime": "2025-10-28T15:05:00Z"
      },
      "labels": {
        "workload": "compute",
        "team": "platform"
      },
      "created_at": "2025-10-28T15:00:00Z",
      "updated_at": "2025-10-28T15:05:00Z"
    },
    {
      "id": "gpu-nodepool",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "spec": {
        "nodeCount": 3,
        "machineType": "n1-highmem-8"
      },
      "status": {
        "phase": "Not Ready",
        "lastTransitionTime": "2025-10-28T15:10:00Z"
      },
      "labels": {
        "workload": "gpu",
        "team": "ml"
      },
      "created_at": "2025-10-28T15:10:00Z",
      "updated_at": "2025-10-28T15:10:00Z"
    }
  ],
  "total": 2
}
```

---

### Step 2: Filter NodePools with Labels

**User Action:**
```bash
# Filter by labels
GET /api/hyperfleet/v1/clusters/<cluster_id>/nodepools?labels=workload:gpu

# Filter by phase
GET /api/hyperfleet/v1/clusters/<cluster_id>/nodepools?phase=Ready

# Combine multiple filters
GET /api/hyperfleet/v1/clusters/<cluster_id>/nodepools?labels=workload:gpu&phase=Ready
```

**System Response / User Sees (Example: Filter by workload=gpu):**
```json
HTTP 200 OK
{
  "items": [
    {
      "id": "gpu-nodepool",
      "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
      "spec": {
        "nodeCount": 3,
        "machineType": "n1-highmem-8"
      },
      "status": {
        "phase": "Not Ready",
        "lastTransitionTime": "2025-10-28T15:10:00Z"
      },
      "labels": {
        "workload": "gpu",
        "team": "ml"
      },
      "created_at": "2025-10-28T15:10:00Z",
      "updated_at": "2025-10-28T15:10:00Z"
    }
  ],
  "total": 1
}
```

**Use Cases:**
- List all nodepools for a specific cluster
- Filter nodepools by workload type
- Monitor nodepools by phase (Ready, Not Ready, etc.)

---

## Journey 9: Update NodePool Configuration (Post-MVP)

### Step 1: Update NodePool Specification

**User Action:**
```bash
PATCH /api/hyperfleet/v1/clusters/<cluster_id>/nodepools/<nodepool_id>  # Post-MVP
Content-Type: application/json

{
  "spec": {
    "nodeCount": 8
  }
}
```

**System Response / User Sees:**
```json
HTTP 200 OK
{
  "id": "my-nodepool",
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "spec": {
    "nodeCount": 8,
    "machineType": "n1-highmem-8"
  },
  "status": {
    "phase": "Not Ready",
    "lastTransitionTime": "2025-10-28T16:00:00Z"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  },
  "created_at": "2025-10-28T15:00:00Z",
  "updated_at": "2025-10-28T16:00:00Z"
}
```

**What Happened:**
- NodePool spec updated (nodeCount changed to 8)
---

### Step 2: Verify Update Completion

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/<cluster_id>/nodepools/<nodepool_id>
```

**System Response / User Sees (After Reconciliation):**
```json
HTTP 200 OK
{
  "id": "my-nodepool",
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "spec": {
    "nodeCount": 8,
    "machineType": "n1-standard-8"
  },
  "status": {
    "phase": "Ready",
    "lastTransitionTime": "2025-10-28T16:05:00Z"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  },
  "created_at": "2025-10-28T15:00:00Z",
  "updated_at": "2025-10-28T16:05:00Z"
}
```

**Success Criteria:**
- All adapters have phase "Complete"
- NodePool status.phase is "Ready"
- NodePool updated with new spec

---

## Journey 10: Delete NodePool (Post-MVP)

### Step 1: Initiate NodePool Deletion

**User Action:**
```bash
DELETE /api/hyperfleet/v1/clusters/<cluster_id>/nodepools/<nodepool_id>  # Post-MVP
```

**System Response / User Sees:**
```json
HTTP 202 Accepted
{
  "id": "my-nodepool",
  "cluster_id": "550e8400-e29b-41d4-a716-446655440000",
  "spec": {
    "nodeCount": 8,
    "machineType": "n1-standard-8"
  },
  "status": {
    "phase": "Terminating",
    "lastTransitionTime": "2025-10-28T17:00:00Z"
  },
  "labels": {
    "workload": "compute",
    "team": "platform"
  },
  "created_at": "2025-10-28T15:00:00Z",
  "updated_at": "2025-10-28T17:00:00Z"
}
```

**What Happened:**
- NodePool status.phase changed to "Terminating"
- Sentinel Operator will detect deletion and trigger cleanup
- Adapters will begin cleanup operations

---

### Step 2: Verify Deletion Completion

**User Action:**
```bash
GET /api/hyperfleet/v1/clusters/<cluster_id>/nodepools/<nodepool_id>
```

**System Response / User Sees:**
```json
HTTP 404 Not Found
{
  "error": "NodePoolNotFound",
  "message": "NodePool 880f0733-h5ce-74g7-d049-779988773333 has been successfully deleted"
}
```

**Success Criteria:**
- NodePool no longer exists in the system
- All adapter cleanup completed
- Nodes removed from cluster

---

## API Endpoints Summary

### Cluster Endpoints
- `POST /clusters` - Create a new cluster
- `GET /clusters` - List all clusters (with filtering support)
- `GET /clusters/{id}` - Get cluster details and high-level status
- `PATCH /clusters/{id}` - Update cluster specification (Post-MVP)
- `DELETE /clusters/{id}` - Delete/deprovision a cluster (Post-MVP)
- `GET /clusters/{id}/statuses` - Get detailed adapter status information

### NodePool Endpoints
- `POST /clusters/{cluster_id}/nodepools` - Create a new nodepool
- `GET /clusters/{cluster_id}/nodepools` - List all nodepools (with filtering support)
- `GET /clusters/{cluster_id}/nodepools/{id}` - Get nodepool details and high-level status
- `PATCH /clusters/{cluster_id}/nodepools/{id}` - Update nodepool specification (Post-MVP)
- `DELETE /clusters/{cluster_id}/nodepools/{id}` - Delete nodepool (Post-MVP)
- `GET /clusters/{cluster_id}/nodepools/{id}/statuses` - Get detailed adapter status information

