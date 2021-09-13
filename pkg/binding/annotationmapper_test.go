package binding

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnnotationBackedBuilderInvalidAnnotation(t *testing.T) {
	type args struct {
		description string
		builder     DefinitionBuilder
	}

	testCases := []args{
		{
			description: "prefix is service.binding but not followed by / or end of string",
			builder: &annotationBackedDefinitionBuilder{
				name: "service.bindingtrololol",
			},
		},
		{
			description: "invalid path",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path=.status.secret",
			},
		},
		{
			description: "invalid path",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path=.status.secret}",
			},
		},
		{
			description: "invalid path",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.secret",
			},
		},
		{
			description: "invalid element type",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/databaseField",
				value: "path={.status.dbCredential},elementType=asdf",
			},
		},
		{
			description: "invalid object type",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/username",
				value: "path={.status.dbCredential},objectType=asdf,valueKey=username",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := tc.builder.Build()
			require.Error(t, err)
		})
	}
}

func TestAnnotationBackedBuilderValidAnnotations(t *testing.T) {
	type args struct {
		description   string
		expectedValue Definition
		builder       DefinitionBuilder
	}

	testCases := []args{
		{
			description: "string definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/username",
				value: "path={.status.dbCredential.username}",
			},
			expectedValue: &stringDefinition{
				outputName: "username",
				definition: definition{
					path: "{.status.dbCredential.username}",
				},
			},
		},
		{
			description: "Ignore non-service binding annotations",
			builder: &annotationBackedDefinitionBuilder{
				name:  "foo",
				value: "bar",
			},
			expectedValue: nil,
		},
		{
			description: "Ignore provisioned service binding annotations",
			builder: &annotationBackedDefinitionBuilder{
				name:  ProvisionedServiceAnnotationKey,
				value: "true",
			},
			expectedValue: nil,
		},
		{
			description: "string definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/anotherUsernameField",
				value: "path={.status.dbCredential.username}",
			},
			expectedValue: &stringDefinition{
				outputName: "anotherUsernameField",
				definition: definition{
					path: "{.status.dbCredential.username}",
				},
			},
		},

		{
			description: "string definition with default username",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.dbCredential.username}",
			},
			expectedValue: &stringDefinition{
				outputName: "",
				definition: definition{
					path: "{.status.dbCredential.username}",
				},
			},
		},

		{
			description: "map from data field definition#Secret#01",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/username",
				value: "path={.status.dbCredential},objectType=Secret,sourceValue=username",
			},
			expectedValue: &mapFromDataFieldDefinition{
				objectType: secretObjectType,
				outputName: "username",
				definition: definition{
					path: "{.status.dbCredential}",
				},
				sourceValue: "username",
			},
		},

		{
			description: "map from data field definition#Secret#02",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/anotherUsernameField",
				value: "path={.status.dbCredential},objectType=Secret,sourceValue=username",
			},
			expectedValue: &mapFromDataFieldDefinition{
				objectType: secretObjectType,
				outputName: "anotherUsernameField",
				definition: definition{
					path: "{.status.dbCredential}",
				},
				sourceValue: "username",
			},
		},

		{
			description: "map from data field definition#Secret#03",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.dbCredential},objectType=Secret",
			},
			expectedValue: &mapFromDataFieldDefinition{
				objectType: secretObjectType,
				outputName: "",
				definition: definition{
					path: "{.status.dbCredential}",
				},
			},
		},

		{
			description: "map from data field definition#ConfigMap",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/username",
				value: "path={.status.dbCredential},objectType=ConfigMap,sourceValue=username",
			},
			expectedValue: &mapFromDataFieldDefinition{
				objectType: configMapObjectType,
				outputName: "username",
				definition: definition{
					path: "{.status.dbCredential}",
				},
				sourceValue: "username",
			},
		},

		{
			description: "map from data field definition#ConfigMap#01",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/anotherUsernameField",
				value: "path={.status.dbCredential},objectType=ConfigMap,sourceValue=username",
			},
			expectedValue: &mapFromDataFieldDefinition{
				objectType: configMapObjectType,
				outputName: "anotherUsernameField",
				definition: definition{
					path: "{.status.dbCredential}",
				},
				sourceValue: "username",
			},
		},

		{
			description: "map from data field definition#ConfigMap#02",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.dbCredential},objectType=ConfigMap,sourceValue=username",
			},
			expectedValue: &mapFromDataFieldDefinition{
				objectType: configMapObjectType,
				definition: definition{
					path: "{.status.dbCredential}",
				},
				sourceValue: "username",
			},
		},

		{
			description: "string of map definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/database",
				value: "path={.status.database},elementType=map",
			},
			expectedValue: &stringOfMapDefinition{
				outputName: "database",
				definition: definition{
					path: "{.status.database}",
				},
			},
		},

		{
			description: "string of map definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/anotherDatabaseField",
				value: "path={.status.database},elementType=map",
			},
			expectedValue: &stringOfMapDefinition{
				outputName: "anotherDatabaseField",
				definition: definition{
					path: "{.status.database}",
				},
			},
		},

		{
			description: "string of map definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.database},elementType=map",
			},
			expectedValue: &stringOfMapDefinition{
				definition: definition{
					path: "{.status.database}",
				},
			},
		},

		{
			description: "slice of maps from path definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			},
			expectedValue: &sliceOfMapsFromPathDefinition{
				definition: definition{
					path: "{.status.bootstrap}",
				},
				sourceKey:   "type",
				sourceValue: "url",
			},
		},

		{
			description: "slice of maps from path definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding/anotherBootstrapField",
				value: "path={.status.bootstrap},elementType=sliceOfMaps,sourceKey=type,sourceValue=url",
			},
			expectedValue: &sliceOfMapsFromPathDefinition{
				outputName: "anotherBootstrapField",
				definition: definition{
					path: "{.status.bootstrap}",
				},
				sourceKey:   "type",
				sourceValue: "url",
			},
		},

		{
			description: "slice of strings from path definition",
			builder: &annotationBackedDefinitionBuilder{
				name:  "service.binding",
				value: "path={.status.bootstrap},elementType=sliceOfStrings,sourceValue=url",
			},
			expectedValue: &sliceOfStringsFromPathDefinition{
				definition: definition{
					path: "{.status.bootstrap}",
				},
				sourceValue: "url",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d, err := tc.builder.Build()
			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, d)
		})
	}
}
