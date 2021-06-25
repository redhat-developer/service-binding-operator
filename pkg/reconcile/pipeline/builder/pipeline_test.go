package builder_test

import (
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	v1alpha12 "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pipeline", func() {
	var (
		mockCtrl   *gomock.Controller
		ctx        *mocks.MockContext
		defHandler = func() *mocks.MockHandler {
			return mocks.NewMockHandler(mockCtrl)
		}
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should not retry if processing was successful", func() {
		h1 := defHandler()
		h1.EXPECT().Handle(ctx)
		h2 := defHandler()
		h2.EXPECT().Handle(ctx)
		p := builder.Builder().WithContextProvider(&ctxProvider{ctx: ctx}).WithHandlers(h1, h2).Build()

		ctx.EXPECT().Close().Return(nil)
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{}).Times(2)

		retry, err := p.Process(&v1alpha12.ServiceBinding{})
		Expect(err).NotTo(HaveOccurred())
		Expect(retry).To(BeFalse())
	})

	It("should report error in case of ctx closing error", func() {
		h1 := defHandler()
		h1.EXPECT().Handle(ctx)
		h2 := defHandler()
		h2.EXPECT().Handle(ctx)
		p := builder.Builder().WithContextProvider(&ctxProvider{ctx: ctx}).WithHandlers(h1, h2).Build()

		err := errors.New("foo")
		ctx.EXPECT().Close().Return(err)
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{}).Times(2)

		retry, err := p.Process(&v1alpha12.ServiceBinding{})
		Expect(err).To(Equal(err))
		Expect(retry).To(BeTrue())
	})

	It("should stop processing if retry requested and propagate error back to caller", func() {
		err := errors.New("foo")

		h1 := defHandler()
		h1.EXPECT().Handle(ctx)
		h2 := func(c pipeline.Context) {
			c.RetryProcessing(err)
		}
		h3 := defHandler()
		p := builder.Builder().WithContextProvider(&ctxProvider{ctx: ctx}).WithHandlers(h1, pipeline.HandlerFunc(h2), h3).Build()

		ctx.EXPECT().RetryProcessing(err)
		ctx.EXPECT().Close().Return(nil)
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{})
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{Retry: true, Stop: true, Err: err})

		retry, err := p.Process(&v1alpha12.ServiceBinding{})
		Expect(err).To(Equal(err))
		Expect(retry).To(BeTrue())
	})

	It("should stop without retry and error and propagate that back to caller", func() {
		h1 := defHandler()
		h1.EXPECT().Handle(ctx)
		h2 := func(c pipeline.Context) {
			c.StopProcessing()
		}
		h3 := defHandler()
		p := builder.Builder().WithContextProvider(&ctxProvider{ctx: ctx}).WithHandlers(h1, pipeline.HandlerFunc(h2), h3).Build()

		ctx.EXPECT().StopProcessing()
		ctx.EXPECT().Close().Return(nil)
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{})
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{Retry: false, Stop: true, Err: nil})

		retry, err := p.Process(&v1alpha12.ServiceBinding{})
		Expect(err).NotTo(HaveOccurred())
		Expect(retry).To(BeFalse())
	})

	It("error on closing context should be propagated back to caller even if handlers did not report any", func() {
		var err = errors.New("foo")
		h1 := defHandler()
		h1.EXPECT().Handle(ctx)
		h2 := defHandler()
		h2.EXPECT().Handle(ctx)
		p := builder.Builder().WithContextProvider(&ctxProvider{ctx: ctx}).WithHandlers(h1, h2).Build()

		ctx.EXPECT().Close().Return(err)
		ctx.EXPECT().FlowStatus().Return(pipeline.FlowStatus{}).Times(2)

		retry, err := p.Process(&v1alpha12.ServiceBinding{})
		Expect(err).To(Equal(err))
		Expect(retry).To(BeTrue())
	})
})

type ctxProvider struct {
	ctx pipeline.Context
}

func (c *ctxProvider) Get(binding *v1alpha12.ServiceBinding) (pipeline.Context, error) {
	return c.ctx, nil
}
