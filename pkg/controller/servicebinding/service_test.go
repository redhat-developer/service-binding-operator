package servicebinding

import (
	"testing"

	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/redhat-developer/service-binding-operator/pkg/testutils"
	"github.com/redhat-developer/service-binding-operator/test/mocks"
)

func init() {
	logf.SetLogger(logf.ZapLogger(true))
}

func TestFindServiceNewSpecCSV(t *testing.T) {
	ns := "find-cr-tests"
	resourceRef := "db-testing"

	f := mocks.NewFake(t, ns)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	db := f.AddMockedDatabaseCR(resourceRef, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	t.Run("golden path", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(), ns, db.GetObjectKind().GroupVersionKind(), resourceRef)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestFindService(t *testing.T) {
	ns := "find-cr-tests"
	name := "db-testing"

	f := mocks.NewFake(t, ns)

	f.AddMockedUnstructuredCSV("cluster-service-version")
	db := f.AddMockedDatabaseCR(name, ns)
	f.AddMockedUnstructuredDatabaseCRD()

	t.Run("missing service namespace", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(), "", db.GetObjectKind().GroupVersionKind(), name)
		require.Error(t, err)
		require.Equal(t, err, errUnspecifiedBackingServiceNamespace)
		require.Nil(t, cr)
	})

	t.Run("golden path", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(), ns, db.GetObjectKind().GroupVersionKind(), name)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestPlannerWithExplicitBackingServiceNamespace(t *testing.T) {
	ns := "planner"
	backingServiceNamespace := "backing-service-namespace"
	name := "db-testing"

	f := mocks.NewFake(t, ns)

	f.AddMockedUnstructuredDatabaseCRD()
	f.AddMockedUnstructuredCSV("cluster-service-version")
	db := f.AddMockedDatabaseCR(name, backingServiceNamespace)
	f.AddNamespacedMockedSecret("db-credentials", backingServiceNamespace, nil)

	t.Run("findService", func(t *testing.T) {
		cr, err := findService(
			f.FakeDynClient(),
			backingServiceNamespace,
			db.GetObjectKind().GroupVersionKind(),
			name,
		)
		require.NoError(t, err)
		require.NotNil(t, cr)
	})
}

func TestFindServiceCRD(t *testing.T) {
	ns := "planner"
	f := mocks.NewFake(t, ns)
	expected := f.AddMockedUnstructuredDatabaseCRD()
	cr := f.AddMockedDatabaseCR("database", ns)

	t.Run("golden path", func(t *testing.T) {
		crd, err := findServiceCRD(f.FakeDynClient(), cr.GetObjectKind().GroupVersionKind())
		require.NoError(t, err)
		require.NotNil(t, crd)
		require.Equal(t, expected, crd)
	})
}

func TestGetObjectType(t *testing.T) {
	type testCase struct {
		name       string
		descriptors []string
		expected   string
	}

	testCases := []testCase{
		{
			name:       "should build proper annotation",
			descriptors: []string{"urn:alm:descriptor:io.kubernetes:ConfigMap"},
			expected: "ConfigMap",
		},
		{
			name:       "should build proper annotation",
			descriptors: []string{"urn:alm:descriptor:io.kubernetes:Secret"},
			expected: "Secret",
		},
		{
			name:       "should build proper annotation",
			descriptors: []string{"incorrect.annotation:Secret"},
			expected: "",
		},
	}

	for _, args := range testCases {
		t.Run(args.name, func(t *testing.T) {
			objectType := getObjectType(args.descriptors)
			require.Equal(t, args.expected, objectType, "Object type is not matching")
		})
	}
}


func TestLoadDescriptor(t *testing.T) {
	type testCase struct {
		name       string
		path       string
		descriptors []string
		root       string
		expected   map[string]string
	}

	testCases := []testCase{
		{
			name:       "should build proper annotation",
			descriptors: []string{
				"urn:alm:descriptor:io.kubernetes:ConfigMap",
				"service.binding:user:sourceKey=user",
			},
			root:       "status",
			path:       "user",
			expected: map[string]string{
				"service.binding/user": "path={.status.user},sourceKey=user,objectType=ConfigMap",
			},
		},
		{
			name:       "should build proper annotation when object type is not specified",
			descriptors: []string{
				"service.binding",
			},
			root:       "status",
			path:       "user",
			expected: map[string]string{
				"service.binding/user": "path={.status.user}",
			},
		},
		{
			name:       "should build proper annotation",
			descriptors: []string{
				"urn:alm:descriptor:io.kubernetes:Secret",
				"service.binding:user:sourceKey=user",
				"service.binding:password:sourceValue=password",
			},
			root:       "status",
			path:       "dbCredentials",
			expected: map[string]string{
				"service.binding/user": "path={.status.dbCredentials},sourceKey=user,objectType=Secret",
				"service.binding/password": "path={.status.dbCredentials},sourceValue=password,objectType=Secret",
			},
		},
		{
			name:       "should build proper annotation",
			descriptors: []string{
				"service.binding:urls:elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			},
			root:       "status",
			path:       "bootstrap",
			expected: map[string]string{
				"service.binding/urls": "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			},
		},
	}

	for _, args := range testCases {
		t.Run(args.name, func(t *testing.T) {
			anns := map[string]string{}
			objectType := getObjectType(args.descriptors)
			for _, desc := range args.descriptors {
				loadDescriptor(anns, args.path, desc, args.root, objectType)
			}
			require.Equal(t, args.expected, anns)
		})
	}
}

func TestBuildOwnerResourceContext(t *testing.T) {
	ns := "planner"
	f := mocks.NewFake(t, ns)

	obj := f.AddMockedUnstructuredSecret("secret")

	type testCase struct {
		inputPath         string
		outputPath        string
		ownerEnvVarPrefix *string
	}

	testCases := []testCase{
		{
			inputPath:         "data.user",
			outputPath:        "user",
			ownerEnvVarPrefix: nil,
		},
	}

	for _, tc := range testCases {
		got, err := buildOwnedResourceContext(
			f.FakeDynClient(),
			obj,
			tc.ownerEnvVarPrefix,
			testutils.BuildTestRESTMapper(),
			tc.inputPath,
			tc.outputPath,
		)
		require.NoError(t, err)
		require.NotNil(t, got)
	}

}
