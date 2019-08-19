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
	count := 0
	conditionFunc := func() (bool, error) {
		count++
		fmt.Printf("\nRetry count: %+d", count)
		err = client.Get(context.TODO(), nsd, typ)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	if ok, err := conditionFunc(); !ok {
		return err
	}

	err = wait.Poll(time.Second*5, time.Minute*5, conditionFunc)
	return err
}

func WaitUntilResourcesFound(client client.Client, options *client.ListOptions, typ runtime.Object) error {
	var err error
	count := 0
	err = wait.Poll(time.Second*5, time.Minute*5, func() (bool, error) {
		count++
		fmt.Printf("\nRetry count: %+d", count)
		err = client.List(context.TODO(), options, typ)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	return err
}
