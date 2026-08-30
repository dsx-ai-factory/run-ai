package artifacts

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// ImagePusher uploads the images a bundle carries to a registry.
type ImagePusher interface {
	CheckPushPermission(ctx context.Context, ref string) error
	// Push uploads the image stored in the OCI layout at dir to ref.
	Push(ctx context.Context, dir, ref, expectedDigest string) error
}

type imagePusher struct {
	keychain    authn.Keychain
	transport   http.RoundTripper
	nameOptions []name.Option
}

type credentialHelper struct {
	username string
	password string
}

func (h credentialHelper) Get(string) (string, string, error) {
	return h.username, h.password, nil
}

func NewImagePusher(credential Credential) (ImagePusher, error) {
	transport, err := newTransport(credential)
	if err != nil {
		return nil, err
	}
	var nameOptions []name.Option
	if credential.Insecure {
		nameOptions = append(nameOptions, name.Insecure)
	}
	return &imagePusher{
		keychain:    authn.NewKeychainFromHelper(credentialHelper{username: credential.Username, password: credential.Password}),
		transport:   transport,
		nameOptions: nameOptions,
	}, nil
}

func (p *imagePusher) CheckPushPermission(ctx context.Context, ref string) error {
	target, err := p.parseReference(ref)
	if err != nil {
		return err
	}
	auth, err := authn.Resolve(ctx, p.keychain, target.Context().Registry)
	if err != nil {
		return fmt.Errorf("resolve authorization for %s: %w", ref, err)
	}
	authTransport, err := transport.NewWithContext(
		ctx,
		target.Context().Registry,
		auth,
		p.transport,
		[]string{target.Scope(transport.PushScope)},
	)
	if err != nil {
		return fmt.Errorf("authenticate push to %s: %w", ref, err)
	}
	repository := target.Context()
	url := fmt.Sprintf("%s://%s/v2/%s/blobs/uploads/",
		repository.Registry.Scheme(),
		repository.Registry.RegistryStr(),
		repository.RepositoryStr(),
	)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("create push check for %s: %w", ref, err)
	}
	client := &http.Client{Transport: authTransport}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("check push access to %s: %w", ref, err)
	}
	defer response.Body.Close()
	if err := transport.CheckError(response, http.StatusAccepted); err != nil {
		return fmt.Errorf("check push access to %s: %w", ref, err)
	}
	location, err := response.Location()
	if err != nil {
		return nil
	}
	cancelRequest, err := http.NewRequestWithContext(ctx, http.MethodDelete, location.String(), nil)
	if err != nil {
		return nil
	}
	cancelResponse, err := client.Do(cancelRequest)
	if err == nil {
		cancelResponse.Body.Close()
	}
	return nil
}

// Push uploads a bundled image. A bundled layout holds exactly one image: the
// pack resolves each reference to one platform before writing it.
func (p *imagePusher) Push(ctx context.Context, dir, ref, expectedDigest string) error {
	if expectedDigest == "" {
		return fmt.Errorf("image %s has no expected digest", ref)
	}
	image, err := layoutImage(dir)
	if err != nil {
		return err
	}
	digest, err := image.Digest()
	if err != nil {
		return fmt.Errorf("digest %s: %w", ref, err)
	}
	if digest.String() != expectedDigest {
		return fmt.Errorf("image %s has digest %s but the lock pins %s", ref, digest.String(), expectedDigest)
	}

	target, err := p.parseReference(ref)
	if err != nil {
		return err
	}

	if err := remote.Write(target, image,
		remote.WithContext(ctx),
		remote.WithAuthFromKeychain(p.keychain),
		remote.WithTransport(p.transport),
	); err != nil {
		return fmt.Errorf("push %s: %w", ref, err)
	}
	return nil
}

func (p *imagePusher) parseReference(ref string) (name.Reference, error) {
	parsed, err := name.ParseReference(ref, p.nameOptions...)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference %q: %w", ref, err)
	}
	return parsed, nil
}

func layoutImage(dir string) (v1.Image, error) {
	index, err := layout.ImageIndexFromPath(dir)
	if err != nil {
		return nil, fmt.Errorf("read layout %s: %w", dir, err)
	}
	manifest, err := index.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("read layout %s: %w", dir, err)
	}
	if len(manifest.Manifests) != 1 {
		return nil, fmt.Errorf("layout %s holds %d image(s), expected 1", dir, len(manifest.Manifests))
	}
	image, err := index.Image(manifest.Manifests[0].Digest)
	if err != nil {
		return nil, fmt.Errorf("read layout %s: %w", dir, err)
	}
	return image, nil
}
