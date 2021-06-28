package naming_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/naming"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Naming handler", func() {
	var (
		mockCtrl     *gomock.Controller
		ctx          *mocks.MockContext
		service      *mocks.MockService
		bindingItems pipeline.BindingItems
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)

		service = mocks.NewMockService(mockCtrl)
		serviceRes := &unstructured.Unstructured{}
		serviceRes.SetKind("Foo")
		serviceRes.SetName("bar")
		service.EXPECT().Resource().Return(serviceRes).MinTimes(1)
		bd1 := &pipeline.BindingItem{
			Name:   "name1",
			Value:  "val1",
			Source: service,
		}
		bd2 := &pipeline.BindingItem{
			Name:   "name2",
			Value:  "val2",
			Source: service,
		}
		bd3 := &pipeline.BindingItem{
			Name:  "name3",
			Value: "val3",
		}
		bindingItems = pipeline.BindingItems{bd1, bd2, bd3}
		ctx.EXPECT().BindingItems().Return(bindingItems)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	type testCase struct {
		template      string
		expectedNames []string
	}

	DescribeTable("assign new names",
		func(tc testCase) {
			ctx.EXPECT().NamingTemplate().Return(tc.template).MinTimes(1)
			naming.Handle(ctx)
			for i, bd := range bindingItems {
				Expect(bd.Name).To(Equal(tc.expectedNames[i]))
			}
		},
		Entry("add prefix", testCase{
			template: "prefix_{{ .name }}",
			expectedNames: []string{
				"prefix_name1",
				"prefix_name2",
				"name3",
			},
		}),
		Entry("add service kind prefix", testCase{
			template: "{{ .service.kind }}_{{ .name }}",
			expectedNames: []string{
				"Foo_name1",
				"Foo_name2",
				"name3",
			},
		}),
		Entry("add uppercased service kind prefix", testCase{
			template: "{{ .service.kind | upper }}_{{ .name }}",
			expectedNames: []string{
				"FOO_name1",
				"FOO_name2",
				"name3",
			},
		}),
		Entry("add service name prefix", testCase{
			template: "{{ .service.name }}_{{ .name }}",
			expectedNames: []string{
				"bar_name1",
				"bar_name2",
				"name3",
			},
		}),
	)

	DescribeTable("template failure stops processing",
		func(template string) {
			ctx.EXPECT().NamingTemplate().Return(template).MinTimes(1)
			ctx.EXPECT().StopProcessing()
			var err error
			ctx.EXPECT().Error(gomock.Any()).Do(func(e error) { err = e })

			ctx.EXPECT().SetCondition(gomock.Any()).Do(func(condition *v1.Condition) {
				Expect(condition).To(Equal(apis.Conditions().NotCollectionReady().Reason(naming.StrategyError).Msg(err.Error()).Build()))
			})
			naming.Handle(ctx)
		},
		Entry("not valid template", "{{ .name"),
		Entry("non existing function", "{{ .name | foo }}"),
		Entry("non existing property", "{{ .wrong | upper }}"),
	)
})
