/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package airgapvalues

// binding writes the values that point one image at the mirror. Each
// implementation is one way a chart consumes an image reference.
type binding interface {
	apply(v values, registry string, image Image)
}

// registryImage fills the {registry, repository, tag} triple most charts use.
type registryImage struct{ path string }

func (b registryImage) apply(v values, registry string, image Image) {
	v.set(b.path+".registry", registry)
	v.set(b.path+".repository", image.Name)
	v.set(b.path+".tag", image.Tag)
}

// bitnamiImage fills a Bitnami subchart's image and marks the mirror as an
// allowed source, which those charts otherwise refuse to render.
type bitnamiImage struct{ subchart string }

func (b bitnamiImage) apply(v values, registry string, image Image) {
	registryImage{path: b.subchart + ".image"}.apply(v, registry, image)
	v.set(b.subchart+".global.security.allowInsecureImages", true)
}

// fullReferenceImage fills the {repository, tag} pair charts use when they have
// no registry field, where the repository is the whole mirrored path.
type fullReferenceImage struct{ path string }

func (b fullReferenceImage) apply(v values, registry string, image Image) {
	v.set(b.path+".repository", registry+"/"+image.Name)
	v.set(b.path+".tag", image.Tag)
}

// runaiConfigImage fills a {registry, name, tag} triple under clusterConfig,
// reaching the images the cluster operator creates from RunaiConfig rather than
// from Helm.
type runaiConfigImage struct{ path string }

func (b runaiConfigImage) apply(v values, registry string, image Image) {
	v.set("clusterConfig."+b.path+".registry", registry)
	v.set("clusterConfig."+b.path+".name", image.Name)
	v.set("clusterConfig."+b.path+".tag", image.Tag)
}
