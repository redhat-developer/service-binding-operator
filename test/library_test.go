package test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAPIUsageAsLibrary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Library Suite")
}

var _ = DescribeTable("Using service binding API as library", func(projectPath string) {

	cmd := exec.Command("sh", "-c", "go mod tidy && go build -mod=mod -o main")
	absProjectPath, err := filepath.Abs(projectPath)
	Expect(err).NotTo(HaveOccurred())

	cmd.Dir = absProjectPath
	cmd.Stdout = GinkgoWriter
	cmd.Stderr = GinkgoWriter
	err = cmd.Run()

	Expect(err).NotTo(HaveOccurred())
},
	Entry("from controller", "_projects/api-controller"),
	Entry("from application consuming just API", "_projects/api-client"),
)
