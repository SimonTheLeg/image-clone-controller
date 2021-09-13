package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DeploymentReconciler struct {
	cl client.Client
	GenericReconciler
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling Deployment", "deployment", req.NamespacedName)

	dep := &appsv1.Deployment{}
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

func (r *DeploymentReconciler) InjectClient(c client.Client) error {
	r.cl = c
	return nil
}

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(contrPredicate(r.Igns)).
		Complete(r)
}
