
# Feature: Cluster Lifecycle Management

## Test Title: Create cluster will succeed via API

### Description

This test case validates the core cluster API operations including creating a new GCP cluster via the Hyperfleet API and verifying it appears in the cluster list with the correct configuration.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Critical |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-01-08 |
| **Updated** | 2026-01-08 |


---

### Preconditions

1. Hyperfleet API server is running and accessible
2. Set the API gateway URL as an environment variable: `export API_URL=<your-api-gateway-url>`

---

### Test Steps

#### Step 1: Create Cluster via API

**Action:**
Send POST request to create a new GCP cluster using the payload from [templates/create_cluster_gcp.json](templates/create_cluster_gcp.json):

```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testcases/templates/create_cluster_gcp.json
```

- <details>
  <summary>Payload example (click to expand)</summary>

See [templates/create_cluster_gcp.json](templates/create_cluster_gcp.json) for the complete cluster creation payload.

Key fields in the payload:
- `kind`: "Cluster"
- `name`: "hp-gcp-cluster-1"
- `labels`: environment and team labels
- `spec.platform.type`: "gcp"
- `spec.platform.gcp`: GCP-specific configuration (projectID, region, zone, network, subnet)
- `spec.release`: OpenShift version and image
- `spec.networking`: cluster and service network CIDRs
- `spec.dns.baseDomain`: base domain for the cluster

</details>

**Expected Result:**
- Response status code is 201 (Created)
- <details>
  <summary>Response example (click to expand)</summary>

  ```json
  {
    "created_by": "system",
    "created_time": "2026-01-08T06:30:10.812472273Z",
    "generation": 1,
    "href": "/api/hyperfleet/v1/clusters/2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
    "id": "2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
    "kind": "Cluster",
    "labels": {
      "environment": "production",
      "team": "platform"
    },
    "name": "hp-gcp-cluster-1",
    "spec": {
      "dns": {
        "baseDomain": "example.com"
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
      "platform": {
        "gcp": {
          "network": "default",
          "projectID": "my-gcp-project",
          "region": "us-central1",
          "subnet": "default-subnet",
          "zone": "us-central1-a"
        },
        "type": "gcp"
      },
      "release": {
        "image": "registry.redhat.io/openshift4/ose-cluster-version-operator:v4.14.0",
        "version": "4.14.0"
      }
    },
    "status": {
      "conditions": [],
      "last_transition_time": "0001-01-01T00:00:00Z",
      "last_updated_time": "0001-01-01T00:00:00Z",
      "observed_generation": 0,
      "phase": "NotReady"
    },
    "updated_by": "system",
    "updated_time": "2026-01-08T06:30:10.812472273Z"
  }
  ```
  </details>
- Verify response fields:
  - `id` is automatically generated, non-empty, lowercase, and unique
  - `href` matches pattern `/api/hyperfleet/v1/clusters/{cluster_id}`
  - `kind` is "Cluster"
  - `name` is "hp-gcp-cluster-1"
  - `labels` contains "environment": "production" and "team": "platform"
  - `created_by` and `updated_by` are populated (currently "system" as placeholder; will change when auth is introduced)
  - `created_time` and `updated_time` are populated and not default values (not "0001-01-01T00:00:00Z")
  - `generation` is 1
  - `status.phase` is "NotReady" (initial state in MVP phase)
  - `status.conditions` is empty array
  - `status.observed_generation` is 0
- All spec fields match the request payload

---

#### Step 2: Verify Cluster API response

**Action:**
Send GET request to retrieve the cluster list:

```bash
curl -G ${API_URL}/api/hyperfleet/v1/clusters --data-urlencode "search=name='<cluster_name>'"
```

**Expected Result:**
- Response status code is 200 (OK)
- <details>
  <summary>Response example (click to expand)</summary>

  ```json
  {
  "items": [
    {
      "created_by": "system",
      "created_time": "2026-01-08T06:30:10.812472Z",
      "generation": 1,
      "href": "/api/hyperfleet/v1/clusters/2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
      "id": "2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
      "kind": "Cluster",
      "labels": {
        "environment": "production",
        "team": "platform"
      },
      "name": "hp-gcp-cluster-1",
      "spec": {
        "dns": {
          "baseDomain": "example.com"
        },
        "networking": {
          "clusterNetwork": [
            {
              "cidr": "10.10.0.0/16",
              "hostPrefix": 24
            }
          ],
          "serviceNetwork": [
            "10.96.0.0/12"
          ]
        },
        "platform": {
          "gcp": {
            "network": "default",
            "projectID": "my-gcp-project",
            "region": "us-central1",
            "subnet": "default-subnet",
            "zone": "us-central1-a"
          },
          "type": "gcp"
        },
        "release": {
          "image": "registry.redhat.io/openshift4/ose-cluster-version-operator:v4.14.0",
          "version": "4.14.0"
        }
      },
      "status": {
        "conditions": [
          {
            "created_time": "2026-01-08T06:30:15Z",
            "last_transition_time": "2026-01-08T06:30:15Z",
            "last_updated_time": "2026-01-08T06:58:00.086394Z",
            "message": "Landing zone namespace is active and ready",
            "observed_generation": 1,
            "reason": "NamespaceReady",
            "status": "True",
            "type": "LandingZoneAdapterSuccessful"
          },
          {
            "created_time": "2026-01-08T06:30:15Z",
            "last_transition_time": "2026-01-08T06:30:40Z",
            "last_updated_time": "2026-01-08T06:58:00.164973Z",
            "message": "GCP environment validated successfully (simulated)",
            "observed_generation": 1,
            "reason": "ValidationPassed",
            "status": "True",
            "type": "DummyValidationAdapterSuccessful"
          }
        ],
        "last_transition_time": "2026-01-08T06:30:15.073896Z",
        "last_updated_time": "2026-01-08T06:58:00.086394Z",
        "observed_generation": 1,
        "phase": "NotReady"
      },
      "updated_by": "system",
      "updated_time": "2026-01-08T06:58:00.169988Z"
    }
  ],
  "kind": "ClusterList",
  "page": 1,
  "size": 1,
  "total": 1
  }
  ```
  </details>
- **ClusterList metadata:**
  - `kind` is "ClusterList"
  - `total` is 1
  - `size` is 1
  - `page` is 1
- **Created cluster appears in the `items` array**
- **System default fields** :
  - `id` matches the ID from Step 1
  - `href` is "/api/hyperfleet/v1/clusters/{cluster_id}"
  - `kind` is "Cluster"
  - `created_by` is populated (currently "system" as placeholder; will change when auth is introduced)
  - `created_time` is populated and not default value (not "0001-01-01T00:00:00Z")
  - `updated_by` is populated (currently "system" as placeholder; will change when auth is introduced)
  - `updated_time` is populated and not default value (not "0001-01-01T00:00:00Z")
  - `generation` is 1
  - `status.phase` is "NotReady" (initial state in MVP phase)
  - `status.conditions` array exists with required fields: `type`, `status`, `reason`, `message`, `created_time`, `last_transition_time`, `last_updated_time`, `observed_generation`
  - `status.observed_generation` matches cluster generation
  - `status.last_transition_time` and `status.last_updated_time` are populated with the real values
- **Cluster request body configured parameters:**
  - All spec fields match the request payload

#### Step 3: Retrieve the specific cluster and monitor its status 

**Action:**
1.  Send GET request to retrieve the specific cluster:

```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id} 
```

**Expected Result:**
- Response status code is 200 (OK)
- <details>
  <summary>Response example (click to expand)</summary>

  ```json
  {
    "created_by": "system",
    "created_time": "2026-01-08T06:30:10.812472Z",
    "generation": 1,
    "href": "/api/hyperfleet/v1/clusters/2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
    "id": "2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
    "kind": "Cluster",
    "labels": {
      "environment": "production",
      "team": "platform"
    },
    "name": "hp-gcp-cluster-1",
    "spec": {
      "dns": {
        "baseDomain": "example.com"
      },
      "networking": {
        "clusterNetwork": [
          {
            "cidr": "10.10.0.0/16",
            "hostPrefix": 24
          }
        ],
        "serviceNetwork": [
          "10.96.0.0/12"
        ]
      },
      "platform": {
        "gcp": {
          "network": "default",
          "projectID": "my-gcp-project",
          "region": "us-central1",
          "subnet": "default-subnet",
          "zone": "us-central1-a"
        },
        "type": "gcp"
      },
      "release": {
        "image": "registry.redhat.io/openshift4/ose-cluster-version-operator:v4.14.0",
        "version": "4.14.0"
      }
    },
    "status": {
      "conditions": [
        {
          "created_time": "2026-01-08T06:30:15Z",
          "last_transition_time": "2026-01-08T06:30:15Z",
          "last_updated_time": "2026-01-08T06:30:55.131925Z",
          "message": "Landing zone namespace is active and ready",
          "observed_generation": 1,
          "reason": "NamespaceReady",
          "status": "True",
          "type": "LandingZoneAdapterSuccessful"
        },
        {
          "created_time": "2026-01-08T06:30:15Z",
          "last_transition_time": "2026-01-08T06:30:40Z",
          "last_updated_time": "2026-01-08T06:30:40.146459Z",
          "message": "GCP environment validated successfully (simulated)",
          "observed_generation": 1,
          "reason": "ValidationPassed",
          "status": "True",
          "type": "DummyValidationAdapterSuccessful"
        }
      ],
      "last_transition_time": "2026-01-08T06:30:15.073896Z",
      "last_updated_time": "2026-01-08T06:30:40.146459Z",
      "observed_generation": 1,
      "phase": "NotReady"
    },
    "updated_by": "system",
    "updated_time": "2026-01-08T06:30:55.140946Z"
  }
  ```
  </details>
- Response contains all cluster metadata fields from Step 1
- Cluster status contains adapter information:
  - `status.conditions` array contains adapter status entries
  - **LandingZoneAdapterSuccessful** condition exists:(example)
    - `type`: "LandingZoneAdapterSuccessful"
    - `status`: "True"
    - `reason`: "NamespaceReady"
    - `message`: "Landing zone namespace is active and ready"
    - `created_time`, `last_transition_time`, `last_updated_time` populated
    - `observed_generation`: 1
- `updated_time` is more recent than `created_time`, indicating the cluster has been processed by adapters

---

#### Step 4: Retrieve the Cluster Adapter Statuses

**Action:**
Send GET request to retrieve the adapter statuses for the cluster:

```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses
```

**Expected Result:**
- Response status code is 200 (OK)
- <details>
  <summary>Response example (click to expand)</summary>

  ```json
  {
    "items": [
      {
        "adapter": "landing-zone-adapter",
        "conditions": [
          {
            "last_transition_time": "2026-01-08T06:30:15Z",
            "message": "Namespace created successfully",
            "reason": "NamespaceCreated",
            "status": "True",
            "type": "Applied"
          },
          {
            "last_transition_time": "2026-01-08T06:30:15Z",
            "message": "Landing zone namespace is active and ready",
            "reason": "NamespaceReady",
            "status": "True",
            "type": "Available"
          },
          {
            "last_transition_time": "2026-01-08T06:30:15Z",
            "message": "All adapter operations completed successfully",
            "reason": "Healthy",
            "status": "True",
            "type": "Health"
          }
        ],
        "created_time": "2026-01-08T06:30:15Z",
        "data": {
          "namespace": {
            "name": "2nmg1slsk2t795jdu1mi1rk2iplgf4k3",
            "status": "Active"
          }
        },
        "last_report_time": "2026-01-08T07:04:00.080806Z",
        "observed_generation": 1
      },
      {
        "adapter": "dummy-validation-adapter",
        "conditions": [
          {
            "last_transition_time": "2026-01-08T06:30:25Z",
            "message": "Validation job applied successfully",
            "reason": "JobApplied",
            "status": "True",
            "type": "Applied"
          },
          {
            "last_transition_time": "2026-01-08T06:30:40Z",
            "message": "GCP environment validated successfully (simulated)",
            "reason": "ValidationPassed",
            "status": "True",
            "type": "Available"
          },
          {
            "last_transition_time": "2026-01-08T06:30:15Z",
            "message": "All adapter operations completed successfully",
            "reason": "Healthy",
            "status": "True",
            "type": "Health"
          }
        ],
        "created_time": "2026-01-08T06:30:15Z",
        "data": {},
        "last_report_time": "2026-01-08T07:04:00.162251Z",
        "observed_generation": 1
      }
    ],
    "kind": "AdapterStatusList",
    "page": 1,
    "size": 2,
    "total": 2
  }
  ```
  </details>
- Response contains AdapterStatusList metadata:
  - `kind` is "AdapterStatusList"
  - `total` matches the number of deployed adapters (at least 2 in MVP: landing-zone-adapter and dummy-validation-adapter)
  - `size` matches `total`
  - `page` is 1
- Response contains `items` array with adapter status entries:
  - **landing-zone-adapter** status exists:
    - `adapter`: Adapter name
    - `created_time` is populated
    - `last_report_time` is populated and recent
    - `observed_generation`: 1
    - `conditions` array contains three required condition types: **Applied**, **Available**, and **Health**
      - Each condition's `status` field is false at the beginning and will be "True" when that specific condition is satisfied
      - **Applied** condition:
        - `type`: "Applied"
        - `status`: "True" (when resources are successfully applied)
        - `reason`: "NamespaceCreated"
        - `message`: "Namespace created successfully"
        - `last_transition_time` is populated
      - **Available** condition:
        - `type`: "Available"
        - `status`: "True" (when resources are available and ready)
        - `reason`: "NamespaceReady"
        - `message`: "Landing zone namespace is active and ready"
        - `last_transition_time` is populated
      - **Health** condition:
        - `type`: "Health"
        - `status`: "True" (when all adapter operations are healthy)
        - `reason`: "Healthy"
        - `message`: "All adapter operations completed successfully"
        - `last_transition_time` is populated
    - `data` contains namespace information :
      - `namespace.name`: matches cluster ID
      - `namespace.status`: "Active"
      - additional fields like KSA will be added in future

---

#### Step 5: Verify Cluster Final State

**Action:**
Send GET request to retrieve the cluster status and verify final state:

```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id} | jq .status
```

**Expected Result:**
- Response status code is 200 (OK)
- <details>
  <summary>Status response example (click to expand)</summary>

  ```json
  {
    "conditions": [
      {
        "created_time": "2026-01-08T06:30:15Z",
        "last_transition_time": "2026-01-08T06:30:15Z",
        "last_updated_time": "2026-01-08T07:10:25.134567Z",
        "message": "Landing zone namespace is active and ready",
        "observed_generation": 1,
        "reason": "NamespaceReady",
        "status": "True",
        "type": "LandingZoneAdapterSuccessful"
      },
      {
        "created_time": "2026-01-08T06:30:15Z",
        "last_transition_time": "2026-01-08T06:30:40Z",
        "last_updated_time": "2026-01-08T06:30:40.146459Z",
        "message": "GCP environment validated successfully (simulated)",
        "observed_generation": 1,
        "reason": "ValidationPassed",
        "status": "True",
        "type": "DummyValidationAdapterSuccessful"
      }
    ],
    "last_transition_time": "2026-01-08T06:30:40.146459Z",
    "last_updated_time": "2026-01-08T07:10:25.134567Z",
    "observed_generation": 1,
    "phase": "Ready"
  }
  ```
  </details>
- Verify cluster final state:
  - **Cluster phase:**
    - `phase` 'Ready'
      - Real available adapters number == Expected adapter number → Cluster phase: Ready
      - Any adapter Available: False → Cluster phase: Not Ready (MVP)
    - `observed_generation` is 1
    - `last_transition_time` is populated (when phase last changed)
    - `last_updated_time` is populated and more recent than creation time
  - **Adapter conditions:**
    - All adapter conditions have `status`: "True"
    - `conditions` array contains adapters information. For instance: "LandingZoneAdapterSuccessful" and    "ValidationAdapterSuccessful"
    - Each condition has valid `created_time`, `last_transition_time`, and `last_updated_time`

---

## Test Title: Resources should be created after cluster creation 

### Description

This test case validates that the cluster adapters have created the expected resources in the deployment environment.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Positive |
| **Priority** | Critical |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-01-08 |
| **Updated** | 2026-01-08 |


---

### Preconditions

1. Hyperfleet API server is running and accessible
2. Set the API gateway URL as an environment variable: `export API_URL=<your-api-gateway-url>`
3. A cluster has been created and processed by adapters (see [Create cluster will succeed via API](#test-title-create-cluster-will-succeed-via-api))
4. kubectl is configured to access the deployment Kubernetes cluster. If you have access to deployed cluster, you use this cmd
   ```bash
   gcloud container clusters get-credentials <cluster_name> --zone=<zone> --project=<gcp_project_id>
   ```
---

### Test Steps

#### Step 1: Check resources in deployed environment

**Action:**
Verify Kubernetes resources created by the adapters:

1. List all namespaces to verify cluster namespace exists:
```bash
kubectl get namespace
```

2. Check pods in the cluster namespace:
```bash
kubectl get pods -n {cluster_id}
```

**Expected Result:**

- **Adapter created resources:**
  - **Landing zone adapter** creates a new namespace named with cluster ID
  - **GCP validation adapter** creates a validation job under the namespace

- **Namespace verification:**
  - Namespace with name matching the cluster ID exists (created by landing-zone-adapter)
  - Namespace status is "Active"
  - Example output:
  ```text
  NAME                                  STATUS   AGE
  2nmg1slsk2t795jdu1mi1rk2iplgf4k3      Active   5m
  default                               Active   1d
  kube-system                           Active   1d
  ```

- **Pods verification:**
  - Validation pod exists in the cluster namespace (created by dummy-validation-adapter)
  - Pod status should be "Completed" without errors
  - Pod should not have any restarts or error states
  - Example output:
  ```text
  NAME                                                     READY   STATUS      RESTARTS   AGE
  gcp-validator-2nmd7vphnua8m388ve9mp8oevah401up-1-99tqr   0/2     Completed   0          3h58
  ```
---

