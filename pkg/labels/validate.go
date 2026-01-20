package labels

import "fmt"

// ValidateLabels verifies that tests contain all required label dimensions
func ValidateLabels(testLabels []string) error {
	hasPriority := false
	hasStability := false
	hasScenario := false

	for _, label := range testLabels {
		switch label {
		// Priority dimension (required)
		case Tier0, Tier1, Tier2:
			hasPriority = true
		// Stability dimension (required)
		case Stable, Informing, Flaky:
			hasStability = true
		// Scenario dimension (required)
		case HappyPath, Negative, Scale:
			hasScenario = true
		// Functionality dimension (optional)
		case Lifecycle, Upgrade:
			// Optional, no validation needed
		// Constraint dimension (optional)
		case Serial, Disruptive, Slow:
			// Optional, no validation needed
		}
	}

	if !hasPriority {
		return fmt.Errorf("missing priority label (tier0/tier1/tier2)")
	}
	if !hasStability {
		return fmt.Errorf("missing stability label (stable/informing/flaky)")
	}
	if !hasScenario {
		return fmt.Errorf("missing scenario label (happy-path/negative/scale)")
	}

	return nil
}
