package e2e

import (
	"log"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/spf13/viper"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/config"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/helper"
	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/logger"
)

var (
	// suiteConfig is loaded once in cmd layer before tests start
	suiteConfig *config.Config
)

func SetSuiteConfig(cfg *config.Config) {
	suiteConfig = cfg
	helper.SetSuiteConfig(cfg)
}

func GetSuiteConfig() *config.Config {
	return suiteConfig
}

var _ = ginkgo.BeforeSuite(func() {
	cfg := GetSuiteConfig()
	if cfg == nil {
		log.Fatalf("Suite config not initialized")
	}

	if err := logger.Init(&cfg.Log, "dev"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	cfg.Display()

	logger.Info("starting hyperfleet-e2e test suite - each test creates temporary resources")

	if junitPath := viper.GetString(config.Tests.JUnitReportPath); junitPath != "" {
		setupJUnitReporter(junitPath)
	}
})

var _ = ginkgo.AfterSuite(func() {
	helper.ClearSuiteConfig()
	logger.Info("test suite completed")
})

// setupJUnitReporter configures JUnit XML report generation
func setupJUnitReporter(junitPath string) {
	ginkgo.ReportAfterSuite("HyperFleet E2E JUnit Report", func(report ginkgo.Report) {
		dir := filepath.Dir(junitPath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			logger.Error("failed to create report directory", "path", dir, "error", err)
			return
		}

		err := reporters.GenerateJUnitReportWithConfig(
			report,
			junitPath,
			reporters.JunitReportConfig{
				OmitSpecLabels:   true,
				OmitLeafNodeType: true,
			},
		)
		if err != nil {
			logger.Error("failed to create JUnit report", "path", junitPath, "error", err)
		} else {
			logger.Info("junit report written", "path", junitPath)
		}
	})
}
