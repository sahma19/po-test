package tests_test

import (
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sahma19/po-test/pkg/tests"
)

func TestPoTest(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Po-test Suite")
}

var _ = ginkgo.Describe("Po-test", func() {
	ginkgo.Context("Success", func() {
		ginkgo.It("Should mutate files and run unit tests", func() {
			testFilename := "prometheus-operator-unittest.yml"
			ruleFilename := "prometheus-operator-rules.yml"

			gomega.Expect(tests.RunUnitTests([]string{testFilename})).To(gomega.Succeed())

			file, err := os.ReadFile(ruleFilename)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(string(file)).To(gomega.ContainSubstring("PrometheusRule"))
		})

		ginkgo.It("Should run tests in relative paths", func() {
			testFilename := "subdir/prometheus-operator-unittest.yml"
			ruleFilename := "subdir/prometheus-operator-rules-subdir.yml"

			gomega.Expect(tests.RunUnitTests([]string{testFilename})).To(gomega.Succeed())

			file, err := os.ReadFile(ruleFilename)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(string(file)).To(gomega.ContainSubstring("PrometheusRule"))
		})
	})

	ginkgo.Context("Failure", func() {
		ginkgo.It("Should report error when tests fail", func() {
			testFilename := "bad-rules-error-test.yml"
			ruleFilename := "bad-rules-error.yml"

			err := tests.RunUnitTests([]string{testFilename})
			gomega.Expect(err).To(gomega.HaveOccurred())

			file, err := os.ReadFile(ruleFilename)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(string(file)).To(gomega.ContainSubstring("PrometheusRule"))
		})
	})
})
