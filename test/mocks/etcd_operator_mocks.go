package mocks

import (
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func etcdClusterMock(ns, name string) *v1beta2.EtcdCluster {
	return &v1beta2.EtcdCluster{
		TypeMeta: v1.TypeMeta{
			Kind:       "etcd.database.coreos.com/v1beta2",
			APIVersion: "EtcdCluster",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       "1234567",
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
			Kind:       "Secret",
			APIVersion: "",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v12.ServiceSpec{
			ClusterIP: "172.30.0.129",
			Ports: []v12.ServicePort{
				{
					Name:       "tcp-1",
					Protocol:   "TCP",
					Port:       33411,
					TargetPort: intstr.IntOrString{},
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
func CreateEtcdClusterMock(ns, name string) (*v1beta2.EtcdCluster, *v12.Service) {
	ec := etcdClusterMock(ns, name)
	sv := etcdClusterServiceMock(ns, name)
	return ec, sv
}
