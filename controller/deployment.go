package controller

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/simontheleg/image-clone-controller/registry"
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type DeploymentReconciler struct {
	cl client.Client
	// Namespaces to ignore
	Igns []string
	// Docker Authentication
	RegClient   registry.BackUp
	DAuth       authn.Authenticator
	BuRegRemote string
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)

	log.Info("Reconciling Deployment", "deployment", req.NamespacedName)

	dep := &appsv1.Deployment{}
	err := r.cl.Get(ctx, req.NamespacedName, dep)
	if err != nil {
		return reconcile.Result{}, err
	}

	newDep := dep
	patchReq := false
	bu := BackUPer{
		Reg:   r.RegClient,
		DAuth: r.DAuth,
	}
	for p, cont := range dep.Spec.Template.Spec.InitContainers {
		ref, err := bu.ensureBackup(ctx, cont.Image, r.BuRegRemote)
		if dep.Spec.Template.Spec.InitContainers[p].Image != ref {
			patchReq = true
			newDep.Spec.Template.Spec.InitContainers[p].Image = ref
		}
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	for p, cont := range dep.Spec.Template.Spec.Containers {
		ref, err := bu.ensureBackup(ctx, cont.Image, r.BuRegRemote)
		if dep.Spec.Template.Spec.Containers[p].Image != ref {
			patchReq = true
			newDep.Spec.Template.Spec.Containers[p].Image = ref
		}
		if err != nil {
			return reconcile.Result{}, err
		}
	}

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
