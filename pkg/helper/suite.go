package helper

import (
	"log"
	"sync"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/client"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/config"
)

var (
	// suiteConfig is loaded once in cmd layer before tests start
	suiteConfig *config.Config
	configMutex sync.RWMutex
)

// SetSuiteConfig sets the global suite configuration for the test suite
func SetSuiteConfig(cfg *config.Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	suiteConfig = cfg
}

// GetSuiteConfig returns the global suite configuration
func GetSuiteConfig() *config.Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return suiteConfig
}

// ClearSuiteConfig clears the global suite configuration
func ClearSuiteConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()
	suiteConfig = nil
}

// New creates a helper instance for testing
// Creates a new helper per test
func New() *Helper {
	cfg := GetSuiteConfig()
	if cfg == nil {
		log.Fatalf("Suite config not initialized")
	}

	h, err := newHelper(cfg)
	if err != nil {
		log.Fatalf("Failed to create helper: %v", err)
	}
	return h
}

// newHelper creates a new Helper instance (internal use)
func newHelper(cfg *config.Config) (*Helper, error) {
	cl, err := client.NewHyperFleetClient(cfg.API.URL, nil)
	if err != nil {
		return nil, err
	}

	k8sClient, err := initK8sClient()
	if err != nil {
		return nil, err
	}

	return &Helper{
		Cfg:       cfg,
		Client:    cl,
		K8sClient: k8sClient,
	}, nil
}
