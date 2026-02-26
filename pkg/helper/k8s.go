package helper

import (
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
)

// buildLabelSelector converts a label map to kubectl label selector string
// e.g., {"app": "test", "tier": "frontend"} -> "app=test,tier=frontend"
func buildLabelSelector(labels map[string]string) string {
    if len(labels) == 0 {
        return ""
    }

    var parts []string
    for k, v := range labels {
        parts = append(parts, fmt.Sprintf("%s=%s", k, v))
    }
    return strings.Join(parts, ",")
}

// VerifyNamespaceActive verifies a namespace exists and is in Active phase
func (h *Helper) VerifyNamespaceActive(ctx context.Context, name string, expectedLabels, expectedAnnotations map[string]string) error {
    logger.Info("verifying namespace status", "namespace", name)

    // Create context with timeout
    cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Get namespace as JSON
    cmd := exec.CommandContext(cmdCtx, "kubectl", "get", "namespace", name, "-o", "json")
    output, err := cmd.Output()
    if err != nil {
        // Capture stderr for error context
        stderr := ""
        if exitErr, ok := err.(*exec.ExitError); ok {
            stderr = string(exitErr.Stderr)
        }
        return fmt.Errorf("failed to get namespace %s: %w (stderr: %s)", name, err, stderr)
    }

    // Parse JSON
    var nsData map[string]interface{}
    if err := json.Unmarshal(output, &nsData); err != nil {
        return fmt.Errorf("failed to parse namespace JSON: %w", err)
    }

    // Verify namespace phase is Active
    status, ok := nsData["status"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("namespace %s has no status field", name)
    }

    phase, ok := status["phase"].(string)
    if !ok || phase != "Active" {
        return fmt.Errorf("namespace %s phase is %v, expected Active", name, phase)
    }

    // Verify labels
    metadata, ok := nsData["metadata"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("namespace %s has no metadata field", name)
    }

    labels, _ := metadata["labels"].(map[string]interface{})
    if err := h.verifyMapContains(labels, expectedLabels, "label"); err != nil {
        return fmt.Errorf("namespace %s: %w", name, err)
    }

    // Verify annotations
    annotations, _ := metadata["annotations"].(map[string]interface{})
    if err := h.verifyMapContains(annotations, expectedAnnotations, "annotation"); err != nil {
        return fmt.Errorf("namespace %s: %w", name, err)
    }

    logger.Info("namespace verified successfully", "namespace", name, "phase", phase)
    return nil
}

// VerifyJobComplete verifies a job exists and has completed successfully.
// Uses expectedLabels to find the job via label selector - if kubectl returns a job,
// it's guaranteed to have those labels (no need to verify them again).
func (h *Helper) VerifyJobComplete(ctx context.Context, namespace string, expectedLabels, expectedAnnotations map[string]string) error {
    // Build label selector from expected labels to find the job
    labelSelector := buildLabelSelector(expectedLabels)
    logger.Info("verifying job status", "namespace", namespace, "label_selector", labelSelector)

    // Validate inputs to prevent command injection
    if err := ValidateK8sName(namespace); err != nil {
        return fmt.Errorf("invalid namespace: %w", err)
    }
    if err := ValidateLabelSelector(labelSelector); err != nil {
        return fmt.Errorf("invalid label selector: %w", err)
    }

    // Create context with timeout
    cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Get job by label selector
    // TODO: Consider using Kubernetes client-go library instead of kubectl CLI
    // #nosec G204 -- namespace and labelSelector are validated above to prevent command injection
    cmd := exec.CommandContext(cmdCtx, "kubectl", "get", "job",
        "-n", namespace,
        "-l", labelSelector,
        "-o", "json")
    output, err := cmd.Output()
    if err != nil {
        // Capture stderr for error context
        stderr := ""
        if exitErr, ok := err.(*exec.ExitError); ok {
            stderr = string(exitErr.Stderr)
        }
        return fmt.Errorf("failed to get job in namespace %s with selector %s: %w (stderr: %s)",
            namespace, labelSelector, err, stderr)
    }

    // Parse JSON list
    var jobList map[string]interface{}
    if err := json.Unmarshal(output, &jobList); err != nil {
        return fmt.Errorf("failed to parse job list JSON: %w", err)
    }

    items, ok := jobList["items"].([]interface{})
    if !ok || len(items) == 0 {
        return fmt.Errorf("no job found in namespace %s with selector %s", namespace, labelSelector)
    }
    if len(items) > 1 {
        return fmt.Errorf("multiple jobs (%d) found in namespace %s with selector %s - expected exactly one", len(items), namespace, labelSelector)
    }

    // Get the first job (should be only one)
    jobData, ok := items[0].(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid job data format")
    }

    // Verify job status
    status, ok := jobData["status"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("job has no status field")
    }

    // Check if job completed successfully
    succeeded, _ := status["succeeded"].(float64)
    conditions, _ := status["conditions"].([]interface{})

    jobComplete := false
    if succeeded > 0 {
        jobComplete = true
    } else {
        // Check for Complete condition
        for _, cond := range conditions {
            condMap, ok := cond.(map[string]interface{})
            if !ok {
                continue
            }
            if condMap["type"] == "Complete" && condMap["status"] == "True" {
                jobComplete = true
                break
            }
        }
    }

    if !jobComplete {
        return fmt.Errorf("job in namespace %s has not completed successfully (succeeded=%v)", namespace, succeeded)
    }

    // Verify annotations
    metadata, ok := jobData["metadata"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("job has no metadata field")
    }

    annotations, _ := metadata["annotations"].(map[string]interface{})
    if err := h.verifyMapContains(annotations, expectedAnnotations, "annotation"); err != nil {
        return fmt.Errorf("job in namespace %s: %w", namespace, err)
    }

    jobName := metadata["name"]
    logger.Info("job verified successfully", "namespace", namespace, "job", jobName, "succeeded", succeeded)
    return nil
}

// VerifyDeploymentAvailable verifies a deployment exists and is available.
// Uses expectedLabels to find the deployment via label selector - if kubectl returns a deployment,
// it's guaranteed to have those labels (no need to verify them again).
func (h *Helper) VerifyDeploymentAvailable(ctx context.Context, namespace string, expectedLabels, expectedAnnotations map[string]string) error {
    // Build label selector from expected labels to find the deployment
    labelSelector := buildLabelSelector(expectedLabels)
    logger.Info("verifying deployment status", "namespace", namespace, "label_selector", labelSelector)

    // Validate inputs to prevent command injection
    if err := ValidateK8sName(namespace); err != nil {
        return fmt.Errorf("invalid namespace: %w", err)
    }
    if err := ValidateLabelSelector(labelSelector); err != nil {
        return fmt.Errorf("invalid label selector: %w", err)
    }

    // Create context with timeout
    cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    // Get deployment by label selector
    // TODO: Consider using Kubernetes client-go library instead of kubectl CLI
    // #nosec G204 -- namespace and labelSelector are validated above to prevent command injection
    cmd := exec.CommandContext(cmdCtx, "kubectl", "get", "deployment",
        "-n", namespace,
        "-l", labelSelector,
        "-o", "json")
    output, err := cmd.Output()
    if err != nil {
        // Capture stderr for error context
        stderr := ""
        if exitErr, ok := err.(*exec.ExitError); ok {
            stderr = string(exitErr.Stderr)
        }
        return fmt.Errorf("failed to get deployment in namespace %s with selector %s: %w (stderr: %s)",
            namespace, labelSelector, err, stderr)
    }

    // Parse JSON list
    var deployList map[string]interface{}
    if err := json.Unmarshal(output, &deployList); err != nil {
        return fmt.Errorf("failed to parse deployment list JSON: %w", err)
    }

    items, ok := deployList["items"].([]interface{})
    if !ok || len(items) == 0 {
        return fmt.Errorf("no deployment found in namespace %s with selector %s", namespace, labelSelector)
    }
    if len(items) > 1 {
        return fmt.Errorf("multiple deployments (%d) found in namespace %s with selector %s - expected exactly one", len(items), namespace, labelSelector)
    }

    // Get the first deployment (should be only one)
    deployData, ok := items[0].(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid deployment data format")
    }

    // Verify deployment status
    status, ok := deployData["status"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("deployment has no status field")
    }

    // Check available replicas
    availableReplicas, _ := status["availableReplicas"].(float64)
    conditions, _ := status["conditions"].([]interface{})

    deployAvailable := false
    if availableReplicas > 0 {
        // Also check for Available condition
        for _, cond := range conditions {
            condMap, ok := cond.(map[string]interface{})
            if !ok {
                continue
            }
            if condMap["type"] == "Available" && condMap["status"] == "True" {
                deployAvailable = true
                break
            }
        }
    }

    if !deployAvailable {
        return fmt.Errorf("deployment in namespace %s is not available (availableReplicas=%v)", namespace, availableReplicas)
    }

    // Verify annotations
    metadata, ok := deployData["metadata"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("deployment has no metadata field")
    }

    annotations, _ := metadata["annotations"].(map[string]interface{})
    if err := h.verifyMapContains(annotations, expectedAnnotations, "annotation"); err != nil {
        return fmt.Errorf("deployment in namespace %s: %w", namespace, err)
    }

    deployName := metadata["name"]
    logger.Info("deployment verified successfully", "namespace", namespace, "deployment", deployName, "availableReplicas", availableReplicas)
    return nil
}

// verifyMapContains checks if actual map contains all expected key-value pairs
func (h *Helper) verifyMapContains(actual map[string]interface{}, expected map[string]string, mapType string) error {
    var missing []string
    var mismatched []string

    for key, expectedValue := range expected {
        actualValue, exists := actual[key]
        if !exists {
            missing = append(missing, key)
            continue
        }

        // Convert actual value to string for comparison
        actualStr := fmt.Sprintf("%v", actualValue)
        if actualStr != expectedValue {
            mismatched = append(mismatched, fmt.Sprintf("%s (expected: %s, actual: %s)", key, expectedValue, actualStr))
        }
    }

    if len(missing) > 0 || len(mismatched) > 0 {
        var errParts []string
        if len(missing) > 0 {
            errParts = append(errParts, fmt.Sprintf("missing %ss: %s", mapType, strings.Join(missing, ", ")))
        }
        if len(mismatched) > 0 {
            errParts = append(errParts, fmt.Sprintf("mismatched %ss: %s", mapType, strings.Join(mismatched, "; ")))
        }
        return fmt.Errorf("%s", strings.Join(errParts, "; "))
    }

    return nil
}
