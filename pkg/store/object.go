package store

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/apiserver/pkg/apis/example/v1"
)

type Object struct {
}

func NewObject() *Object {
	return &Object{}
}

func (o Object) ListNamespacedObjects(gvk string, namespace string) (runtime.Object, error) {
	// TODO: implement properly
	return &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind: "PodList",
		},
		Items: []v1.Pod{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod",
					Namespace:         "ns",
					CreationTimestamp: metav1.Now(),
				},
			},
		},
	}, nil
}

func (o Object) GetNamespacedObject(resource string, nn types.NamespacedName) (runtime.Object, error) {
	// TODO: implement properly
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod",
			Namespace:         "ns",
			CreationTimestamp: metav1.Now(),
		},
	}, nil
}

func (o Object) GetClusterObject(resource string, name string) (runtime.Object, error) {
	// TODO: implement properly
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod",
			Namespace:         "ns",
			CreationTimestamp: metav1.Now(),
		},
	}, nil
}
