package main

import (
	"flag"
	"os"

	"github.com/simontheleg/image-clone-controller/controller"
	"github.com/simontheleg/image-clone-controller/registry"
	kconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type config struct {
	// Kubecontext to use
	context string
	// Namespaces to ignore
	ignNs []string
	// Docker Remote of Backup Registry
	buRegRemote string
	// Location of Docker config
	dockerConfFile string
	// Subconfig to pick in case multiple exist
	dockerConfKey string
}

func defaultConf() *config {
	return &config{
		context:        "",
		ignNs:          []string{"kube-system", "local-path-storage"}, // for the demo to work properly on kind, also ignore local-path-storage
		buRegRemote:    "imageclonebackupregistry/",
		dockerConfFile: "/docker/dockerconfig.json",
		dockerConfKey:  "dockerhub",
	}
}

func main() {
	logf.SetLogger(zap.New())
	var log = logf.Log.WithName("main")

	conf := defaultConf()

	flag.StringVar(&conf.context, "kubecontext", conf.context, "kubernetes context when running locally")
	flag.StringVar(&conf.dockerConfFile, "dockerconf", conf.dockerConfFile, "docker config location")
	flag.StringVar(&conf.buRegRemote, "bureg", conf.buRegRemote, "remote registry to use for backup")
	flag.Parse()

	dConf, err := os.Open(conf.dockerConfFile)
	if err != nil {
		log.Error(err, "could not access dockerconfig")
		os.Exit(1)
	}
	dAuth, err := registry.AuthFromConfig(conf.dockerConfKey, dConf)
	if err != nil {
		log.Error(err, "could not parse dockerconfig")
		os.Exit(1)
	}

	kcfg, err := kconfig.GetConfigWithContext(conf.context)
	if err != nil {
		log.Error(err, "could not obtain kubeconfig")
		os.Exit(1)
	}

	var mgr manager.Manager
	mgr, err = manager.New(kcfg, manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager from kubeconfig")
		os.Exit(1)
	}

	dRec := controller.DeploymentReconciler{
		Igns:        conf.ignNs,
		RegClient:   &registry.RegistryBackUp{},
		BuRegRemote: conf.buRegRemote,
		DAuth:       dAuth,
	}
	err = dRec.SetupWithManager(mgr)
	if err != nil {
		log.Error(err, "could not create Deployments controller")
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "could not start manager")
		os.Exit(1)
	}

}
