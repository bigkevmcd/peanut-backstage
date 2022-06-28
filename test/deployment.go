package test

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NewDeployment creates and returns a Deployment resource, with functional opts applied.
func NewDeployment(name, namespace string, opts ...func(runtime.Object)) appsv1.Deployment {
	p := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, o := range opts {
		o(&p)
	}
	return p
}

func WithLabels(m map[string]string) func(runtime.Object) {
	var accessor = meta.NewAccessor()
	return func(obj runtime.Object) {
		accessor.SetLabels(obj, m)
	}
}

func WithAnnotations(m map[string]string) func(runtime.Object) {
	var accessor = meta.NewAccessor()
	return func(obj runtime.Object) {
		accessor.SetAnnotations(obj, m)
	}
}
