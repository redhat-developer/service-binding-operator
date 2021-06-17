package builder_test

import (
	c "context"
	"reflect"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/service-binding-operator/api/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/builder"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default Pipeline", func() {

	var (
		mockCtrl *gomock.Controller
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should bind service to app successfully", func() {
		ns := "ns1"
		serviceName := "s1"
		serviceGVR := schema.GroupVersionResource{Group: "services", Version: "v1", Resource: "databases"}
		serviceGVK := serviceGVR.GroupVersion().WithKind("Database")
		serviceRef := v1alpha1.Service{
			NamespacedRef: v1alpha1.NamespacedRef{
				Ref: v1alpha1.Ref{
					Group:    serviceGVR.Group,
					Version:  serviceGVR.Version,
					Resource: serviceGVR.Resource,
					Name:     serviceName,
				},
			},
		}
		appGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
		appGVK := appGVR.GroupVersion().WithKind("Deployment")
		appName := "app1"
		appRef := v1alpha1.Application{
			Ref: v1alpha1.Ref{
				Group:    appGVR.Group,
				Version:  appGVR.Version,
				Resource: appGVR.Resource,
				Name:     appName,
			},
		}
		sb := &v1alpha1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sb1",
				Namespace: ns,
				UID:       "uid1",
			},
			Spec: v1alpha1.ServiceBindingSpec{
				BindAsFiles: false,
				Services:    []v1alpha1.Service{serviceRef},
				Application: &appRef,
			},
		}
		sb.SetGroupVersionKind(v1alpha1.GroupVersionKind)

		service := &unstructured.Unstructured{Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"service.binding":      "path={.status.foo}",
					"service.binding/bar2": "path={.status.foo2}",
				},
			},
			"status": map[string]interface{}{
				"foo":  "val1",
				"foo2": "val2",
				"foo3": "val3",
			},
		}}
		service.SetName(serviceName)
		service.SetNamespace(ns)
		service.SetGroupVersionKind(serviceGVK)

		app := deployment(appName, []corev1.Container{
			{
				Image: "foo",
			},
		})
		app.SetNamespace(ns)
		app.SetGroupVersionKind(appGVK)

		appUnstructured, err := converter.ToUnstructured(app)
		Expect(err).NotTo(HaveOccurred())
		sbUnstructured, err := converter.ToUnstructured(sb)
		Expect(err).NotTo(HaveOccurred())

		client := client(service, appUnstructured, sbUnstructured)

		typeLookup := mocks.NewMockK8STypeLookup(mockCtrl)
		typeLookup.EXPECT().ResourceForReferable(gomock.Any()).DoAndReturn(func(r kubernetes.Referable) (*schema.GroupVersionResource, error) {
			if reflect.DeepEqual(r, &appRef) {
				return &appGVR, nil
			} else {
				return &serviceGVR, nil
			}
		}).MinTimes(1)

		p := builder.DefaultBuilder.WithContextProvider(context.Provider(client, typeLookup)).Build()

		retry, err := p.Process(sb)
		Expect(err).NotTo(HaveOccurred())
		Expect(retry).To(BeFalse())

		u, err := client.Resource(v1alpha1.GroupVersionResource).Namespace(sb.Namespace).Get(c.Background(), sb.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		updatedSB := v1alpha1.ServiceBinding{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &updatedSB)

		Expect(err).NotTo(HaveOccurred())
		Expect(updatedSB.Status.Secret).NotTo(BeEmpty())
		Expect(updatedSB.Status.Conditions).To(HaveLen(3))
		Expect(existCondition(updatedSB.Status.Conditions, v1alpha1.BindingReady, metav1.ConditionTrue)).To(BeTrue())
		Expect(existCondition(updatedSB.Status.Conditions, v1alpha1.InjectionReady, metav1.ConditionTrue)).To(BeTrue())
		Expect(existCondition(updatedSB.Status.Conditions, v1alpha1.CollectionReady, metav1.ConditionTrue)).To(BeTrue())

		u, err = client.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}).Namespace(sb.Namespace).Get(c.Background(), updatedSB.Status.Secret, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		secret := &corev1.Secret{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, secret)
		Expect(err).NotTo(HaveOccurred())
		Expect(secret.StringData).To(Equal(map[string]string{
			"DATABASE_FOO":  "val1",
			"DATABASE_BAR2": "val2",
		}))

		u, err = client.Resource(appGVR).Namespace(sb.Namespace).Get(c.Background(), appName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		updatedApp := &appsv1.Deployment{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, updatedApp)
		Expect(err).NotTo(HaveOccurred())
		Expect(updatedApp.Spec.Template.Spec.Containers[0].EnvFrom[0].SecretRef.Name).To(Equal(updatedSB.Status.Secret))
	})
})

func existCondition(conditions []metav1.Condition, conditionType string, status metav1.ConditionStatus) bool {
	for _, c := range conditions {
		if c.Status == status && c.Type == conditionType {
			return true
		}
	}
	return false
}

func client(objs ...runtime.Object) *fake.FakeDynamicClient {
	return fake.NewSimpleDynamicClient(runtime.NewScheme(), objs...)
}

func deployment(name string, containers []corev1.Container) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
}
