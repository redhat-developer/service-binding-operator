module github.com/redhat-developer/service-binding-operator

go 1.14

require (
	cloud.google.com/go v0.45.1 // indirect
	github.com/coreos/etcd-operator v0.9.4
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.4
	github.com/google/go-containerregistry v0.0.0-20191218175032-34fb8ff33bed // indirect
	github.com/imdario/mergo v0.3.8
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/custom-resource-status v0.0.0-20190822192428-e62f2f3b79f3
	github.com/operator-backing-service-samples/postgresql-operator v0.0.0-20191023140509-5c3697ed3069
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200130164400-12c06cfc05c4
	github.com/operator-framework/operator-sdk v0.15.2
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.2
	k8s.io/api v0.17.1
	k8s.io/apiextensions-apiserver v0.17.1
	k8s.io/apimachinery v0.17.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0
	k8s.io/gengo v0.0.0-20191010091904-7fa3014cb28f
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	knative.dev/pkg v0.0.0-20191221032535-9fda5bd59a67 // indirect
	knative.dev/serving v0.9.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.2.4

)

// Pinned to kubernetes-1.16.2
replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible

	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm
	github.com/openshift/api => github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20190923180330-3b6373338c9b
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.0.0-beta.5.0.20200123114618-5e3c7d7eb86a

	// Pin to kube 1.16
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918162238-f783a3654da8
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190918163234-a9c1f33e9fb9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190918163108-da9fdfce26bb
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190918163402-db86a8c7bb21
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190918162944-7a93a0ddadd8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190918162534-de037b596c1e
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190918162820-3b5c1246eb18
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190918164019-21692a0861df
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190918162654-250a1838aa2c
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190918163543-cfa506e53441
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190918162108-227c654b2546
	k8s.io/node-api => k8s.io/node-api v0.0.0-20190918163711-2299658ad911
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190918161442-d4c9c65c82af
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20190918162410-e45c26d066f2
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20190918161628-92eb3cb7496c
)
