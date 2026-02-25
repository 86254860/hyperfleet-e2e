package helper

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
)

var (
    // k8sNameRegex validates Kubernetes resource names (DNS-1123 subdomain format)
    k8sNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
    // labelSelectorRegex validates Kubernetes label selector format (key=value,key=value)
    // Supports optional prefix (e.g., hyperfleet.io/cluster-id=value)
    labelSelectorRegex = regexp.MustCompile(`^([a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?/)?[a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?=[a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?(,([a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?/)?[a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?=[a-zA-Z0-9]([-a-zA-Z0-9_.]*[a-zA-Z0-9])?)*$`)
)

// ValidateK8sName validates a Kubernetes resource name (namespace, resource name, etc.)
func ValidateK8sName(name string) error {
    if name == "" {
        return fmt.Errorf("name cannot be empty")
    }
    if len(name) > 253 {
        return fmt.Errorf("name %q exceeds maximum length of 253 characters", name)
    }
    if !k8sNameRegex.MatchString(name) {
        return fmt.Errorf("name %q is invalid: must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character", name)
    }
    return nil
}

// ValidateLabelSelector validates a Kubernetes label selector string
func ValidateLabelSelector(selector string) error {
    if selector == "" {
        return fmt.Errorf("label selector cannot be empty")
    }
    if !labelSelectorRegex.MatchString(selector) {
        return fmt.Errorf("label selector %q is invalid: must be in format 'key=value' or 'key1=value1,key2=value2'", selector)
    }
    return nil
}

// HasAdapterCondition checks if an adapter condition with the given type and status exists in the conditions list
func (h *Helper) HasAdapterCondition(conditions []openapi.AdapterCondition, condType string, status openapi.AdapterConditionStatus) bool {
    for _, cond := range conditions {
        if cond.Type == condType && cond.Status == status {
            return true
        }
    }
    return false
}

// HasResourceCondition checks if a resource condition with the given type and status exists in the conditions list
func (h *Helper) HasResourceCondition(conditions []openapi.ResourceCondition, condType string, status openapi.ResourceConditionStatus) bool {
    for _, cond := range conditions {
        if cond.Type == condType && cond.Status == status {
            return true
        }
    }
    return false
}

// GetCondition retrieves a condition by type from the conditions list
func (h *Helper) GetCondition(conditions []openapi.AdapterCondition, condType string) *openapi.AdapterCondition {
    for i := range conditions {
        if conditions[i].Type == condType {
            return &conditions[i]
        }
    }
    return nil
}

// AllConditionsTrue checks if all specified condition types have status True
func (h *Helper) AllConditionsTrue(conditions []openapi.AdapterCondition, condTypes []string) bool {
    for _, condType := range condTypes {
        if !h.HasAdapterCondition(conditions, condType, openapi.AdapterConditionStatusTrue) {
            return false
        }
    }
    return true
}

// AnyConditionFalse checks if any of the specified condition types have status False
func (h *Helper) AnyConditionFalse(conditions []openapi.AdapterCondition, condTypes []string) bool {
    for _, condType := range condTypes {
        if h.HasAdapterCondition(conditions, condType, openapi.AdapterConditionStatusFalse) {
            return true
        }
    }
    return false
}

// AdapterNameToClusterConditionType converts an adapter name to its corresponding cluster condition type.
// Examples:
//   - "cl-namespace" -> "ClNamespaceSuccessful"
//   - "cl-job" -> "ClJobSuccessful"
//   - "cl-deployment" -> "ClDeploymentSuccessful"
func (h *Helper) AdapterNameToClusterConditionType(adapterName string) string {
    // Split adapter name by "-" (e.g., "cl-namespace" -> ["cl", "namespace"])
    parts := strings.Split(adapterName, "-")

    // Capitalize each part and join them
    var result strings.Builder
    for _, part := range parts {
        if len(part) > 0 {
            result.WriteString(strings.ToUpper(part[:1]) + part[1:])
        }
    }

    // Add "Successful" suffix
    result.WriteString("Successful")

    return result.String()
}
