package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"
	orasremote "oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	mediaTypeFile = "application/vnd.selebrow.file"
	artifactType  = "application/vnd.selebrow.artifact"

	ociFile = "org.opencontainers.image.title"
)

type Registry struct {
	repository *orasremote.Repository
}

func (r *Registry) CheckImageExists(ctx context.Context, tag string) (bool, error) {
	_, err := r.repository.Resolve(ctx, fmt.Sprintf("%s:%s", r.repository.Reference.String(), tag))
	if err != nil {
		if errors.Is(err, errdef.ErrNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *Registry) UploadFiles(tag string, files ...string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	fs, err := file.New(wd)
	if err != nil {
		return err
	}
	defer fs.Close()

	ctx := context.Background()

	fileDescriptors := make([]v1.Descriptor, 0, len(files))
	for _, f := range files {
		descriptor, err := fs.Add(ctx, f, mediaTypeFile, "")
		if err != nil {
			return err
		}

		fileDescriptors = append(fileDescriptors, descriptor)
	}

	manifestDescriptor, err := oras.PackManifest(
		ctx,
		fs,
		oras.PackManifestVersion1_1,
		artifactType,
		oras.PackManifestOptions{
			Layers: fileDescriptors,
		},
	)

	if err := fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		return err
	}

	_, err = oras.Copy(ctx, fs, tag, r.repository, tag, oras.DefaultCopyOptions)
	return err
}

func (r *Registry) GetDownloadLinks(tag string, files map[string]struct{}) (map[string]string, error) {
	_, content, err := oras.FetchBytes(context.Background(), r.repository, tag, oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, err
	}

	var manifest v1.Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, err
	}

	scheme := "https"
	if r.repository.PlainHTTP {
		scheme = "http"
	}

	baseURL := fmt.Sprintf(
		"%s://%s/v2/%s/blobs/",
		scheme,
		r.repository.Reference.Registry,
		r.repository.Reference.Repository,
	)

	results := make(map[string]string)
	for _, layer := range manifest.Layers {
		file := layer.Annotations[ociFile]
		_, ok := files[file]
		if !ok {
			continue
		}

		results[file] = baseURL + layer.Digest.String()
	}

	return results, nil
}

func (r *Registry) TagImage(src, dst string) error {
	ctx := context.Background()
	ref, err := r.repository.Resolve(ctx, src)
	if err != nil {
		return err
	}

	return r.repository.Tag(ctx, ref, dst)
}

func InitRegistry(insecure bool, registry, repo, image string) (*Registry, error) {
	remoterepo, err := orasremote.NewRepository(fmt.Sprintf("%s/%s/%s", registry, repo, image))
	if err != nil {
		return nil, err
	}

	var credential auth.Credential

	token := os.Getenv("GH_TOKEN")
	user := os.Getenv("GH_USER")

	if token != "" {
		credential.Password = token
		credential.Username = user
	}

	remoterepo.PlainHTTP = insecure

	remoterepo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: auth.StaticCredential(registry, credential),
	}

	return &Registry{repository: remoterepo}, nil
}
