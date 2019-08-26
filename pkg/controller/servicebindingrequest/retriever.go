package servicebindingrequest

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/plan"

)

// Retriever reads all data referred in plan instance, and store in a secret.
type Retriever struct {
	ctx           context.Context // request context
	client        client.Client   // Kubernetes API client
	collectors    []Collector
	plan          *plan.Plan             // plan instance
	logger        logr.Logger       // logger instance
	data          map[string][]byte // data retrieved
	volumeKeys    []string
	bindingPrefix string
}

const (
	basePrefix              = "binding:env:object"
	secretPrefix            = basePrefix + ":secret"
	configMapPrefix         = basePrefix + ":configmap"
	attributePrefix         = "binding:env:attribute"
	volumeMountSecretPrefix = "binding:volumemount:secret"
)

// RegisterCollector registers a collector.
func (r *Retriever) RegisterCollector(c Collector) {
	r.collectors = append(r.collectors, c)
}

// saveDataOnSecret create or update secret that will store the data collected.
func (r *Retriever) saveDataOnSecret() error {
	secretObj := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.plan.Name,
			Namespace: r.plan.Ns,
		},
		Data: r.data,
	}

	err := r.client.Create(r.ctx, secretObj)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return r.client.Update(r.ctx, secretObj)
}

// Retrieve loop and read data pointed by the references in plan instance.
func (r *Retriever) Retrieve() error {

	for _, c := range r.collectors {
		c.Collect()
		// c.Collect() returns BindableMetadata, use it to populate data structures in Retriever.
	}

	return r.saveDataOnSecret()
}

// NewRetriever instantiate a new retriever instance.
func NewRetriever(ctx context.Context, client client.Client, plan *plan.Plan, bindingPrefix string) *Retriever {
	return &Retriever{
		ctx:           ctx,
		client:        client,
		logger:        logf.Log.WithName("retriever"),
		plan:          plan,
		data:          make(map[string][]byte),
		volumeKeys:    []string{},
		bindingPrefix: bindingPrefix,
	}
}
