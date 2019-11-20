package mocks

import (
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	"github.com/operator-framework/operator-sdk/pkg/test"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func etcdClusterMock(ns, name string) *v1beta2.EtcdCluster {
	return &v1beta2.EtcdCluster{
		TypeMeta: v1.TypeMeta{
			Kind:       v1beta2.EtcdClusterResourceKind,
			APIVersion: v1beta2.SchemeGroupVersion.Version,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1beta2.ClusterSpec{
			Version: "3.2.13",
			Size:    5,
		},
	}
}

func etcdClusterServiceMock(ns, name string) *v12.Service {
	return &v12.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v12.ServiceSpec{
			ClusterIP: "172.30.255.254",
			Ports: []v12.ServicePort{
				{
					Port: 8080,
				},
			},
		},
	}
}

// CreateEtcdClusterMock returns all the resources required to setup an etcd cluster
// using etcd-operator.
// It creates following resources.
// 1. EtcdCluster resource.
// 2. Service(this gets created in etcd reconcile loop).
func CreateEtcdClusterMock(ns, name string, f *test.Framework) (*v1beta2.EtcdCluster, *v12.Service) {
	ec := etcdClusterMock(ns, name)
	sv := etcdClusterServiceMock(ns, name)
	return ec, sv
}
