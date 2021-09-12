package controller

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/simontheleg/image-clone-controller/registry"
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
