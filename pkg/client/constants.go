package client

// Condition types used by adapters
const (
	ConditionTypeApplied   = "Applied"   // Resources created successfully
	ConditionTypeAvailable = "Available" // Work completed successfully
	ConditionTypeHealth    = "Health"    // No unexpected errors
)

// Condition types used by cluster-level resources (clusters, nodepools)
const (
	ConditionTypeReady = "Ready" // Resource is ready for use
)
