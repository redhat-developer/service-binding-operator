package envvars

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {

	type testCase struct {
		name     string
		expected map[string]string
		src      interface{}
		path     []string
	}

	testCases := []testCase{
		{
			name: "should create envvars without prefix",
			expected: map[string]string{
				"status_listeners_0_type":             "secure",
				"status_listeners_0_addresses_0_host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"status_listeners_0_addresses_0_port": "9093",
			},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvars with service prefix",
			expected: map[string]string{
				"kafka_status_listeners_0_type":             "secure",
				"kafka_status_listeners_0_addresses_0_host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"kafka_status_listeners_0_addresses_0_port": "9093",
			},
			path: []string{"kafka"},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvars with binding and service prefixes",
			expected: map[string]string{
				"binding_kafka_status_listeners_0_type":             "secure",
				"binding_kafka_status_listeners_0_addresses_0_host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"binding_kafka_status_listeners_0_addresses_0_port": "9093",
			},
			path: []string{"binding", "kafka"},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvars without prefix",
			expected: map[string]string{
				"status_listeners_0_type":             "secure",
				"status_listeners_0_addresses_0_host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
				"status_listeners_0_addresses_0_port": "9093",
			},
			path: []string{""},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"listeners": []map[string]interface{}{
						{
							"type": "secure",
							"addresses": []map[string]interface{}{
								{
									"host": "my-cluster-kafka-bootstrap.coffeeshop.svc",
									"port": "9093",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "should create envvar for int64 type",
			expected: map[string]string{
				"status_value": "-9223372036",
			},
			src: map[string]interface{}{
				"status": map[string]interface{}{
					"value": int64(-9223372036),
				},
			},
		},
		{
			name: "should create envvar for float64 type",
			expected: map[string]string{
				"": "100.72",
			},
			src: float64(100.72),
		},
		{
			name: "should create envvar for empty string type",
			expected: map[string]string{
				"": "",
			},
			src: "",
		},
		{
			name: "should create envvars for each string in slice",
			expected: map[string]string{
				"tags_0": "knowledge",
				"tags_1": "is",
				"tags_2": "power",
			},
			src: map[string]interface{}{
				"tags": []string{
					"knowledge",
					"is",
					"power",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := Build(tc.src, tc.path...)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
