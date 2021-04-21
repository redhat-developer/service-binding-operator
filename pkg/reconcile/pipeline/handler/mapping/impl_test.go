package mapping_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/mapping"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Mapping Handler", func() {

	var (
		mockCtrl     *gomock.Controller
		ctx          *mocks.MockContext
		services     []*mocks.MockService
		bindingItems pipeline.BindingItems
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
		services = []*mocks.MockService{mocks.NewMockService(mockCtrl), mocks.NewMockService(mockCtrl), mocks.NewMockService(mockCtrl)}
		ctx.EXPECT().Services().Return([]pipeline.Service{services[0], services[1], services[2]}, nil)
		bindingItems = []*pipeline.BindingItem{
			{
				Name:  "foo",
				Value: "val1",
			},
			{
				Name:  "foo2",
				Value: "val2",
			},
		}
		ctx.EXPECT().BindingItems().Return(bindingItems)
		u := &unstructured.Unstructured{Object: map[string]interface{}{
			"bar": "bla",
		}}
		srvId := "id1"
		services[0].EXPECT().Id().Return(&srvId).MinTimes(1)
		services[0].EXPECT().Resource().Return(u)

		u2 := &unstructured.Unstructured{Object: map[string]interface{}{
			"bar2": "bla2",
		}}
		u2.SetName("n1")
		u2.SetNamespace("ns1")

		srvId2 := "id2"
		services[2].EXPECT().Id().Return(&srvId2).MinTimes(1)
		services[2].EXPECT().Resource().Return(u2)

		services[1].EXPECT().Id().Return(nil).MinTimes(1)
	})

	DescribeTable("successful processing", func(template string, expected string) {

		mappings := map[string]string{
			"foo3": template,
		}
		ctx.EXPECT().Mappings().Return(mappings)
		ctx.EXPECT().AddBindingItem(&pipeline.BindingItem{Name: "foo3", Value: expected})

		mapping.Handle(ctx)
	},
		Entry("property referred via service id", "{{ .id1.bar }}", "bla"),
		Entry("property referred via service id", "{{ .id1.bar }}_{{ .id2.metadata.name }}", "bla_n1"),
		Entry("use existing bindins", "{{ .foo }}_{{ .foo2 }}", "val1_val2"))

	DescribeTable("failed processing", func(template string) {

		mappings := map[string]string{
			"foo3": template,
		}
		ctx.EXPECT().Mappings().Return(mappings)
		ctx.EXPECT().StopProcessing()
		ctx.EXPECT().Error(gomock.Any())

		mapping.Handle(ctx)
	},
		Entry("bad template", "{{ .id1.bar "))

})
