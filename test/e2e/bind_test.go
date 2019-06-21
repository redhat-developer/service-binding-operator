package e2e

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	"github.com/redhat-developer/service-binding-operator/pkg/apis"
	appsv1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	retryInterval        = time.Second * 5
	timeout              = time.Minute * 15
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestDeploy(t *testing.T) {
	sbr := &appsv1alpha1.ServiceBindingRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingRequest",
			APIVersion: "apps.openshift.io/v1alpha1",
		},
	}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, sbr)
	require.NoError(t, err, "Failed to add custom resource scheme to framework.")

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	require.NoError(t, err, "Failed to initialize cluster resources")
	t.Log("Initialized cluster resources")

	ctxNamespace, err := ctx.GetNamespace()
	require.NoError(t, err, "Failed to get namespace where operator needs to run")

	// get global framework variables
	f := framework.Global
	t.Logf("namespace: %s", ctxNamespace)
	testNamespace := os.Getenv("TEST_NAMESPACE")
	require.Equalf(t, testNamespace, ctxNamespace, "Test context namespace `%s` do not match the expected namespace '%s'", ctxNamespace, testNamespace)

	// wait for service-binding-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, ctxNamespace, "service-binding-operator", 1, retryInterval, timeout)
	require.NoError(t, err, "failed while waiting for operator deployment")
	t.Log("Service binding operator up and running")

}

func newBindedResource(namespace string) {
	// TODO: return the binded resource (postgresql db)
}

func newAppPod(namespace string) *corev1.Pod {
	labels := map[string]string{
		"app": "test-app",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-pod",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

func newServiceBindingRequestCR(namespace string) *appsv1alpha1.ServiceBindingRequest {
	return &appsv1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service-binding-request",
			Namespace: namespace,
		},
		Spec: appsv1alpha1.ServiceBindingRequestSpec{
			BackingSelector: appsv1alpha1.BackingSelector{
				ResourceName:    "test-resource-name",
				ResourceVersion: "test-resource-version",
			},
			ApplicationSelector: appsv1alpha1.ApplicationSelector{
				MatchLabels: map[string]string{
					"app": "test-app",
				},
				ResourceKind: "Pod",
			},
		},
	}
}
