package pipeline_test

import (
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

var _ = Describe("Binding Items", func() {
	DescribeTable("returned map contains (binding name, binding values) pairs", func(expectedMap map[string]string, items ...*pipeline.BindingItem) {

		bindingItems := pipeline.BindingItems(items)

		Expect(bindingItems.AsMap()).To(Equal(expectedMap))
	},
		Entry("empty map", map[string]string{}),
		Entry("two entries",
			map[string]string{
				"foo":  "v1",
				"foo2": "v2",
			},
			&pipeline.BindingItem{Name: "foo", Value: "v1"}, &pipeline.BindingItem{Name: "foo2", Value: "v2"}),
		Entry("two entries with the same name",
			map[string]string{
				"foo2": "v2",
			},
			&pipeline.BindingItem{Name: "foo2", Value: "v1"}, &pipeline.BindingItem{Name: "foo2", Value: "v2"}),
		Entry("entry with non-string type",
			map[string]string{
				"foo2": "2",
			}, &pipeline.BindingItem{Name: "foo2", Value: 2}),
	)
})
