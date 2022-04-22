package project_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-developer/service-binding-operator/apis"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/handler/project"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/mocks"
)

var _ = Describe("Inject Bindings as Env vars handler", func() {
	var (
		mockCtrl *gomock.Controller
		ctx      *mocks.MockContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("successful processing", func() {
		var (
			deploymentsUnstructured    []*unstructured.Unstructured
			deploymentsUnstructuredOld []*unstructured.Unstructured
			secretName                 string
		)

		BeforeEach(func() {
			var apps []pipeline.Application
			secretName = "secret1"
			ctx.EXPECT().BindAsFiles().Return(false).MinTimes(1)
			ctx.EXPECT().BindingSecretName().Return(secretName)
			d1 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d2 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "bla",
								},
							},
						},
					},
				},
			})

			for _, d := range []*appsv1.Deployment{d1, d2} {
				u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(d)
				if err != nil {
					Fail(err.Error())
				}
				res := &unstructured.Unstructured{Object: u}
				deploymentsUnstructured = append(deploymentsUnstructured, res)
				deploymentsUnstructuredOld = append(deploymentsUnstructuredOld, res.DeepCopy())
				app := mocks.NewMockApplication(mockCtrl)
				containers, _, _ := converter.NestedResources(&corev1.Container{}, u, strings.Split("spec.template.spec.containers", ".")...)
				var metaContainers []pipeline.MetaContainer
				for _, c := range containers {
					name, _, _ := unstructured.NestedString(c, "name")
					metaContainers = append(metaContainers, pipeline.MetaContainer{
						Data:        c,
						Name:        name,
						Env:         []string{"env"},
						EnvFrom:     []string{"envFrom"},
						VolumeMount: []string{"volumeMounts"},
					})
				}
				template := pipeline.MetaPodSpec{
					Containers: metaContainers,
					Volume:     strings.Split("spec.template.spec.volumes", "."),
					Data:       u,
				}
				app.EXPECT().BindablePods().Return(&template, nil)
				apps = append(apps, app)
			}

			ctx.EXPECT().Applications().Return(apps, nil)
			ctx.EXPECT().EnvBindings().Return([]*pipeline.EnvBinding{})
		})

		It("should inject secret ref in envFrom block", func() {
			project.BindingsAsEnv(ctx)
			for i, old := range deploymentsUnstructuredOld {
				Expect(deploymentsUnstructured[i]).NotTo(Equal(old))
			}
			expected := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
							},
						},
					},
				},
			})
			var d appsv1.Deployment
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(deploymentsUnstructured[0].Object, &d)
			Expect(err).NotTo(HaveOccurred())
			Expect(&d).To(Equal(expected))
			expected = deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "bla",
								},
							},
						},
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
							},
						},
					},
				},
			})
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(deploymentsUnstructured[1].Object, &d)
			Expect(err).NotTo(HaveOccurred())
			Expect(&d).To(Equal(expected))
		})

	})

	Context("env bindings are set", func() {
		var (
			deploymentsUnstructured    []*unstructured.Unstructured
			deploymentsUnstructuredOld []*unstructured.Unstructured
			secretName                 string
		)

		BeforeEach(func() {
			var apps []pipeline.Application
			secretName = "secret1"
			ctx.EXPECT().BindAsFiles().Return(true).MinTimes(1)
			ctx.EXPECT().BindingSecretName().Return(secretName)
			ctx.EXPECT().EnvBindings().Return([]*pipeline.EnvBinding{
				{
					Var:  "e1",
					Name: "b1",
				},
				{
					Var:  "e2",
					Name: "b2",
				},
			})
			d1 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					Name:  "d1",
				},
			})
			d2 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					Name:  "d2",
					Env: []corev1.EnvVar{
						{
							Name:  "foo",
							Value: "bla",
						},
					},
				},
			})

			for _, d := range []*appsv1.Deployment{d1, d2} {
				u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(d)
				if err != nil {
					Fail(err.Error())
				}
				res := &unstructured.Unstructured{Object: u}
				deploymentsUnstructured = append(deploymentsUnstructured, res)
				deploymentsUnstructuredOld = append(deploymentsUnstructuredOld, res.DeepCopy())
				app := mocks.NewMockApplication(mockCtrl)
				containers, _, _ := converter.NestedResources(&corev1.Container{}, u, strings.Split("spec.template.spec.containers", ".")...)
				var metaContainers []pipeline.MetaContainer
				for _, c := range containers {
					name, _, _ := unstructured.NestedString(c, "name")
					metaContainers = append(metaContainers, pipeline.MetaContainer{
						Data:        c,
						Name:        name,
						Env:         []string{"env"},
						EnvFrom:     []string{"envFrom"},
						VolumeMount: []string{"volumeMounts"},
					})
				}
				template := pipeline.MetaPodSpec{
					Containers: metaContainers,
					Volume:     strings.Split("spec.template.spec.volumes", "."),
					Data:       u,
				}
				app.EXPECT().BindablePods().Return(&template, nil)
				apps = append(apps, app)
			}

			ctx.EXPECT().Applications().Return(apps, nil)
		})

		It("should inject secret ref in env block", func() {
			project.BindingsAsEnv(ctx)
			for i, old := range deploymentsUnstructuredOld {
				Expect(deploymentsUnstructured[i]).NotTo(Equal(old))
			}
			expected := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					Name:  "d1",
					Env: []corev1.EnvVar{
						{
							Name: "e1",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "b1",
								},
							},
						},
						{
							Name: "e2",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "b2",
								},
							},
						},
					},
				},
			})
			var d appsv1.Deployment
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(deploymentsUnstructured[0].Object, &d)
			Expect(err).NotTo(HaveOccurred())
			Expect(&d).To(Equal(expected))
			expected = deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					Name:  "d2",
					Env: []corev1.EnvVar{
						{
							Name:  "foo",
							Value: "bla",
						},
						{
							Name: "e1",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "b1",
								},
							},
						},
						{
							Name: "e2",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: secretName,
									},
									Key: "b2",
								},
							},
						},
					},
				},
			})
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(deploymentsUnstructured[1].Object, &d)
			Expect(err).NotTo(HaveOccurred())
			Expect(&d).To(Equal(expected))
		})
	})

})

var _ = Describe("Injection Preflight checks", func() {
	var (
		mockCtrl *gomock.Controller
		ctx      *mocks.MockContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should retry processing if error when reading applications", func() {
		err := errors.New("foo")
		ctx.EXPECT().Applications().Return(nil, err)
		ctx.EXPECT().RetryProcessing(err)
		ctx.EXPECT().SetCondition(apis.Conditions().CollectionReady().DataCollected().Build())
		ctx.EXPECT().SetCondition(apis.Conditions().NotInjectionReady().ApplicationNotFound().Msg(err.Error()).Build())
		project.PreFlightCheck()(ctx)
	})

	It("should stop processing if no applications declared", func() {
		ctx.EXPECT().Applications().Return([]pipeline.Application{}, nil)
		ctx.EXPECT().StopProcessing()
		ctx.EXPECT().SetCondition(apis.Conditions().CollectionReady().DataCollected().Build())
		ctx.EXPECT().SetCondition(apis.Conditions().NotInjectionReady().Reason(apis.EmptyApplicationReason).Build())
		project.PreFlightCheck()(ctx)
	})

	It("should stop processing if mandatory bindings are missing", func() {
		err := errors.New("Mandatory binding 'type' not found")
		ctx.EXPECT().Applications().Return([]pipeline.Application{mocks.NewMockApplication(mockCtrl)}, nil)
		ctx.EXPECT().BindingItems().Return(pipeline.BindingItems{&pipeline.BindingItem{Name: "foo", Value: "val1"}, &pipeline.BindingItem{Name: "bar", Value: "val2"}})
		ctx.EXPECT().StopProcessing()
		ctx.EXPECT().Error(err)
		ctx.EXPECT().SetCondition(apis.Conditions().CollectionReady().DataCollected().Build())
		ctx.EXPECT().SetCondition(apis.Conditions().NotInjectionReady().Reason(apis.RequiredBindingNotFound).Msg(err.Error()).Build())
		project.PreFlightCheck("foo", "type")(ctx)
	})

	It("successful if all mandatory bindings are present", func() {
		ctx.EXPECT().Applications().Return([]pipeline.Application{mocks.NewMockApplication(mockCtrl)}, nil)
		ctx.EXPECT().BindingItems().Return(pipeline.BindingItems{&pipeline.BindingItem{Name: "foo", Value: "val1"}, &pipeline.BindingItem{Name: "bar", Value: "val2"}})
		ctx.EXPECT().SetCondition(apis.Conditions().CollectionReady().DataCollected().Build())
		project.PreFlightCheck("foo", "bar")(ctx)
	})
})

var _ = Describe("Injection Postflight checks", func() {
	var (
		mockCtrl *gomock.Controller
		ctx      *mocks.MockContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should trigger rebinding when labels are present", func() {
		ctx.EXPECT().SetCondition(apis.Conditions().InjectionReady().Reason("ApplicationUpdated").Build())
		ctx.EXPECT().HasLabelSelector().Return(true)
		ctx.EXPECT().DelayReprocessing(nil)
		project.PostFlightCheck(ctx)
	})

	It("should not trigger rebinding when labels are not present", func() {
		ctx.EXPECT().SetCondition(apis.Conditions().InjectionReady().Reason("ApplicationUpdated").Build())
		ctx.EXPECT().HasLabelSelector().Return(false)
		project.PostFlightCheck(ctx)
	})
})

var _ = Describe("Inject bindings as files", func() {
	var (
		mockCtrl *gomock.Controller
		ctx      *mocks.MockContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
		ctx.EXPECT().BindAsFiles().Return(true)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("successful processing", func() {
		var (
			deploymentsUnstructured    []*unstructured.Unstructured
			deploymentsUnstructuredOld []*unstructured.Unstructured
			secretName                 string
			bindingName                string
		)

		BeforeEach(func() {
			var apps []pipeline.Application
			secretName = "secret1"
			bindingName = "sb1"
			ctx.EXPECT().BindingSecretName().Return(secretName)
			d1 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d2 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/foo",
						},
					},
				},
			})
			d3 := deployment("d3", []corev1.Container{
				{
					Image: "foo2",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/foo",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "existing mount",
							MountPath: "/mount",
						},
					},
				},
			})
			d4 := deployment("d4", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d4.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "existing volume",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/some/path",
						},
					},
				},
			}
			d5 := deployment("d5", []corev1.Container{
				{
					Image: "foo",
					Env: []corev1.EnvVar{
						{
							Name:  "SOME_ENV",
							Value: "val1",
						},
					},
				},
			})
			d6 := deployment("d6", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d6.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			d7 := deployment("d6", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d7.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "oldsecret",
						},
					},
				},
			}
			d8 := deployment("d6", []corev1.Container{
				{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bla",
						},
					},
				},
			})
			d9 := deployment("d6", []corev1.Container{
				{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/foo",
						},
					},
				},
			})
			for _, d := range []*appsv1.Deployment{d1, d2, d3, d4, d5, d6, d7, d8, d9} {
				u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(d)
				if err != nil {
					Fail(err.Error())
				}
				res := &unstructured.Unstructured{Object: u}
				deploymentsUnstructured = append(deploymentsUnstructured, res)
				deploymentsUnstructuredOld = append(deploymentsUnstructuredOld, res.DeepCopy())
				app := mocks.NewMockApplication(mockCtrl)
				containers, _, _ := converter.NestedResources(&corev1.Container{}, u, strings.Split("spec.template.spec.containers", ".")...)
				var metaContainers []pipeline.MetaContainer
				for _, c := range containers {
					name, _, _ := unstructured.NestedString(c, "name")
					metaContainers = append(metaContainers, pipeline.MetaContainer{
						Data:        c,
						Name:        name,
						Env:         []string{"env"},
						EnvFrom:     []string{"envFrom"},
						VolumeMount: []string{"volumeMounts"},
					})
				}
				template := pipeline.MetaPodSpec{
					Containers: metaContainers,
					Volume:     strings.Split("spec.template.spec.volumes", "."),
					Data:       u,
				}
				app.EXPECT().BindablePods().Return(&template, nil)
				apps = append(apps, app)
			}

			ctx.EXPECT().Applications().Return(apps, nil)
			ctx.EXPECT().BindingName().Return(bindingName).AnyTimes()
		})
		It("should mount binding secret as volume", func() {
			project.BindingsAsFiles(ctx)
			for i, old := range deploymentsUnstructuredOld {
				Expect(deploymentsUnstructured[i]).NotTo(Equal(old))
			}
			expectedD1 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/bindings",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bindings/sb1",
						},
					},
				},
			})
			expectedD1.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			expectedD2 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/foo",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/foo/sb1",
						},
					},
				},
			})
			expectedD2.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			expectedD3 := deployment("d3", []corev1.Container{
				{
					Image: "foo2",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/foo",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "existing mount",
							MountPath: "/mount",
						},
						{
							Name:      bindingName,
							MountPath: "/foo/sb1",
						},
					},
				},
			})
			expectedD3.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}

			expectedD4 := deployment("d4", []corev1.Container{
				{
					Image: "foo",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/bindings",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bindings/sb1",
						},
					},
				},
			})
			expectedD4.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "existing volume",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/some/path",
						},
					},
				},
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			expectedD5 := deployment("d5", []corev1.Container{
				{
					Image: "foo",
					Env: []corev1.EnvVar{
						{
							Name:  "SOME_ENV",
							Value: "val1",
						},
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/bindings",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bindings/sb1",
						},
					},
				},
			})
			expectedD5.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			expectedD6 := deployment("d6", []corev1.Container{
				{
					Image: "foo",
					Env: []corev1.EnvVar{
						{
							Name:  "SERVICE_BINDING_ROOT",
							Value: "/bindings",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bindings/sb1",
						},
					},
				},
			})
			expectedD6.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			for i, expectedDeployment := range []*appsv1.Deployment{expectedD1, expectedD2, expectedD3, expectedD4, expectedD5, expectedD6, expectedD6, expectedD6, expectedD6} {
				d := &appsv1.Deployment{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(deploymentsUnstructured[i].Object, d)
				Expect(err).NotTo(HaveOccurred())
				Expect(d).To(Equal(expectedDeployment), fmt.Sprintf("%v", i))
			}

		})
	})
})

var _ = Describe("Unbind handler", func() {
	var (
		mockCtrl *gomock.Controller
		ctx      *mocks.MockContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = mocks.NewMockContext(mockCtrl)
		ctx.EXPECT().UnbindRequested().Return(true)
		ctx.EXPECT().StopProcessing()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should stop processing when there is error getting applications", func() {
		ctx.EXPECT().Applications().Return(nil, errors.New("foo"))
		project.Unbind(ctx)
	})
	It("should stop processing when there are no applications", func() {
		ctx.EXPECT().Applications().Return([]pipeline.Application{}, nil)
		project.Unbind(ctx)
	})
	Context("successful processing", func() {
		var (
			deploymentsUnstructured    []*unstructured.Unstructured
			deploymentsUnstructuredOld []*unstructured.Unstructured
			secretName                 string
			bindingName                string
		)

		BeforeEach(func() {
			var apps []pipeline.Application
			secretName = "secret1"
			bindingName = "binding1"
			ctx.EXPECT().BindingSecretName().Return(secretName)
			ctx.EXPECT().BindingName().Return(bindingName)
			ctx.EXPECT().EnvBindings().Return([]*pipeline.EnvBinding{})
			d1 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d2 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "bla",
								},
							},
						},
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
							},
						},
					},
				},
			})
			d3 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secretName,
								},
							},
						},
					},
				},
			})
			d4 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bla",
						},
					},
				},
			})
			d5 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      bindingName,
							MountPath: "/bla",
						},
						{
							Name:      "bla",
							MountPath: "/bla2",
						},
					},
				},
			})
			d6 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d6.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
			}
			d7 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d7.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: bindingName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secretName,
						},
					},
				},
				{
					Name: "foo",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "foo",
						},
					},
				},
			}
			for _, d := range []*appsv1.Deployment{d1, d2, d3, d4, d5, d6, d7} {
				u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(d)
				if err != nil {
					Fail(err.Error())
				}
				res := &unstructured.Unstructured{Object: u}
				deploymentsUnstructured = append(deploymentsUnstructured, res)
				deploymentsUnstructuredOld = append(deploymentsUnstructuredOld, res.DeepCopy())
				app := mocks.NewMockApplication(mockCtrl)
				containers, _, _ := converter.NestedResources(&corev1.Container{}, u, strings.Split("spec.template.spec.containers", ".")...)
				var metaContainers []pipeline.MetaContainer
				for _, c := range containers {
					name, _, _ := unstructured.NestedString(c, "name")
					metaContainers = append(metaContainers, pipeline.MetaContainer{
						Data:        c,
						Name:        name,
						Env:         []string{"env"},
						EnvFrom:     []string{"envFrom"},
						VolumeMount: []string{"volumeMounts"},
					})
				}
				template := pipeline.MetaPodSpec{
					Containers: metaContainers,
					Volume:     strings.Split("spec.template.spec.volumes", "."),
					Data:       u,
				}
				app.EXPECT().BindablePods().Return(&template, nil)
				apps = append(apps, app)
			}

			ctx.EXPECT().Applications().Return(apps, nil)
		})
		It("should remove secret refs", func() {
			project.Unbind(ctx)
			Expect(deploymentsUnstructured[0]).To(Equal(deploymentsUnstructuredOld[0]))
			Expect(deploymentsUnstructured[1]).NotTo(Equal(deploymentsUnstructuredOld[1]))
			Expect(deploymentsUnstructured[2]).NotTo(Equal(deploymentsUnstructuredOld[2]))
			Expect(deploymentsUnstructured[3]).NotTo(Equal(deploymentsUnstructuredOld[3]))
			Expect(deploymentsUnstructured[4]).NotTo(Equal(deploymentsUnstructuredOld[4]))
			Expect(deploymentsUnstructured[5]).NotTo(Equal(deploymentsUnstructuredOld[5]))
			Expect(deploymentsUnstructured[6]).NotTo(Equal(deploymentsUnstructuredOld[6]))
			d2 := deployment("d2", []corev1.Container{
				{
					Image: "foo2",
					EnvFrom: []corev1.EnvFromSource{
						{
							SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "bla",
								},
							},
						},
					},
				},
			})
			d3 := deployment("d2", []corev1.Container{
				{
					Image:   "foo2",
					EnvFrom: []v1.EnvFromSource{},
				},
			})
			d4 := deployment("d1", []corev1.Container{
				{
					Image:        "foo",
					VolumeMounts: []v1.VolumeMount{},
				},
			})
			d5 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "bla",
							MountPath: "/bla2",
						},
					},
				},
			})
			d6 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d6.Spec.Template.Spec.Volumes = []corev1.Volume{}
			d7 := deployment("d1", []corev1.Container{
				{
					Image: "foo",
				},
			})
			d7.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: "foo",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "foo",
						},
					},
				},
			}
			for i, expectedDeployment := range []*appsv1.Deployment{nil, d2, d3, d4, d5, d6, d7} {
				if expectedDeployment == nil {
					continue
				}
				d := &appsv1.Deployment{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(deploymentsUnstructured[i].Object, d)
				Expect(err).NotTo(HaveOccurred())
				Expect(d.Spec.Template).To(Equal(expectedDeployment.Spec.Template))
			}
		})
	})
})

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
