package naming_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestNamingHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Naming Handlers Suite")
}
