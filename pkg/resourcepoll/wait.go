package resourcepoll

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

func WaitUntilResourceFound(client client.Client, nsd types.NamespacedName, typ runtime.Object) error {
	var err error
	err = wait.Poll(time.Second*5, time.Minute*2, func() (bool, error) {
		err = client.Get(context.TODO(), nsd, typ)
		fmt.Printf("\nType is %+v and error is %+v", typ, err)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return err
}

func WaitUntilResourcesFound(client client.Client, nsd *client.ListOptions, typ runtime.Object) error {
	var err error
	err = wait.Poll(time.Second*5, time.Minute*2, func() (bool, error) {
		err = client.List(context.TODO(), nsd, typ)
		fmt.Printf("\nType is %+v and error is %+v", typ, err)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return err
}
