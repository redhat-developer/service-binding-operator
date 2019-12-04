package mocks

import (
	"encoding/json"
	"fmt"
	"strings"

	ocv1 "github.com/openshift/api/route/v1"
	pgv1alpha1 "github.com/operator-backing-service-samples/postgresql-operator/pkg/apis/postgresql/v1alpha1"
	olmv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olminstall "github.com/operator-framework/operator-lifecycle-manager/pkg/controller/install"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	"github.com/redhat-developer/service-binding-operator/pkg/converter"
)

// resource details employed in mocks
const (
	CRDName            = "postgresql.baiju.dev"
	CRDVersion         = "v1alpha1"
	CRDKind            = "Database"
	OperatorKind       = "ServiceBindingRequest"
	OperatorAPIVersion = "apps.openshift.io/v1alpha1"
)

var (
	// DBNameSpecDesc default spec descriptor to inform the database name.
	DBNameSpecDesc = olmv1alpha1.SpecDescriptor{
		DisplayName:  "Database Name",
		Description:  "Database Name",
		Path:         "dbName",
		XDescriptors: []string{"binding:env:attribute"},
	}
	ImageSpecDesc = olmv1alpha1.SpecDescriptor{
		Path:         "image",
		DisplayName:  "Image",
		Description:  "Image Name",
		XDescriptors: nil,
	}
	// DBNameSpecDesc default spec descriptor to inform the database name.
	DBNameSpecIp = olmv1alpha1.SpecDescriptor{
		DisplayName:  "Database IP",
		Description:  "Database IP",
		Path:         "dbConnectionIp",
		XDescriptors: []string{"binding:env:attribute"},
	}
	// DBConfigMapSpecDesc spec descriptor to describe a operator that export username and password
	// via config-map, instead of a usual secret.
	DBConfigMapSpecDesc = olmv1alpha1.SpecDescriptor{
		DisplayName: "DB ConfigMap",
		Description: "Database ConfigMap",
		Path:        "dbConfigMap",
		XDescriptors: []string{
			"urn:alm:descriptor:io.kubernetes:ConfigMap",
			"binding:env:object:configmap:user",
			"binding:env:object:configmap:password",
		},
	}
	// DBPasswordCredentialsOnEnvStatusDesc status descriptor to describe a database operator that
	// publishes username and password over a secret. Default approach.
	DBPasswordCredentialsOnEnvStatusDesc = olmv1alpha1.StatusDescriptor{
		DisplayName: "DB Password Credentials",
		Description: "Database credentials secret",
		Path:        "dbCredentials",
		XDescriptors: []string{
			"urn:alm:descriptor:io.kubernetes:Secret",
			"binding:env:object:secret:user",
			"binding:env:object:secret:password",
		},
	}
	// DBPasswordCredentialsOnVolumeMountStatusDesc status descriptor to describe a operator that
	// informs credentials via a volume.
	DBPasswordCredentialsOnVolumeMountStatusDesc = olmv1alpha1.StatusDescriptor{
		DisplayName: "DB Password Credentials",
		Description: "Database credentials secret",
		Path:        "dbCredentials",
		XDescriptors: []string{
			"urn:alm:descriptor:io.kubernetes:Secret",
			"binding:volumemount:secret:user",
			"binding:volumemount:secret:password",
		},
	}
)

func DatabaseCRDMock(ns string) apiextensionv1beta1.CustomResourceDefinition {
	CRDPlural := "databases"
	FullCRDName := CRDPlural + "." + CRDName
	annotations := map[string]string{
		"servicebindingoperator.redhat.io/status.dbConfigMap-password": "binding:env:object:configmap",
	}

	crd := apiextensionv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   ns,
			Name:        FullCRDName,
			Annotations: annotations,
		},
		Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
			Group:   CRDName,
			Version: CRDVersion,
			Scope:   apiextensionv1beta1.NamespaceScoped,
			Names: apiextensionv1beta1.CustomResourceDefinitionNames{
				Plural: CRDPlural,
				Kind:   CRDKind,
			},
		},
	}

	return crd
}

func UnstructuredDatabaseCRDMock(ns string) (*unstructured.Unstructured, error) {
	crd := DatabaseCRDMock(ns)
	return converter.ToUnstructured(&crd)
}

type PostgresDatabaseSpec struct {
	Username string `json:"username"`
}

type PostgresDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PostgresDatabaseSpec `json:"spec,omitempty"`
}

func PostgresDatabaseCRMock(ns, name string) PostgresDatabase {
	return PostgresDatabase{
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
		Spec: PostgresDatabaseSpec{
			Username: "redhatdeveloper",
		},
	}
}

func UnstructuredPostgresDatabaseCRMock(ns, name string) (*unstructured.Unstructured, error) {
	c := PostgresDatabaseCRMock(ns, name)
	return converter.ToUnstructured(&c)
}

//
// Usage of TypeMeta in Mocks
//
// 	Usually TypeMeta should not be explicitly defined in mocked objects, however, on using
//  it via *unstructured.Unstructured it could not find this CR without it.
//

// crdDescriptionMock based for mocked objects.
func crdDescriptionMock(
	specDescriptor []olmv1alpha1.SpecDescriptor,
	statusDescriptors []olmv1alpha1.StatusDescriptor,
) olmv1alpha1.CRDDescription {
	return olmv1alpha1.CRDDescription{
		Name:              strings.ToLower(fmt.Sprintf("%s.%s", CRDKind, CRDName)),
		DisplayName:       CRDKind,
		Description:       "mock-crd-description",
		Kind:              CRDKind,
		Version:           CRDVersion,
		SpecDescriptors:   specDescriptor,
		StatusDescriptors: statusDescriptors,
	}
}

// ClusterServiceVersionListMock returns a list with a single CSV object inside, reusing mock.
func ClusterServiceVersionListMock(ns, name string) *olmv1alpha1.ClusterServiceVersionList {
	return &olmv1alpha1.ClusterServiceVersionList{
		Items: []olmv1alpha1.ClusterServiceVersion{ClusterServiceVersionMock(ns, name)},
	}
}

// CRDDescriptionMock based on PostgreSQL operator, returning a mock using default third party
// operator setup.
func CRDDescriptionMock() olmv1alpha1.CRDDescription {
	return crdDescriptionMock(
		[]olmv1alpha1.SpecDescriptor{DBNameSpecDesc, ImageSpecDesc},
		[]olmv1alpha1.StatusDescriptor{DBPasswordCredentialsOnEnvStatusDesc},
	)
}

// CRDDescriptionConfigMapMock based on PostgreSQL operator, returns a mock using configmap based
// spec-descriptor
func CRDDescriptionConfigMapMock() olmv1alpha1.CRDDescription {
	return crdDescriptionMock(
		[]olmv1alpha1.SpecDescriptor{DBConfigMapSpecDesc, ImageSpecDesc},
		[]olmv1alpha1.StatusDescriptor{DBPasswordCredentialsOnEnvStatusDesc},
	)
}

// CRDDescriptionVolumeMountMock based on PostgreSQL operator, returns a mock having credentials
// in a volume.
func CRDDescriptionVolumeMountMock() olmv1alpha1.CRDDescription {
	return crdDescriptionMock(
		[]olmv1alpha1.SpecDescriptor{DBNameSpecDesc},
		[]olmv1alpha1.StatusDescriptor{DBPasswordCredentialsOnVolumeMountStatusDesc},
	)
}

// clusterServiceVersionMock base object to create a CSV.
func clusterServiceVersionMock(
	ns,
	name string,
	crdDescription olmv1alpha1.CRDDescription,
) olmv1alpha1.ClusterServiceVersion {
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
				Owned: []olmv1alpha1.CRDDescription{crdDescription},
			},
		},
	}
}

// ClusterServiceVersionMock based on PostgreSQL operator having what's expected as defaults.
func ClusterServiceVersionMock(ns, name string) olmv1alpha1.ClusterServiceVersion {
	return clusterServiceVersionMock(ns, name, CRDDescriptionMock())
}

// UnstructuredClusterServiceVersionMock unstructured object based on ClusterServiceVersionMock.
func UnstructuredClusterServiceVersionMock(ns, name string) (*unstructured.Unstructured, error) {
	csv := ClusterServiceVersionMock(ns, name)
	return converter.ToUnstructured(&csv)
}

// ClusterServiceVersionVolumeMountMock based on PostgreSQL operator.
func ClusterServiceVersionVolumeMountMock(ns, name string) olmv1alpha1.ClusterServiceVersion {
	return clusterServiceVersionMock(ns, name, CRDDescriptionVolumeMountMock())
}

// UnstructuredClusterServiceVersionVolumeMountMock returns ClusterServiceVersionVolumeMountMock as
// unstructured object
func UnstructuredClusterServiceVersionVolumeMountMock(
	ns string,
	name string,
) (*unstructured.Unstructured, error) {
	csv := ClusterServiceVersionVolumeMountMock(ns, name)
	return converter.ToUnstructured(&csv)
}

// ClusterServiceVersionListVolumeMountMock returns a list with a single CSV object inside, reusing mock.
func ClusterServiceVersionListVolumeMountMock(ns, name string) *olmv1alpha1.ClusterServiceVersionList {
	return &olmv1alpha1.ClusterServiceVersionList{
		Items: []olmv1alpha1.ClusterServiceVersion{ClusterServiceVersionVolumeMountMock(ns, name)},
	}
}

// DatabaseCRMock based on PostgreSQL operator, returning a instantiated object.
func DatabaseCRMock(ns, name string) *pgv1alpha1.Database {
	return &pgv1alpha1.Database{
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
			DBName:    "test-db",
		},
		Status: pgv1alpha1.DatabaseStatus{
			DBCredentials: "db-credentials",
		},
	}
}

func RouteCRMock(ns, name string) *ocv1.Route {
	return &ocv1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Route",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: ocv1.RouteSpec{
			Host: "https://openshift.cluster.com/host_url",
		},
		Status: ocv1.RouteStatus{},
	}
}

// UnstructuredDatabaseCRMock returns a unstructured version of DatabaseCRMock.
func UnstructuredDatabaseCRMock(ns, name string) (*unstructured.Unstructured, error) {
	db := DatabaseCRMock(ns, name)
	return converter.ToUnstructured(&db)
}

// SecretMock returns a Secret based on PostgreSQL operator usage.
func SecretMock(ns, name string) *corev1.Secret {
	return &corev1.Secret{
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

// ConfigMapMock returns a dummy config-map object.
func ConfigMapMock(ns, name string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string]string{
			"user":     "user",
			"password": "password",
		},
	}
}

// ServiceBindingRequestMock return a binding-request mock of informed name and match labels.
func ServiceBindingRequestMock(
	ns string,
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	matchLabels map[string]string,
	bindUnannotated bool,
) *v1alpha1.ServiceBindingRequest {
	return &v1alpha1.ServiceBindingRequest{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: v1alpha1.ServiceBindingRequestSpec{
			MountPathPrefix: "/var/redhat",
			CustomEnvVar: []v1alpha1.CustomEnvMap{
				{
					Name:  "IMAGE_PATH",
					Value: "spec.imagePath",
				},
			},
			BackingServiceSelector: v1alpha1.BackingServiceSelector{
				Group:       CRDName,
				Version:     CRDVersion,
				Kind:        CRDKind,
				ResourceRef: backingServiceResourceRef,
			},
			ApplicationSelector: v1alpha1.ApplicationSelector{
				Group:       "apps",
				Version:     "v1",
				Resource:    "deployments",
				ResourceRef: applicationResourceRef,
				MatchLabels: matchLabels,
			},
			DetectBindingResources: bindUnannotated,
		},
	}
}

// UnstructuredServiceBindingRequestMock returns a unstructured version of SBR.
func UnstructuredServiceBindingRequestMock(
	ns string,
	name string,
	backingServiceResourceRef string,
	applicationResourceRef string,
	matchLabels map[string]string,
) (*unstructured.Unstructured, error) {
	sbr := ServiceBindingRequestMock(
		ns, name, backingServiceResourceRef, applicationResourceRef, matchLabels, false)
	return converter.ToUnstructuredAsGVK(&sbr, v1alpha1.SchemeGroupVersion.WithKind(OperatorKind))
}

// DeploymentListMock returns a list of DeploymentMock.
func DeploymentListMock(ns, name string, matchLabels map[string]string) appsv1.DeploymentList {
	return appsv1.DeploymentList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentList",
			APIVersion: "apps/v1",
		},
		Items: []appsv1.Deployment{DeploymentMock(ns, name, matchLabels)},
	}
}

// UnstructuredDeploymentMock converts the DeploymentMock to unstructured.
func UnstructuredDeploymentMock(
	ns,
	name string,
	matchLabels map[string]string,
) (*unstructured.Unstructured, error) {
	d := DeploymentMock(ns, name, matchLabels)
	return converter.ToUnstructured(&d)
}

// DeploymentMock creates a mocked Deployment object of busybox.
func DeploymentMock(ns, name string, matchLabels map[string]string) appsv1.Deployment {
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels:    matchLabels,
		},
		Spec: appsv1.DeploymentSpec{
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

//ThirdLevel ...
type ThirdLevel struct {
	Something string `json:"something"`
}

// NestedImage ...
type NestedImage struct {
	Name       string     `json:"name"`
	ThirdLevel ThirdLevel `json:"third"`
}

// NestedDatabaseSpec ...
type NestedDatabaseSpec struct {
	Image NestedImage `json:"image"`
}

// NestedDatabase ...
type NestedDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NestedDatabaseSpec `json:"spec,omitempty"`
}

// NestedDatabaseCRMock based on PostgreSQL operator, returning a instantiated object.
func NestedDatabaseCRMock(ns, name string) NestedDatabase {
	return NestedDatabase{
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
		Spec: NestedDatabaseSpec{
			Image: NestedImage{
				Name: "postgres",
				ThirdLevel: ThirdLevel{
					Something: "somevalue",
				},
			},
		},
	}
}

// UnstructuredNestedDatabaseCRMock returns a unstructured object from NestedDatabaseCRMock.
func UnstructuredNestedDatabaseCRMock(ns, name string) (*unstructured.Unstructured, error) {
	db := NestedDatabaseCRMock(ns, name)
	return converter.ToUnstructured(&db)
}

// ConfigMapDatabaseSpec ...
type ConfigMapDatabaseSpec struct {
	DBConfigMap string `json:"dbConfigMap"`
	ImageName   string
	Image       string
}

// ConfigMapDatabase ...
type ConfigMapDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigMapDatabaseSpec `json:"spec,omitempty"`
}

// DatabaseConfigMapMock returns a local ConfigMapDatabase object.
func DatabaseConfigMapMock(ns, name, configMapName string) *ConfigMapDatabase {
	return &ConfigMapDatabase{
		TypeMeta: metav1.TypeMeta{
			Kind:       CRDKind,
			APIVersion: fmt.Sprintf("%s/%s", CRDName, CRDVersion),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Spec: ConfigMapDatabaseSpec{
			DBConfigMap: configMapName,
			Image:       "docker.io/postgres",
			ImageName:   "postgres",
		},
	}
}

// UnstructuredDatabaseConfigMapMock returns a unstructured version of DatabaseConfigMapMock.
func UnstructuredDatabaseConfigMapMock(ns, name, configMapName string) (*unstructured.Unstructured, error) {
	db := DatabaseConfigMapMock(ns, name, configMapName)
	return converter.ToUnstructured(&db)
}
