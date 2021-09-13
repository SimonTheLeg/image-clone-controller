package controller

import (
	"context"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/simontheleg/image-clone-controller/registry"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BackUPer struct {
	Reg   registry.BackUp
	DAuth authn.Authenticator
}

func (b *BackUPer) ensureBackup(ctx context.Context, image string, newReg string) (newImage string, err error) {
	log := log.FromContext(ctx)

	orgRef, err := name.ParseReference(image)
	if err != nil {
		return "", err
	}
	buRef, err := name.ParseReference(registry.GenBackUpReference(newReg, orgRef))
	if err != nil {
		return "", err
	}

	exists, err := b.Reg.ReferenceExists(buRef, remote.WithAuth(b.DAuth))
	if err != nil {
		return "", err
	}
	if exists {
		log.Info("Image already exists in remote. No need to copy", "image", buRef.Context().RepositoryStr(), "remote", buRef.Context().RegistryStr())
	} else {
		log.Info("Creating backup for image", "orig", orgRef.Context().RepositoryStr(), "backup", buRef.Context().RepositoryStr())
		err := b.Reg.BackUpImage(orgRef, buRef, nil, []remote.Option{remote.WithAuth(b.DAuth)})
		if err != nil {
			return "", err
		}
	}

	log.Info("Successfully finished backup", "image", buRef.Context().RepositoryStr(), "remote", buRef.Context().RegistryStr())
	return buRef.Name(), nil
}

type GenericReconciler struct {
	Igns        []string
	RegClient   registry.BackUp
	DAuth       authn.Authenticator
	BuRegRemote string
}

// patchPodSpecAndImage ensures that images are backed up and returns a patched PodTemplateSpec.
// It will leave the old object intact and return a pointer to a patched copy
func (r *GenericReconciler) patchPodSpecAndImage(ctx context.Context, old corev1.PodTemplateSpec) (patchReq bool, upd *corev1.PodTemplateSpec, err error) {
	bu := BackUPer{
		Reg:   r.RegClient,
		DAuth: r.DAuth,
	}

	upd = &old
	for p, cont := range old.Spec.InitContainers {
		ref, err := bu.ensureBackup(ctx, cont.Image, r.BuRegRemote)
		if old.Spec.InitContainers[p].Image != ref {
			patchReq = true
			upd.Spec.InitContainers[p].Image = ref
		}
		if err != nil {
			return false, nil, err
		}
	}

	for p, cont := range old.Spec.Containers {
		ref, err := bu.ensureBackup(ctx, cont.Image, r.BuRegRemote)
		if old.Spec.Containers[p].Image != ref {
			patchReq = true
			upd.Spec.Containers[p].Image = ref
		}
		if err != nil {
			return false, nil, err
		}
	}
	return patchReq, upd, nil
}
