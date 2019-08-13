package resourcepoll

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type onWait func() error
type onFind func() error

func WaitUntilResourceFound(
	client client.Client,
	nsd types.NamespacedName,
	typ runtime.Object,
	onWait onWait,
	onFind onFind) (err error) {
	count := 0
	err = onWait()
	if err != nil {
		return
	}
	err = wait.Poll(time.Second * 5, time.Minute*5, func() (bool, error) {
		count++
		err = client.Get(context.TODO(), nsd, typ)
		if err != nil {
			return false, err
		}
		err = onFind()
		if err != nil {
			return true, err
		}
		return true, nil
	})
	return
}

func WaitUntilResourcesFound(client client.Client,
	options *client.ListOptions,
	typ runtime.Object,
	onWait onWait,
	onFind onFind) (err error) {
	count := 0
	err = onWait()
	if err != nil {
		return
	}
	err = wait.Poll(time.Second*5, time.Minute*5, func() (bool, error) {
		count++
		err = client.List(context.TODO(), options, typ)
		if err != nil {
			return false, err
		}
		err = onFind()
		if err != nil {
			return true, err
		}
		return true, nil
	})
	return
}
