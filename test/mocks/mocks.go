package mocks

import (
	"encoding/json"
	"fmt"

	pgv1alpha1 "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olminstall "github.com/operator-framework/operator-lifecycle-manager/pkg/controller/install"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
)

// resource details employed in mocks
const (
	CRDName            = "postgresql.baiju.dev"
	CRDVersion         = "v1alpha1"
	CRDKind            = "Database"
	OperatorKind       = "ServiceBindingRequest"
	OperatorAPIVersion = "apps.openshift.io/v1alpha1"
)

// ClusterServiceVersionMock based on PostgreSQL operator.
func ClusterServiceVersionMock(ns, name string) olmv1alpha1.ClusterServiceVersion {
	strategy := olminstall.StrategyDetailsDeployment{
		DeploymentSpecs: []olminstall.StrategyDeploymentSpec{{
			Name: "deployment",
			Spec: appsv1.DeploymentSpec{},
		}},
	}

	strategyJSON, _ := json.Marshal(strategy)

	return olmv1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: olmv1alpha1.ClusterServiceVersionSpec{
			DisplayName: "mock-database-csv",
			InstallStrategy: olmv1alpha1.NamedInstallStrategy{
				StrategyName:    "deployment",
				StrategySpecRaw: strategyJSON,
			},
			CustomResourceDefinitions: olmv1alpha1.CustomResourceDefinitions{
				Owned: []olmv1alpha1.CRDDescription{CRDDescriptionMock()},
			},
		},
	}
}

// ClusterServiceVersionListMock returns a list with a single CSV object inside, reusing mock.
func ClusterServiceVersionListMock(ns, name string) olmv1alpha1.ClusterServiceVersionList {
	return olmv1alpha1.ClusterServiceVersionList{
		Items: []olmv1alpha1.ClusterServiceVersion{ClusterServiceVersionMock(ns, name)},
	}
}

// CRDDescriptionMock based on PostgreSQL operator, returning a mock that defines database
// credentials entry with OLM descriptors.
func CRDDescriptionMock() olmv1alpha1.CRDDescription {
	return olmv1alpha1.CRDDescription{
		Name:        fmt.Sprintf("%s.%s", CRDKind, CRDName),
		DisplayName: CRDKind,
		Description: "mock-crd-description",
		Kind:        CRDKind,
		Version:     CRDVersion,
		SpecDescriptors: []olmv1alpha1.SpecDescriptor{{
			DisplayName:  "Database Name",
			Description:  "Database Name",
			Path:         "dbName",
			XDescriptors: []string{"urn:alm:descriptor:servicebindingrequest:env:attribute"},
		}},
		StatusDescriptors: []olmv1alpha1.StatusDescriptor{{
			DisplayName: "DB Password Credentials",
			Description: "Database credentials secret",
			Path:        "dbCredentials",
			XDescriptors: []string{
				"urn:alm:descriptor:io.kubernetes:Secret",
				"urn:alm:descriptor:servicebindingrequest:env:object:secret:user",
				"urn:alm:descriptor:servicebindingrequest:env:object:secret:password",
			},
		}},
	}
}

// DatabaseCRMock based on PostgreSQL operator, returning a instantiated object.
func DatabaseCRMock(ns, name string) pgv1alpha1.Database {
	return pgv1alpha1.Database{
		// usually TypeMeta should not be explicitly defined in mocked objects, however, on using
		// it via *unstructured.Unstructured it could not find this CR without it.
		TypeMeta: metav1.TypeMeta{
			Kind:       CRDKind,
			APIVersion: fmt.Sprintf("%s/%s", CRDName, CRDVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: pgv1alpha1.DatabaseSpec{
			Image:     "docker.io/postgres:latest",
			ImageName: "postgres",
		},
		Status: pgv1alpha1.DatabaseStatus{
			DBCredentials: "db-credentials",
		},
	}
}

// DatabaseCRListMock returns a list with a single database CR inside, reusing existing mock.
func DatabaseCRListMock(ns, name string) pgv1alpha1.DatabaseList {
	return pgv1alpha1.DatabaseList{
		Items: []pgv1alpha1.Database{DatabaseCRMock(ns, name)},
	}
}

// SecretMock returns a Secret based on PostgreSQL operator usage.
func SecretMock(ns, name string) corev1.Secret {
	return corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("password"),
		},
	}
}

// ServiceBindingRequestMock return a binding-request mock of informed name and match labels.
func ServiceBindingRequestMock(
	ns, name, resourceRef string, matchLabels map[string]string,
) v1alpha1.ServiceBindingRequest {
	return v1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			BackingServiceSelector: v1alpha1.BackingServiceSelector{
				ResourceName:    CRDName,
				ResourceVersion: CRDVersion,
				ResourceRef:     resourceRef,
			},
			ApplicationSelector: v1alpha1.ApplicationSelector{
				ResourceKind: "Deployment",
				MatchLabels:  matchLabels,
			},
		},
	}
}

// DeploymentMock creates a mocked Deployment object of busybox.
func DeploymentMock(ns, name string, matchLabels map[string]string) extv1beta1.Deployment {
	return extv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels:    matchLabels,
		},
		Spec: extv1beta1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      name,
					Labels:    matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "busybox",
						Image:   "busybox:latest",
						Command: []string{"sleep", "3600"},
					}},
				},
			},
		},
	}
}
