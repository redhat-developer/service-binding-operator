package collect_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBindingHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Binding Handlers Suite")
}
