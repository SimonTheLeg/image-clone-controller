package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestDaemonSetController(t *testing.T) {
	tt := map[string]struct {
		name        string
		namespace   string
		imgs        []string
		initImgs    []string
		buReg       string
		expImgs     []string
		expInitImgs []string
	}{
		"should update": {
			name:        "test",
			namespace:   "test",
			imgs:        []string{"simontheleg/debug-pod:latest"},
			initImgs:    []string{},
			buReg:       "test",
			expImgs:     []string{"index.docker.io/test/simontheleg_debug-pod:latest"},
			expInitImgs: []string{},
		},
		"nothing to update": {
			name:        "test",
			namespace:   "test",
			imgs:        []string{"index.docker.io/test/simontheleg_debug-pod:latest"},
			initImgs:    []string{"index.docker.io/test/istio_proxy_init:1.0.2"},
			buReg:       "test",
			expImgs:     []string{"index.docker.io/test/simontheleg_debug-pod:latest"},
			expInitImgs: []string{"index.docker.io/test/istio_proxy_init:1.0.2"},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			ds := dsFromImages(tc.imgs, tc.initImgs, tc.name, tc.namespace)

			c := fake.NewClientBuilder().WithRuntimeObjects(ds).Build()

			rec := &DaemonSetReconciler{
				cl: c,
				GenericReconciler: GenericReconciler{
					RegClient:   &mockImgExistsReg{},
					BuRegRemote: tc.buReg,
				},
			}

			req := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tc.name,
					Namespace: tc.namespace,
				},
			}

			res, err := rec.Reconcile(context.Background(), req)
			if res.Requeue {
				t.Error("reconciliation was requeued when it should not")
			}
			if err != nil {
				t.Errorf("Error: exp nil, got '%s'", err)
			}

			gotDs := &appsv1.DaemonSet{}
			// we can infer if the newly get image matches our expecation, the reconciler has
			// correctly decided whether to update or not
			err = rec.cl.Get(context.Background(), req.NamespacedName, gotDs)
			if err != nil {
				t.Fatalf("could not get DaemonSet: '%v'", err)
			}

			if len(gotDs.Spec.Template.Spec.Containers) != len(tc.expImgs) {
				t.Errorf("Expected %d containers, got %d", len(tc.expImgs), len(gotDs.Spec.Template.Spec.Containers))
			}
			if len(gotDs.Spec.Template.Spec.InitContainers) != len(tc.expInitImgs) {
				t.Errorf("Expected %d InitContainers, got %d", len(tc.expInitImgs), len(gotDs.Spec.Template.Spec.InitContainers))
			}

			for p := range tc.expImgs {
				if gotDs.Spec.Template.Spec.Containers[p].Image != tc.expImgs[p] {
					t.Errorf("Containers: Exp image '%s', got '%s'", tc.expImgs[p], gotDs.Spec.Template.Spec.Containers[p].Image)
				}
			}
			for p := range tc.expInitImgs {
				if gotDs.Spec.Template.Spec.InitContainers[p].Image != tc.expInitImgs[p] {
					t.Errorf("Containers: Exp initImage '%s', got '%s'", tc.expInitImgs[p], gotDs.Spec.Template.Spec.InitContainers[p].Image)
				}
			}

		})
	}
}

func dsFromImages(images []string, initImages []string, name, namespace string) *appsv1.DaemonSet {
	ret := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{},
				},
			},
		},
	}

	for _, img := range images {
		ret.Spec.Template.Spec.Containers = append(ret.Spec.Template.Spec.Containers, corev1.Container{Image: img})
	}

	for _, img := range initImages {
		ret.Spec.Template.Spec.InitContainers = append(ret.Spec.Template.Spec.InitContainers, corev1.Container{Image: img})
	}

	return ret
}
