
# Feature: Cluster Adapter Status Management

## Test Title: The adapter's status should work well when its work meets error

### Description

This test case validates that the Hyperfleet API correctly reports adapter failure statuses when an adapter's work progress encounters errors during cluster provisioning. It ensures that failure conditions are properly reflected in both the cluster status and the adapter status endpoints, with appropriate error messages and status transitions.

---

| **Field** | **Value** |
|-----------|-----------|
| **Pos/Neg** | Negative |
| **Priority** | High |
| **Status** | Draft |
| **Automation** | Not Automated |
| **Version** | MVP |
| **Created** | 2026-01-27 |
| **Updated** | 2026-01-27 |


---

### Preconditions

1. The HyperFleet platform has been deployed successfully
2. Kubectl is configured to access the deployment GKE cluster
3. Know how to use [hyperfleet-chart](https://github.com/openshift-hyperfleet/hyperfleet-chart/) to deploy the HyperFleet platform
4. Use the job adapter(in this case, the example is gcp_validation_adapter ) test mode to do the test temporarily until the real mode is ready

---

### Test Steps

#### Step 1: Configure job adapter to simulate failure scenario

**Action:**
1. Modify the template file [gcp_validation_adapter_values.yaml](../../../testdata/payloads/clusters/gcp_validation_adapter_values.yaml) and change `dummy.simulateResult` from "success" to "failure"

2. Generate the values file with environment variable substitution:

```bash
envsubst < testdata/payloads/clusters/gcp_validation_adapter_values.yaml > values_changed.yaml
```

3. Upgrade the deployed HyperFleet platform using Helm:

```bash
helm upgrade $RELEASE_NAME . \
    -f values.yaml \
    -f values_changed.yaml \
    -n $NAMESPACE_NAME
```

  Parameters:
  - `RELEASE_NAME`: The deployed HyperFleet release name
  - `NAMESPACE_NAME`: The deployed GKE namespace name
  - `values.yaml`: Default values file from [hyperfleet-chart/examples/gcp-pubsub/values.yaml](https://github.com/openshift-hyperfleet/hyperfleet-chart/blob/main/examples/gcp-pubsub/values.yaml)

**Expected Result:**
- HyperFleet platform upgrades successfully
- Related adapter pod is restarted
- <details>
  <summary>Response example (click to expand)</summary>
  Release "hyperfleet" has been upgraded. Happy Helming!
  NAME: hyperfleet
  LAST DEPLOYED: Tue Jan 20 10:00:16 2026
  NAMESPACE: yingzhan-prow-default
  STATUS: deployed
  REVISION: 6
  DESCRIPTION: Upgrade complete
  TEST SUITE: None
  NOTES:
  HyperFleet for Google Cloud Platform has been deployed!
  Components:
    - HyperFleet API: Cluster lifecycle management
    - Sentinel (clusters): Resource polling and event publishing
    - Landing Zone Adapter: Namespace creation
    - Validation GCP Adapter (clusters): GCP cluster validation
  Broker Configuration:
    Type: Google Pub/Sub
    Project: YOUR_GCP_PROJECT
  Workload Identity Federation:
    IAM permissions are granted directly to Kubernetes service accounts via terraform.
    No GCP service accounts or annotations needed.
    If you see permission errors, verify:
    1. Terraform was applied with the correct namespace (yingzhan-prow-default)
    2. Service account names match terraform configuration:
      - sentinel
      - landing-zone-adapter
      - validation-gcp-adapter
  To check the status:
    kubectl get pods -n yingzhan-prow-default
  To view logs:
    kubectl logs -n yingzhan-prow-default -l app.kubernetes.io/name=hyperfleet-api
    kubectl logs -n yingzhan-prow-default -l app.kubernetes.io/name=sentinel
    kubectl logs -n yingzhan-prow-default -l app.kubernetes.io/name=adapter-landing-zone
    kubectl logs -n yingzhan-prow-default -l app.kubernetes.io/name=validation-gcp
  </details>

---

#### Step 2: Expose HyperFleet API Service

**Action:**
Get the external IP

```bash
kubectl get svc hyperfleet-api -n $NAMESPACE_NAME
```

**Expected Result:**
- External IP is assigned to the service
- Example output:
  ```text
  NAME             TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                                        AGE
  hyperfleet-api   LoadBalancer   10.102.106.30   34.31.96.225   8000:32666/TCP,8080:30918/TCP,9090:31231/TCP   3d18h
  ```
- Set the external IP as an environment variable:
  ```bash
  export API_URL="http://34.31.96.225:8000"
  ```

---

#### Step 3: Create Cluster via API

**Action:**
Send POST request to create a new GCP cluster using the payload from [gcp.json](../../../testdata/payloads/clusters/gcp.json):

```bash
curl -X POST ${API_URL}/api/hyperfleet/v1/clusters \
  -H "Content-Type: application/json" \
  -d @testdata/payloads/clusters/gcp.json
```

**Expected Result:**
- Response status code is 201 (Created)
- Initial `status.phase` is "NotReady"(MVP)

---

#### Step 4: Verify Adapter Available status Shows False

**Action:**
Wait 2-3 minutes for adapters to process the cluster, then send GET request to retrieve the cluster adapter statuses:

```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id}/statuses | jq .
```

**Expected Result:**
- Response status code is 200 (OK)
- **Landing Zone Adapter** shows successful status:
  - All conditions (`Applied`, `Available`, `Health`) have `status`: "True"
  - No errors or failures
- **Test Adapter** shows failure status:
  - `Applied` condition: `status` is "True" (job was created successfully)
  - `Available` condition: `status` is "False" with:
    - `reason`: "MissingPermissions"
    - `message`: Contains detailed information about the simulated failure (e.g., "Service account lacks required IAM permissions (simulated)")
  - `Health` condition: `status` is "True" (adapter health is good, only validation failed)
- <details>
  <summary>Response example (click to expand)</summary>

  ```json
  {
  "items": [
    {
      "adapter": "landing-zone-adapter",
      "conditions": [
        {
          "last_transition_time": "2026-01-20T03:37:13Z",
          "message": "Namespace created successfully",
          "reason": "NamespaceCreated",
          "status": "True",
          "type": "Applied"
        },
        {
          "last_transition_time": "2026-01-20T03:37:13Z",
          "message": "Landing zone namespace is active and ready",
          "reason": "NamespaceReady",
          "status": "True",
          "type": "Available"
        },
        {
          "last_transition_time": "2026-01-20T03:37:13Z",
          "message": "All adapter operations completed successfully",
          "reason": "Healthy",
          "status": "True",
          "type": "Health"
        }
      ],
      "created_time": "2026-01-20T03:37:13Z",
      "data": {
        "namespace": {
          "name": "2nuakq24e3f22ft6060fda7mm3n5khsa",
          "status": "Active"
        }
      },
      "last_report_time": "2026-01-20T03:37:38.461177Z",
      "observed_generation": 1
    },
    {
      "adapter": "dummy-validation-adapter",
      "conditions": [
        {
          "last_transition_time": "2026-01-20T03:37:23Z",
          "message": "Validation job applied successfully",
          "reason": "JobApplied",
          "status": "True",
          "type": "Applied"
        },
        {
          "last_transition_time": "2026-01-20T03:37:13Z",
          "message": "Service account lacks required IAM permissions (simulated)",
          "reason": "MissingPermissions",
          "status": "False",
          "type": "Available"
        },
        {
          "last_transition_time": "2026-01-20T03:37:13Z",
          "message": "All adapter operations completed successfully",
          "reason": "Healthy",
          "status": "True",
          "type": "Health"
        }
      ],
      "created_time": "2026-01-20T03:37:13Z",
      "data": {},
      "last_report_time": "2026-01-20T03:37:38.553638Z",
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

---

#### Step 5: Verify Cluster Remains in NotReady State

**Action:**
Send GET request to retrieve the cluster status and verify it reflects the adapter failure:

```bash
curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/{cluster_id} | jq .status
```

**Expected Result:**
- Response status code is 200 (OK)
- `status.phase` is "NotReady" (cluster does not transition to "Ready" when any adapter fails)
- `status.conditions` array contains both adapter conditions:
  - **LandingZoneAdapterSuccessful**: `status` is "True"
  - **DummyValidationAdapterSuccessful**: `status` is "False" with failure details
- `status.last_updated_time` is populated and recent
- <details>
  <summary>Response example (click to expand)</summary>

  ```bash
  curl -X GET ${API_URL}/api/hyperfleet/v1/clusters/2nuakq24e3f22ft6060fda7mm3n5khsa | jq .status.phase
  "NotReady"
  ```

  Full status object example:
  ```json
  "status": {
    "conditions": [
      {
        "created_time": "2026-01-20T03:37:13Z",
        "last_transition_time": "2026-01-20T03:37:13Z",
        "last_updated_time": "2026-01-20T03:45:53.396489Z",
        "message": "Landing zone namespace is active and ready",
        "observed_generation": 1,
        "reason": "NamespaceReady",
        "status": "True",
        "type": "LandingZoneAdapterSuccessful"
      },
      {
        "created_time": "2026-01-20T03:37:13Z",
        "last_transition_time": "2026-01-20T03:37:13Z",
        "last_updated_time": "2026-01-20T03:45:53.512442Z",
        "message": "Service account lacks required IAM permissions (simulated)",
        "observed_generation": 1,
        "reason": "MissingPermissions",
        "status": "False",
        "type": "DummyValidationAdapterSuccessful"
      }
    ],
    "last_transition_time": "2026-01-20T03:37:13.437331Z",
    "last_updated_time": "2026-01-20T03:45:53.396489Z",
    "observed_generation": 1,
    "phase": "NotReady"
  }
  ```
  </details>

---

