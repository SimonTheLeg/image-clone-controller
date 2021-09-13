package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DaemonSetReconciler struct {
	cl client.Client
	GenericReconciler
}

func (r *DaemonSetReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling DaemonSet", "deployment", req.NamespacedName)

	dep := &appsv1.DaemonSet{}
	err := r.cl.Get(ctx, req.NamespacedName, dep)
	if err != nil {
		return reconcile.Result{}, err
	}

	patchReq, upd, err := r.GenericReconciler.patchPodSpecAndImage(ctx, dep.Spec.Template)
	if err != nil {
		return reconcile.Result{}, err
	}

	newDep := dep
	newDep.Spec.Template = *upd

	if patchReq {
		log.Info("Patch required", "deployment", req.NamespacedName)
		err = r.cl.Update(ctx, newDep)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		log.Info("No patch required", "deployment", req.NamespacedName)
	}

	return reconcile.Result{}, nil
}

func (r *DaemonSetReconciler) InjectClient(c client.Client) error {
	r.cl = c
	return nil
}

func (r *DaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		WithEventFilter(contrPredicate(r.Igns)).
		Complete(r)
}
