package registry

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func TestGenBackUpReference(t *testing.T) {
	tt := map[string]struct {
		reg string
		img string
		exp string
	}{
		"short image": {
			reg: "imageclonebackupregistry/",
			img: "nginx:latest",
			exp: "imageclonebackupregistry/library_nginx:latest",
		},
		"nested image": {
			reg: "imageclonebackupregistry/",
			img: "simontheleg/debug-pod:latest",
			exp: "imageclonebackupregistry/simontheleg_debug-pod:latest",
		},
		"non Dockerhub image": {
			reg: "imageclonebackupregistry/",
			img: "quay.io/prometheus/node-exporter:v1.2.2",
			exp: "imageclonebackupregistry/prometheus_node-exporter:v1.2.2",
		},
		"image already in backup registry": {
			reg: "imageclonebackupregistry/",
			img: "imageclonebackupregistry/simontheleg_debug-pod:latest",
			exp: "imageclonebackupregistry/simontheleg_debug-pod:latest",
		},
		"registry must be escaped": {
			reg: "noslashregistry",
			img: "noslashregistry/simontheleg_debug-pod:latest",
			exp: "noslashregistry/simontheleg_debug-pod:latest",
		},
	}

	for n, tc := range tt {
		t.Run(n, func(t *testing.T) {
			ref, err := name.ParseReference(tc.img)
			if err != nil {
				t.Fatal(err)
			}
			got := GenBackUpReference(tc.reg, ref)

			if got != tc.exp {
				t.Errorf("Exp: '%s', got '%s'", tc.exp, got)
			}
		})
	}

}

// Integration tests begin here
func TestImageExistsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	r := RegistryBackUp{}

	ref, _ := name.ParseReference("imageclonebackupregistry/nginx:latest")
	exists, err := r.ReferenceExists(ref)

	fmt.Printf("Exists: %t\n", exists)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	} else {
		fmt.Println("Error: nil")
	}
}

func TestBackUpImageIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	r := RegistryBackUp{}

	confFile, err := os.Open("/dockerconfig.json")
	if err != nil {
		t.Fatal(err)
	}

	auth, err := AuthFromConfig("dockerhub", confFile)
	if err != nil {
		t.Fatal(err)
	}

	src, _ := name.ParseReference("nginx:1.21.0")
	dest, _ := name.ParseReference("imageclonebackupregistry/nginx:1.21.0")
	err = r.BackUpImage(src, dest, nil, []remote.Option{remote.WithAuth(auth)})

	if err != nil {
		fmt.Println(err)
	}
}

func TestSomething(t *testing.T) {
	ref, _ := name.ParseReference("simontheleg/debug-pod")
	fmt.Println(ref.Name())
}
