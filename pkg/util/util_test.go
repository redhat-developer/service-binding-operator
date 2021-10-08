package util

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Util Suite")
}

var _ = Describe("Merge Map", func() {

	It("should merge two maps", func() {
		m1 := map[string]string {
			"foo": "bar",
		}
		m2 := map[string]string {
			"bar": "foo",
		}
		m3 := MergeMaps(m1, m2)
		Expect(m3).To(Equal(map[string]string{
			"foo": "bar",
			"bar": "foo",
		}))
	})

	It("should return src map if dst is uninitialized", func() {
		var m1 map[string]string
		m2 := map[string]string {
			"bar": "foo",
		}
		m3 := MergeMaps(m1, m2)
		Expect(m3).To(Equal(m2))
	})
	It("should return dst map if src is uninitialized", func() {
		var m1 map[string]string
		m2 := map[string]string {
			"bar": "foo",
		}
		m3 := MergeMaps(m2, m1)
		Expect(m3).To(Equal(m2))
	})
})
