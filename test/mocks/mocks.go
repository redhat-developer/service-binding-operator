package mocks

import (
	"encoding/json"
	"fmt"

	dbv1alpha1 "github.com/baijum/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olminstall "github.com/operator-framework/operator-lifecycle-manager/pkg/controller/install"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// names employed in mocks
const (
	CRDName    = "postgresql.baiju.dev"
	CRDVersion = "v1alpha1"
	CRDKind    = "Database"
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterServiceVersion",
			APIVersion: "operators.coreos.com/v1alpha1",
		},
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

// CRDDescriptionMock based on PostgreSQL operator, returning a mock that defines database
// credentials entry with OLM descriptors.
func CRDDescriptionMock() olmv1alpha1.CRDDescription {
	return olmv1alpha1.CRDDescription{
		Name:        CRDName,
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

// DatabaseCRDMock based on PostgreSQL operator, returning a instantiated object
func DatabaseCRDMock(ns, name string) dbv1alpha1.Database {
	return dbv1alpha1.Database{
		TypeMeta: metav1.TypeMeta{
			Kind:       CRDKind,
			APIVersion: fmt.Sprintf("%s/%s", CRDName, CRDVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: dbv1alpha1.DatabaseSpec{
			Image:     "docker.io/postgres:latest",
			ImageName: "postgres",
		},
		Status: dbv1alpha1.DatabaseStatus{
			DBCredentials: "db-credentials",
		},
	}
}

// SecretMock returns a Secret based on PostgreSQL operator usage.
func SecretMock(ns, name string) corev1.Secret {
	return corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
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
