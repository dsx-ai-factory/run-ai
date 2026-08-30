package artifactlock

import (
	"fmt"
	"os"
	"sort"

	"github.com/run-ai/runai/build/airgap/pkg/lock"
)

type ImageCollection struct {
	Profile  string
	Platform lock.Platform
	Locks    []lock.CollectedImageLock
}

func CollectImageLocks(paths []string) (ImageCollection, error) {
	if len(paths) == 0 {
		return ImageCollection{}, fmt.Errorf("at least one image lock is required")
	}
	collection := ImageCollection{}
	names := map[string]struct{}{}
	sources := map[string]string{}
	for index, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return ImageCollection{}, fmt.Errorf("read image lock %s: %w", path, err)
		}
		imageLock, err := lock.ParseImageLock(contents)
		if err != nil {
			return ImageCollection{}, fmt.Errorf("parse image lock %s: %w", path, err)
		}
		if index == 0 {
			collection.Profile = imageLock.Spec.Profile
			collection.Platform = imageLock.Spec.Platform
		} else {
			if imageLock.Spec.Profile != collection.Profile {
				return ImageCollection{}, fmt.Errorf("image locks must use the same profile")
			}
			if imageLock.Spec.Platform != collection.Platform {
				return ImageCollection{}, fmt.Errorf("image locks must use the same platform")
			}
		}
		if _, exists := names[imageLock.Metadata.Name]; exists {
			return ImageCollection{}, fmt.Errorf("image lock names must be unique: %s", imageLock.Metadata.Name)
		}
		names[imageLock.Metadata.Name] = struct{}{}
		images := append([]lock.Image(nil), imageLock.Spec.Images...)
		sort.Slice(images, func(i, j int) bool {
			return images[i].Name < images[j].Name
		})
		for _, image := range images {
			if previous, exists := sources[image.Source]; exists && previous != image.Image {
				return ImageCollection{}, fmt.Errorf("conflicting digests for image source %s", image.Source)
			}
			sources[image.Source] = image.Image
		}
		collection.Locks = append(collection.Locks, lock.CollectedImageLock{
			Name:    imageLock.Metadata.Name,
			Version: imageLock.Metadata.Version,
			Digest:  lock.Digest(contents),
			Images:  images,
		})
	}
	sort.Slice(collection.Locks, func(i, j int) bool {
		return collection.Locks[i].Name < collection.Locks[j].Name
	})
	seen := map[string]lock.Image{}
	for index := range collection.Locks {
		images := collection.Locks[index].Images[:0]
		for _, image := range collection.Locks[index].Images {
			key := image.Name + "\x00" + image.Source
			previous, exists := seen[key]
			if exists {
				if previous != image {
					return ImageCollection{}, fmt.Errorf("conflicting lock entries for image %s", image.Name)
				}
				continue
			}
			seen[key] = image
			images = append(images, image)
		}
		collection.Locks[index].Images = images
	}
	return collection, nil
}
