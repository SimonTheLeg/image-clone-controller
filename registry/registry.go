package registry

import (
	"io"
	"net/http"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

type BackUp interface {
	ReferenceExists(name.Reference, ...remote.Option) (bool, error)
	BackUpImage(name.Reference, name.Reference, []remote.Option, []remote.Option) error
}

type RegistryBackUp struct{}

var _ BackUp = (*RegistryBackUp)(nil)

// ReferenceExists checks if the specified reference exists in the registry.
// For private registries you can pass credentials as options.
func (*RegistryBackUp) ReferenceExists(ref name.Reference, opts ...remote.Option) (bool, error) {
	_, err := remote.Get(ref, opts...)
	if err != nil {
		if cast, ok := err.(*transport.Error); ok {
			if cast.StatusCode == http.StatusNotFound {
				return false, nil
			} else {
				return false, err
			}
		}
	}
	return true, nil
}

// BackUpImage copies a docker image from one registry to another.
// To check if the destination image already exists, call ReferenceExists first.
func (*RegistryBackUp) BackUpImage(srcRef, destRef name.Reference, srcOpts, destOpts []remote.Option) error {
	img, err := remote.Image(srcRef, srcOpts...)
	if err != nil {
		return err
	}

	err = remote.Write(destRef, img, destOpts...)
	if err != nil {
		return err
	}

	return nil
}

// AuthFromConfig extracts a remote.Option compatible authn.Authenticator from a Docker config.
// It will automatically invoke any key stores or credential helpers if needed
func AuthFromConfig(reg string, conf io.Reader) (authn.Authenticator, error) {
	cf, err := config.LoadFromReader(conf)
	if err != nil {
		return nil, err
	}

	cfg, err := cf.GetAuthConfig(reg)
	if err != nil {
		return nil, err
	}

	return authn.FromConfig(authn.AuthConfig{
		Username:      cfg.Username,
		Password:      cfg.Password,
		Auth:          cfg.Auth,
		IdentityToken: cfg.IdentityToken,
		RegistryToken: cfg.RegistryToken,
	}), nil
}

// GenBackUpReference returns an escaped Reference, which includes the BackupRegistry and is Docker compatible
// It ensures the repo is not multi-nested
func GenBackUpReference(reg string, ref name.Reference) string {
	if reg[len(reg)-1:] != "/" {
		reg += "/"
	}
	escpRepo := strings.Replace(ref.Context().RepositoryStr(), reg, "", 1)
	escpRepo = strings.Replace(escpRepo, "/", "_", -1)
	return reg + escpRepo + ":" + ref.Identifier()
}
