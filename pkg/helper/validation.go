package helper

import (
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/api/openapi"
)

// HasCondition checks if a condition with the given type and status exists in the conditions list
func (h *Helper) HasCondition(conditions []openapi.AdapterCondition, condType string, status openapi.ConditionStatus) bool {
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
		if !h.HasCondition(conditions, condType, openapi.True) {
			return false
		}
	}
	return true
}

// AnyConditionFalse checks if any of the specified condition types have status False
func (h *Helper) AnyConditionFalse(conditions []openapi.AdapterCondition, condTypes []string) bool {
	for _, condType := range condTypes {
		if h.HasCondition(conditions, condType, openapi.False) {
			return true
		}
	}
	return false
}
