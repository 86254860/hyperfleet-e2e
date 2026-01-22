package labels

// Priority labels - Business value dimension: determines failure handling priority
const (
	Tier0 = "tier0" // Critical path: fix immediately, blocks merge
	Tier1 = "tier1" // Important: non-critical but frequently used, fix within 24h
	Tier2 = "tier2" // Edge case: low-frequency scenarios, run in scheduled jobs only
)

// Stability labels - Test quality dimension: determines CI gate policy
const (
	Stable    = "stable"    // Production-ready: stable and reliable, must pass to merge (Blocking)
	Informing = "informing" // Observation period: new test onboarding (Non-blocking)
	Flaky     = "flaky"     // Known unstable: quarantined for investigation
)

// Scenario labels - Test path dimension: describes test design intent
const (
	HappyPath = "happy-path" // Normal workflow: ideal path
	Negative  = "negative"   // Error handling: edge cases and failure scenarios
	Scale     = "scale"      // Performance: stress tests or large-scale resource scenarios
)

// Functionality labels - Feature category dimension: describes test coverage target
const (
	Lifecycle = "lifecycle" // Full lifecycle: Create -> Ready -> Delete
	Upgrade   = "upgrade"   // Version compatibility: smooth upgrades
)

// Constraint labels - Execution constraint dimension: determines scheduling strategy
const (
	Serial     = "serial"     // Must run serially: cannot run in parallel
	Disruptive = "disruptive" // Destructive testing: fault injection
	Slow       = "slow"       // Long-running: execution time exceeds 5-10 minutes
)
