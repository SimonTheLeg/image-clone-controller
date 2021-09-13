package controller

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/simontheleg/image-clone-controller/registry"
	corev1 "k8s.io/api/core/v1"
)

type mockCounter struct {
	referenceExistsCalled int
	backUpImageCalled     int
}

type mockImgExistsReg struct {
	mockCounter
}

func (m *mockImgExistsReg) ReferenceExists(ref name.Reference, opts ...remote.Option) (bool, error) {
	m.referenceExistsCalled++
	return true, nil
}
func (m *mockImgExistsReg) BackUpImage(srcRef, destRef name.Reference, srcOpts, destOpts []remote.Option) error {
	m.backUpImageCalled++
	return nil
}

type mockImgNotExistsReg struct {
	mockCounter
}

func (m *mockImgNotExistsReg) ReferenceExists(ref name.Reference, opts ...remote.Option) (bool, error) {
	m.referenceExistsCalled++
	return false, nil
}
func (m *mockImgNotExistsReg) BackUpImage(srcRef, destRef name.Reference, srcOpts, destOpts []remote.Option) error {
	m.backUpImageCalled++
	return nil
}

var _ registry.BackUp = (*mockImgExistsReg)(nil)

func TestEnsureBackUp(t *testing.T) {

	type ts = struct {
		name       string
		mReg       registry.BackUp
		img        string
		newReg     string
		expImg     string
		expErr     error
		expREcalls int
		expBIcalls int
	}

	tc := ts{
		name:       "Image exists in library",
		mReg:       &mockImgExistsReg{},
		img:        "test/library_nginx:latest",
		newReg:     "test",
		expImg:     "index.docker.io/test/library_nginx:latest",
		expErr:     nil,
		expREcalls: 1,
		expBIcalls: 0,
	}

	t.Run(tc.name, func(t *testing.T) {
		b := BackUPer{
			Reg:   tc.mReg,
			DAuth: nil,
		}

		gImg, gErr := b.ensureBackup(context.Background(), tc.img, tc.newReg)

		if gErr != tc.expErr {
			t.Errorf("Err: Want '%s', got '%s'", tc.expErr, gErr)
		}
		if tc.expImg != gImg {
			t.Errorf("Image: Want '%s', got '%s'", tc.expImg, gImg)
		}
		if tc.expREcalls != tc.mReg.(*mockImgExistsReg).referenceExistsCalled {
			t.Errorf("ReferenceExistsCalled: Want '%d', got '%d'", tc.expREcalls, tc.mReg.(*mockImgExistsReg).referenceExistsCalled)
		}
		if tc.expBIcalls != tc.mReg.(*mockImgExistsReg).backUpImageCalled {
			t.Errorf("BackUpImageCalled: Want '%d', got '%d'", tc.expBIcalls, tc.mReg.(*mockImgExistsReg).backUpImageCalled)
		}
	})

	tc = ts{
		name:       "Image does not exist in library",
		mReg:       &mockImgNotExistsReg{},
		img:        "test/library_nginx:latest",
		newReg:     "test",
		expImg:     "index.docker.io/test/library_nginx:latest",
		expErr:     nil,
		expREcalls: 1,
		expBIcalls: 1,
	}

	t.Run(tc.name, func(t *testing.T) {
		b := BackUPer{
			Reg:   tc.mReg,
			DAuth: nil,
		}

		gImg, gErr := b.ensureBackup(context.Background(), tc.img, tc.newReg)

		if gErr != tc.expErr {
			t.Errorf("Err: Want '%s', got '%s'", tc.expErr, gErr)
		}
		if tc.expImg != gImg {
			t.Errorf("Image: Want '%s', got '%s'", tc.expImg, gImg)
		}
		if tc.expREcalls != tc.mReg.(*mockImgNotExistsReg).referenceExistsCalled {
			t.Errorf("ReferenceExistsCalled: Want '%d', got '%d'", tc.expREcalls, tc.mReg.(*mockImgExistsReg).referenceExistsCalled)
		}
		if tc.expBIcalls != tc.mReg.(*mockImgNotExistsReg).backUpImageCalled {
			t.Errorf("BackUpImageCalled: Want '%d', got '%d'", tc.expBIcalls, tc.mReg.(*mockImgExistsReg).backUpImageCalled)
		}
	})

}

func TestPatchPodSpecAndImage(t *testing.T) {
	tt := map[string]struct {
		imgs        []string
		initImgs    []string
		buReg       string
		expImgs     []string
		expInitImgs []string
		expPatch    bool
	}{
		"image to patch, no initContainer": {
			imgs:        []string{"simontheleg/debug-pod:latest"},
			initImgs:    []string{},
			buReg:       "test",
			expImgs:     []string{"index.docker.io/test/simontheleg_debug-pod:latest"},
			expInitImgs: []string{},
			expPatch:    true,
		},
		"nothing to patch": {
			imgs:        []string{"index.docker.io/test/simontheleg_debug-pod:latest"},
			initImgs:    []string{"index.docker.io/test/istio_proxy_init:1.0.2"},
			buReg:       "test",
			expImgs:     []string{"index.docker.io/test/simontheleg_debug-pod:latest"},
			expInitImgs: []string{"index.docker.io/test/istio_proxy_init:1.0.2"},
			expPatch:    false,
		},
		"mix of containers and init containers": {
			imgs:        []string{"simontheleg/debug-pod:latest", "index.docker.io/test/library_nginx:latest"},
			initImgs:    []string{"index.docker.io/test/istio_proxy_init:1.0.2", "quay.io/prometheus/node-exporter:v1.2.2"},
			buReg:       "test",
			expImgs:     []string{"index.docker.io/test/simontheleg_debug-pod:latest", "index.docker.io/test/library_nginx:latest"},
			expInitImgs: []string{"index.docker.io/test/istio_proxy_init:1.0.2", "index.docker.io/test/prometheus_node-exporter:v1.2.2"},
			expPatch:    true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {

			rec := &GenericReconciler{
				RegClient:   &mockImgExistsReg{},
				BuRegRemote: tc.buReg,
			}

			ps := specFromImages(tc.imgs, tc.initImgs)

			gotPatch, gotPts, err := rec.patchPodSpecAndImage(context.Background(), *ps)
			if err != nil {
				t.Fatal("patchPodSpecAndImage should not return an error")
			}

			if tc.expPatch != gotPatch {
				t.Errorf("expPatcH: want '%t', got '%t'", tc.expPatch, gotPatch)
			}

			if len(gotPts.Spec.Containers) != len(tc.expImgs) {
				t.Errorf("Expected %d containers, got %d", len(tc.expImgs), len(gotPts.Spec.Containers))
			}
			if len(gotPts.Spec.InitContainers) != len(tc.expInitImgs) {
				t.Errorf("Expected %d InitContainers, got %d", len(tc.expInitImgs), len(gotPts.Spec.InitContainers))
			}

			for p := range tc.expImgs {
				if gotPts.Spec.Containers[p].Image != tc.expImgs[p] {
					t.Errorf("Containers: Exp image '%s', got '%s'", tc.expImgs[p], gotPts.Spec.Containers[p].Image)
				}
			}
			for p := range tc.expInitImgs {
				if gotPts.Spec.InitContainers[p].Image != tc.expInitImgs[p] {
					t.Errorf("Containers: Exp initImage '%s', got '%s'", tc.expInitImgs[p], gotPts.Spec.InitContainers[p].Image)
				}
			}

		})
	}
}

func specFromImages(images, initImages []string) *corev1.PodTemplateSpec {
	ret := &corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers:     []corev1.Container{},
			InitContainers: []corev1.Container{},
		},
	}

	for _, img := range images {
		ret.Spec.Containers = append(ret.Spec.Containers, corev1.Container{Image: img})
	}

	for _, img := range initImages {
		ret.Spec.InitContainers = append(ret.Spec.InitContainers, corev1.Container{Image: img})
	}

	return ret
}
