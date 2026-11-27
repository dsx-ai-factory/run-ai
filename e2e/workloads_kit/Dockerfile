# syntax=docker/dockerfile:1.7
# SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Every test this kit runs (--grep @k8s-flav-workloads-kit) is API-only: no
# `page` fixture, no chromium.launch(), no BrowserEngine. Playwright is used
# purely as a test runner, which needs no browser binaries — so this stays on a
# plain Node base instead of mcr.microsoft.com/playwright (~1.2GB of chromium
# that is never executed and carries most of that image's CVE findings).
#
# Node.js itself is the official nodejs.org build, copied out of the node image
# onto UBI, rather than Red Hat's ubi9/nodejs-24-minimal image.
# The RPM route installs the nodejs:24 module RPMs (nodejs, nodejs-libs,
# nodejs-full-i18n, npm), and vulnerability scanners match Red Hat's
# component-level "affected / no fix planned" entries for that module by RPM
# name alone — e.g. CVE-2025-64756 (glob CLI injection) is reported HIGH
# against all four RPMs even though the npm they ship bundles glob 13.x, which
# was never affected. With no module RPMs in the rpmdb that matcher has
# nothing to hit; the official build ships node plus npm >= 11.16 (bundled
# glob ^13) and full ICU, so nothing functional is lost.
#
# If a UI test is ever tagged into this kit, none of the above holds — Playwright
# publishes no browser builds for UBI, so that would mean returning to
# mcr.microsoft.com/playwright:<exact @playwright/test version>-jammy as the
# final base.
#
# The final image is a ubi9/ubi-micro root filesystem with only the packages
# start.sh needs, installed with `dnf --installroot` from a full ubi9 stage. It
# has no package manager: on ubi-minimal, microdnf alone pulled glib2, libdnf,
# librepo and gnupg2 into the rpmdb, and through them libblkid, libmount,
# libsmartcols, libuuid and gnutls -- packages with open HIGH-severity
# findings, none of them used by either kit. The rpmdb is kept, so scanners
# still see every RPM that ships (unlike a FROM scratch copy, which only hides
# them).
#
# The final image ships no npm either: npm's own bundled node_modules were the
# source of a rolling set of npm-ecosystem CVE rows that the suite's
# package.json cannot override. Everything npm is needed for -- npm ci, the
# global allure-commandline install -- happens in throwaway stages, and
# kit-common/kit-utils.sh runs ./node_modules/.bin/playwright rather than npx.
#
# artifact-metadata.json sets skip_container_validation, as control-plane/
# frontend does, so the policies in common/opa-policies/dockerfile are not run
# against this file. That is what lets the node image below serve as a
# COPY --from source, kubectl and helm build in golang stages, and stages chain
# via `FROM base`. Nothing enforces the final image's base any more: keep the
# final and rootfs stages on ubi9/ubi-micro.
#
# The suite comes from the build context. This file lives in the repository it
# packages, so the image carries the e2e/ tree of the checkout it is built
# from: the tests, workloads_kit/start.sh and kit-common/ all come from one
# commit. The build context must be e2e/ (the directory that holds
# package.json), and e2e/.dockerignore is what keeps local leftovers --
# node_modules, reports, logs, env files, saved tokens, kit tarballs -- out of
# it. Build with `make build` from workloads_kit/, or from e2e/ with:
#   docker buildx build -f workloads_kit/Dockerfile .
# Untracked files that .dockerignore does not name are copied too, so build
# what you ship from a clean checkout of that commit.
#
# compatibility_kit/Dockerfile is built the same way; keep the two in step.

# The Go toolchain and the golang.org/x modules that kubectl and helm are both
# built with (see their stages). One set of pins, so the two binaries cannot
# drift apart. go1.26.11 is past the fixed versions for the go1.26.9 and
# go1.26.10 standard library advisories (CVE-2026-97033 to CVE-2026-97039 and
# CVE-2026-97043); x/net 0.63.0 is past GO-2026-6860 and GO-2026-6866 (both
# fixed in 0.62.0) and needs x/text >= 0.44.0.
ARG GO_VERSION=1.26.11
ARG X_TEXT_VERSION=0.44.0
ARG X_NET_VERSION=0.63.0

# ---------------------------------------------------------------------------
# Upstream tool image. A COPY --from source only -- nothing runs in it and
# none of its packages reach the final image.
# ---------------------------------------------------------------------------
# The official image installs the nodejs.org release tarball. Its npm 11.19.1
# carries glob ^13, outside CVE-2025-64756's vulnerable ranges. 24.22.3 is
# past the fixed version for the 24.22.2 advisories (CVE-2026-123863 to
# CVE-2026-123866, CVE-2026-123872, CVE-2026-123874 and CVE-2026-126051).
FROM node:24.22.3-bookworm-slim AS node

# ---------------------------------------------------------------------------
# kubectl: built from source rather than copied from registry.k8s.io/kubectl.
# Throwaway stage -- only /out/kubectl is promoted.
# ---------------------------------------------------------------------------
# The official v1.37.1 binary is built with go1.26.8 and golang.org/x/text
# v0.40.0 / x/net v0.57.0, which are affected by GO-2026-6629 (x/text, HIGH;
# fixed in 0.41.0), by GO-2026-6860 (x/net, HIGH; fixed in 0.62.0) and by the
# standard library advisories named above.
# kubectl upstream is a 10-line main over published modules, so
# this compiles that exact main.go (kubernetes/cmd/kubectl/kubectl.go at
# v1.37.1) against k8s.io/kubectl, component-base and client-go v0.37.1 with
# only the golang.org/x modules raised to the pins above, and whatever other
# golang.org/x modules those pins require. Every other module matches the
# official binary.
#
# Cross-compiled on the build platform (CGO off, like upstream), so an
# emulated arm64 build does not run the Go compiler under QEMU. Bump
# KUBECTL_VERSION and the x/ pins together; `go get` fails the build if a pin
# is lower than what the k8s modules require.
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS kubectl
ARG TARGETOS
ARG TARGETARCH
ARG X_TEXT_VERSION
ARG X_NET_VERSION
ARG KUBECTL_VERSION=1.37.1
WORKDIR /src
COPY <<'EOF' main.go
/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	"k8s.io/component-base/cli"
	"k8s.io/component-base/logs"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/util"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	// We need to manually parse the arguments looking for verbosity flag and
	// set appropriate level here, because in the normal flow the flag parsing,
	// including the logging verbosity, happens inside cli.RunNoErrOutput.
	// Doing it here ensures we can continue using klog during kubectl command
	// construction, which includes handling plugins and parsing .kuberc file,
	// for example.
	logs.GlogSetter(cmd.GetLogVerbosity(os.Args)) // nolint:errcheck
	command := cmd.NewDefaultKubectlCommand()
	if err := cli.RunNoErrOutput(command); err != nil {
		// Pretty-print the error and exit with an error.
		util.CheckErr(err)
	}
}
EOF
RUN M="v0.${KUBECTL_VERSION#1.}" \
    && go mod init kubectl \
    && go get k8s.io/kubectl@$M k8s.io/component-base@$M k8s.io/client-go@$M \
              golang.org/x/text@v${X_TEXT_VERSION} golang.org/x/net@v${X_NET_VERSION} \
    && go mod tidy
# Version strings as the upstream release build stamps them, so
# `kubectl version --client` still reports v1.37.1.
RUN V="v${KUBECTL_VERSION}" && MINOR="$(echo "$KUBECTL_VERSION" | cut -d. -f2)" && \
    LD="-s -w" && \
    for p in k8s.io/component-base/version k8s.io/client-go/pkg/version; do \
      LD="$LD -X $p.gitVersion=$V -X $p.gitMajor=1 -X $p.gitMinor=$MINOR -X $p.gitTreeState=clean"; \
    done && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -trimpath -ldflags "$LD" -o /out/kubectl . \
    && go version -m /out/kubectl > /tmp/buildinfo \
    && grep -E "golang.org/x/text[[:space:]]+v${X_TEXT_VERSION}([[:space:]]|$)" /tmp/buildinfo \
    && grep -E "golang.org/x/net[[:space:]]+v${X_NET_VERSION}([[:space:]]|$)" /tmp/buildinfo

# ---------------------------------------------------------------------------
# helm: built from source rather than downloaded from get.helm.sh.
# Throwaway stage -- only /out/helm is promoted.
# ---------------------------------------------------------------------------
# 3.22.1 is the newest Helm 3 release. Its official binary is built with
# go1.26.9 and golang.org/x/net v0.60.0, which are affected by GO-2026-6860
# (x/net, HIGH; fixed in 0.62.0) and by the standard library advisories named
# above, and no Helm 3 release carries the fixes. So this compiles the v3.22.1
# source -- the helm.sh/helm/v3 module, which `go mod download` checks against
# the Go checksum database -- with the same toolchain and golang.org/x pins as
# kubectl. The pins raise the other golang.org/x modules they depend on with
# them (x/crypto to 0.59.0, x/sync, x/sys and x/term); every other module is
# the one the release's go.mod selects, including oras-go 2.6.2 (past
# GHSA-fxhp-mv3v-67qp, the tar hardlink escape).
#
# The version strings are the ones the release build stamps, so `helm version`
# still reports v3.22.1 and its commit. HELM_COMMIT is the commit the v3.22.1
# tag points at: the build fails if the module proxy recorded a different one
# for that version. Bump HELM_VERSION and HELM_COMMIT together. Cross-compiled
# on the build platform, like kubectl.
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS helm
ARG TARGETOS
ARG TARGETARCH
ARG X_TEXT_VERSION
ARG X_NET_VERSION
ARG HELM_VERSION=3.22.1
ARG HELM_COMMIT=d6c1a0e4027caa6abc7892e7ea9f0cda7447baa3
WORKDIR /src
RUN go mod download "helm.sh/helm/v3@v${HELM_VERSION}" \
    && grep -q "\"Hash\":\"${HELM_COMMIT}\"" \
         "$(go env GOMODCACHE)/cache/download/helm.sh/helm/v3/@v/v${HELM_VERSION}.info" \
    && cp -r "$(go env GOMODCACHE)/helm.sh/helm/v3@v${HELM_VERSION}/." . \
    && chmod -R u+w . \
    && go get golang.org/x/text@v${X_TEXT_VERSION} golang.org/x/net@v${X_NET_VERSION}
# LDFLAGS as helm's Makefile sets them for a release: the version, the commit,
# and the Kubernetes minor of the client-go module helm is built against.
RUN P=helm.sh/helm/v3 && \
    K8S_MINOR="$(go list -f '{{.Version}}' -m k8s.io/client-go | cut -d. -f2)" && \
    LD="-s -w -X $P/internal/version.version=v${HELM_VERSION} -X $P/internal/version.metadata=" && \
    LD="$LD -X $P/internal/version.gitCommit=${HELM_COMMIT} -X $P/internal/version.gitTreeState=clean" && \
    for p in pkg/lint/rules pkg/chartutil; do \
      LD="$LD -X $P/$p.k8sVersionMajor=1 -X $P/$p.k8sVersionMinor=$K8S_MINOR"; \
    done && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
      go build -trimpath -ldflags "$LD" -o /out/helm ./cmd/helm \
    && go version -m /out/helm > /tmp/buildinfo \
    && grep -E "golang.org/x/text[[:space:]]+v${X_TEXT_VERSION}([[:space:]]|$)" /tmp/buildinfo \
    && grep -E "golang.org/x/net[[:space:]]+v${X_NET_VERSION}([[:space:]]|$)" /tmp/buildinfo

# ---------------------------------------------------------------------------
# base: ubi-minimal plus node and npm under /usr/local -- the build stages'
# toolchain. deps builds on it, and the final image copies the same node
# binary, so node-pty's native addon (built in deps, tied to that binary's
# ABI) runs against exactly what compiled it.
# ---------------------------------------------------------------------------
FROM registry.access.redhat.com/ubi9/ubi-minimal AS base

# Only node, npm and the headers node-gyp builds against -- not the image's
# yarn or corepack.
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/npm
COPY --from=node /usr/local/include/node /usr/local/include/node
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx && \
    node --version && npm --version

# ---------------------------------------------------------------------------
# deps-js: install the suite's node_modules on the build platform. Throwaway
# stage -- /app/e2e is promoted into deps below.
# ---------------------------------------------------------------------------
# npm is JavaScript running JavaScript: under QEMU, as when an arm64 image is
# built on an amd64 host, `npm ci` takes several times as long as it does
# natively and comes to dominate the build.
# Everything npm installs is platform-independent files except two things:
# native addons compiled by node-gyp, and optional dependencies with os/cpu
# fields, which npm selects for the target with --os and --cpu. So, like the
# kubectl stage, this runs on $BUILDPLATFORM with install scripts off, and the
# suite's one native addon (node-pty, see deps) is rebuilt for the target below.
#
# The lockfile's only os/cpu-gated package is fsevents (darwin, optional), so
# --cpu changes nothing today; it is here so a future optional dependency with
# per-arch binaries lands as the arm64 variant in the arm64 image. npm names
# architectures like process.arch (x64, arm64), TARGETARCH like Go (amd64).
#
# The image is named here rather than `FROM --platform=$BUILDPLATFORM base`:
# BuildKit honours --platform on an image reference but builds that form for
# the target platform, so npm would still run under QEMU. Keep the tag in step
# with the `node` stage above: the npm that writes node_modules should be the
# one the kit's node binary ships.
FROM --platform=$BUILDPLATFORM node:24.22.3-bookworm-slim AS deps-js
ARG TARGETARCH

# Keep browser binaries out of the image: this kit runs API-only tests, and
# chromium would bring most of a browser image's CVE findings with it.
# Redundant with --ignore-scripts, kept so the intent survives if that flag
# ever goes.
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

# The manifests go in first, so the install layer is rebuilt when the
# dependencies change rather than on every edit to a test.
WORKDIR /app/e2e
COPY package.json package-lock.json ./
RUN case "$TARGETARCH" in amd64) CPU=x64 ;; *) CPU="$TARGETARCH" ;; esac && \
    npm ci --ignore-scripts --os=linux --cpu="$CPU" --no-audit --no-fund

# The suite itself, from the build context (e2e/, filtered by
# e2e/.dockerignore). That file excludes node_modules, so this cannot
# overwrite the install above with a host's copy.
COPY . .

# ---------------------------------------------------------------------------
# deps: compile node-pty for the target platform. Throwaway stage -- only
# /app/e2e is promoted into the final image.
# ---------------------------------------------------------------------------
# node-pty is the suite's only native module and publishes no Linux prebuilds,
# so it is compiled with node-gyp, which needs python3, make and gcc-c++.
# Nothing in the kit needs that toolchain afterwards, so it stays in this stage:
# shipping it put python3, python3-libs, python-unversioned-command and
# python3-pip-wheel -- and their open CVE rows -- in the final image. node-pty
# itself has to ship: src/modules/cli/index.ts re-exports RunaiCliV2, which
# imports it at module load, and every test file and the kit's global setup
# import that barrel, so Playwright cannot even collect tests without it.
# This is the only part of the install that still runs under emulation in an
# arm64 build: one small C++ addon, built against the node binary the final image
# ships (see base). The require() proves the addon loads before it is promoted.
FROM base AS deps

RUN microdnf install -y --nodocs \
    python3 \
    make \
    gcc-c++ \
    && microdnf clean all \
    && rm -rf /var/cache/yum

COPY --from=deps-js /app/e2e /app/e2e
WORKDIR /app/e2e
RUN npm rebuild node-pty && \
    node -e "require('node-pty')"

# ---------------------------------------------------------------------------
# tools: the Temurin JRE and allure-commandline, prepared where curl, tar and
# npm exist so the final image needs none of them. Throwaway stage -- only
# /opt/java/openjdk and the allure-commandline package are promoted.
# ---------------------------------------------------------------------------
FROM base AS tools

# curl is already in ubi-minimal; tar and gzip unpack the JRE, findutils drives
# the jackson swap.
RUN microdnf install -y --nodocs tar gzip findutils \
    && microdnf clean all \
    && rm -rf /var/cache/yum

# Java for allure-commandline's report generation. The JRE RPM would pull
# python3, python3-libs, python3-pip-wheel and expat into the rpmdb through nss
# -> crypto-policies-scripts, with open CVE rows and no runtime use here; the
# Temurin build is self-contained (its own cacerts, no rpm dependencies).
#
# This is the Eclipse Temurin release archive, pinned by version and per-arch
# SHA-256: the same files and sums the eclipse-temurin image's own Dockerfile
# installs. It is downloaded rather than copied out of
# eclipse-temurin:17.0.21.1_2-jre-ubi9-minimal because that tag was published
# without a linux/arm64 image. 17.0.21.1+2 is past the fixed versions for the
# 17.0.20.1 advisories (CVE-2026-91415, fixed in 17.0.21; CVE-2026-114938,
# CVE-2026-114939, CVE-2026-114940 and CVE-2026-114942, fixed in 17.0.21.1).
ARG TEMURIN_VERSION=17.0.21.1_2
RUN ARCH="$(uname -m)" && \
    case "$ARCH" in \
      x86_64) ARCH=x64; SHA256=5a092a0bc21de4c3d855480edd7945be9cfe157b38cdb84dac280511893bb009 ;; \
      aarch64) SHA256=ed968eb2cde06ba38c3d957956cab7faced9181aaeebed4904f89e2e6efa8d11 ;; \
      *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;; \
    esac && \
    RELEASE="jdk-$(echo "$TEMURIN_VERSION" | sed 's/_/%2B/')" && \
    curl -fsSLo /tmp/jre.tar.gz \
      "https://github.com/adoptium/temurin17-binaries/releases/download/${RELEASE}/OpenJDK17U-jre_${ARCH}_linux_hotspot_${TEMURIN_VERSION}.tar.gz" && \
    printf '%s  %s\n' "$SHA256" /tmp/jre.tar.gz | sha256sum --check - && \
    mkdir -p /opt/java/openjdk && \
    tar -xzf /tmp/jre.tar.gz -C /opt/java/openjdk --strip-components=1 --no-same-owner && \
    rm -f /tmp/jre.tar.gz && \
    /opt/java/openjdk/bin/java -version

# Allure commandline for report generation. Only its package directory ships;
# the final image runs it with node and the Temurin JRE.
RUN npm install -g allure-commandline

# allure-commandline (2.46.1 and earlier) bundles jackson-databind and
# jackson-core 2.22.2 in dist/lib and in the jira/xray plugins' lib dirs, which
# carry HIGH advisories (databind: GHSA-cxp5-3px4-pw24, GHSA-wv8q-qhhj-9h54; core:
# GHSA-p6pp-m3f8-5c89, GHSA-7hhh-6rmp-j9qf; all fixed in 2.22.3). Swap every
# copy for 2.22.3 from Maven Central, pinned by SHA-256. It is a patch release
# on the same 2.22 line as the bundled jackson-annotations. The launcher lists
# the jars by name in its CLASSPATH, so rename them there too; the plugins load
# whatever is in their lib/ dir.
ARG JACKSON_VERSION=2.22.3
RUN set -eu; \
    ALLURE_DIST=/usr/local/lib/node_modules/allure-commandline/dist; \
    for spec in \
      jackson-databind:556db5439e206114346043f68d200497dc96a0bca62a360a81784092ebd0e0a9 \
      jackson-core:8a501126a385b25841915d839508f8a66e2a0dbc8a6709d055ef3b3e852b094c; do \
      artifact="${spec%%:*}"; sha256="${spec#*:}"; \
      curl -fsSLo "/tmp/${artifact}.jar" \
        "https://repo1.maven.org/maven2/com/fasterxml/jackson/core/${artifact}/${JACKSON_VERSION}/${artifact}-${JACKSON_VERSION}.jar"; \
      printf '%s  %s\n' "$sha256" "/tmp/${artifact}.jar" | sha256sum --check -; \
      for old in $(find "$ALLURE_DIST" -name "${artifact}-*.jar"); do \
        rm -f "$old"; \
        cp "/tmp/${artifact}.jar" "$(dirname "$old")/${artifact}-${JACKSON_VERSION}.jar"; \
      done; \
      sed -i -E "s#${artifact}-[0-9.]+\.jar#${artifact}-${JACKSON_VERSION}.jar#g" \
        "$ALLURE_DIST/bin/allure"; \
      rm -f "/tmp/${artifact}.jar"; \
      test "$(find "$ALLURE_DIST" -name "${artifact}-*.jar" | wc -l)" -eq 3; \
      ! find "$ALLURE_DIST" -name "${artifact}-*.jar" ! -name "${artifact}-${JACKSON_VERSION}.jar" | grep -q .; \
      grep -q "${artifact}-${JACKSON_VERSION}.jar" "$ALLURE_DIST/bin/allure"; \
    done

# ---------------------------------------------------------------------------
# rootfs: ubi9/ubi-micro plus the packages the kit needs at runtime, installed
# with `dnf --installroot` from a full ubi9 stage, so no package manager reaches
# the final image (see the header comment). Throwaway stage -- only /rootfs is
# promoted.
# ---------------------------------------------------------------------------
FROM registry.access.redhat.com/ubi9/ubi-micro AS micro
FROM registry.access.redhat.com/ubi9/ubi AS rootfs

COPY --from=micro / /rootfs

# ubi-micro already has bash and coreutils-single (base64, date, stat, stdbuf,
# tee, tr, wc, ...). On top of it: what start.sh, kit-common/kit-utils.sh and
# the allure launcher call (find, grep, sed, tar, zip, unzip, gzip), libstdc++
# for node, and ca-certificates for TLS. curl-minimal (UBI9's /usr/bin/curl) is
# required: the kit's global setup (kit-shared-global-setup.ts ->
# RunaiCliV2.downloadAndInstall) fetches the runai CLI installer with
# `curl -fsSL .../cli/installer/unix`, and that installer checks for curl and
# downloads the CLI binary with it. Without curl the setup only warns, and
# every CLI-driven test fails with ".runai/bin/runai: No such file or
# directory". libcurl-minimal is named explicitly: left to resolve
# curl-minimal's libcurl.so.4 on its own, dnf picks the full libcurl, which
# drags in krb5-libs, openldap, cyrus-sasl-lib, libssh, libpsl, libidn2,
# shadow-utils and audit-libs -- 29 packages for an HTTPS download.
# The upgrade first pulls every published ubi9 erratum into
# micro's own packages. Weak deps stay off so nothing else lands in the rpmdb.
# git is deliberately absent (the suite arrives via COPY --from and its
# lockfile resolves every package from the npm registry).
#
# Scriptlets and triggers are off: rpm runs /bin/sh ones chrooted into
# /rootfs, and when arm64 is built under BuildKit's built-in QEMU emulator
# (no binfmt_misc on the host) that chroot cannot reach the emulator, so
# they fail with "failed to exec scriptlet interpreter /bin/sh". The few that
# matter on top of ubi-micro are replayed in the final stage instead.
RUN dnf upgrade -y --installroot /rootfs --releasever 9 \
      --setopt=reposdir=/etc/yum.repos.d --setopt=install_weak_deps=0 \
      --setopt=tsflags=nodocs,noscripts,notriggers \
    && dnf install -y --installroot /rootfs --releasever 9 \
      --setopt=reposdir=/etc/yum.repos.d --setopt=install_weak_deps=0 \
      --setopt=tsflags=nodocs,noscripts,notriggers \
      ca-certificates \
      curl-minimal \
      libcurl-minimal \
      findutils \
      grep \
      gzip \
      libstdc++ \
      sed \
      tar \
      unzip \
      zip \
    && dnf clean all --installroot /rootfs \
    && rm -rf /rootfs/var/cache/dnf /rootfs/var/log/dnf*

# ---------------------------------------------------------------------------
# final image
# ---------------------------------------------------------------------------
# FROM ubi-micro rather than scratch keeps Red Hat's image labels and default
# environment; the rootfs copy then layers the upgraded and added packages
# (and their rpmdb) over it.
FROM registry.access.redhat.com/ubi9/ubi-micro

COPY --from=rootfs /rootfs/ /

# Replay the scriptlets the rootfs stage skipped, here where every RUN --
# native or emulated -- executes normally: glibc-common's ldconfig trigger,
# p11-kit-trust's libnssckbi alternative, and ca-certificates' ca-legacy /
# update-ca-trust. (The others are Lua scriptlets that only seed files
# ubi-micro already has.) Fail if the CA bundle Go tools read was not built.
RUN ldconfig \
    && alternatives --install /usr/lib64/libnssckbi.so libnssckbi.so /usr/lib64/pkcs11/p11-kit-trust.so 30 \
    && ca-legacy install \
    && update-ca-trust \
    && test -s /etc/pki/tls/certs/ca-bundle.crt

# Set working directory
WORKDIR /app

# node only -- npm and npx stay behind in the base stage.
COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=tools /opt/java/openjdk /opt/java/openjdk
COPY --from=kubectl /out/kubectl /usr/local/bin/kubectl
COPY --from=helm /out/helm /usr/local/bin/helm
COPY --from=tools /usr/local/lib/node_modules/allure-commandline /usr/local/lib/node_modules/allure-commandline
RUN ln -s ../lib/node_modules/allure-commandline/bin/allure /usr/local/bin/allure

# Set environment variables
ENV HOME=/home/runai
ENV KUBECONFIG=/home/runai/.kube/config
ENV JAVA_HOME=/opt/java/openjdk
ENV PATH="$JAVA_HOME/bin:$PATH"
# Allow legacy OpenSSL algorithms if needed (for Node.js 18+)
ENV NODE_OPTIONS="--openssl-legacy-provider"
# Keep browser binaries out of the image (see the deps stage).
ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1

# e2e/NOTICE inventories the suite's third-party dependencies; regenerate it
# with build/scripts/generate-notice.py when package-lock.json changes.
COPY NOTICE NOTICE

# start.sh hardcodes /app/e2e. The suite arrives with node_modules already
# installed (and node-pty already compiled) by the deps stage.
COPY --from=deps /app/e2e /app/e2e
WORKDIR /app/e2e

# No `npx playwright install` — see the PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD note above.

# Fail the build if any runtime tool cannot start on this rootfs, and if npm
# or npx made it into the image. curl is the runai CLI installer's hard
# dependency (see the rootfs stage).
RUN curl --version && \
    node --version && \
    java -version && \
    kubectl version --client && \
    helm version --short && \
    allure --version && \
    ./node_modules/.bin/playwright --version && \
    ! command -v npm && ! command -v npx

# Build guard: fail the build if any glob copy in the image --
# globally installed or from the suite's node_modules -- falls inside
# CVE-2025-64756's vulnerable ranges [10.2.0, 10.5.0) or [11.0.0, 11.1.0).
RUN status=0 && \
    for p in $(find /usr/local/lib/node_modules /app/e2e/node_modules \
                 -path '*/node_modules/glob/package.json' 2>/dev/null); do \
      v="$(node -p "require('${p}').version")" && \
      case "$v" in \
        10.2.*|10.3.*|10.4.*|11.0.*) echo "vulnerable glob ${v} at ${p}" >&2; status=1 ;; \
      esac; \
    done && \
    exit "$status"

# Promote the kit entry point. start.sh sources kit-common/kit-utils.sh
# relative to /app/e2e; both ship from this checkout, so neither can fall
# behind the other. What this image needs from kit-utils.sh is that it writes
# the kubeconfig to ${KUBECONFIG:-$HOME/.kube/config} (writable by uid 65532)
# and runs ./node_modules/.bin/playwright, as there is no npx here. start.sh
# has no `set -e`, so calling a function kit-utils.sh lacks is just "command
# not found" and the step is silently skipped. Fail the build if start.sh
# calls a function kit-utils.sh does not define, if the credential mask in
# kit-utils.sh does not work with this image's sed, or if either file still
# refers to npx or a hardcoded /root/.kube outside a comment.
RUN cp /app/e2e/workloads_kit/start.sh /app/e2e/start.sh && \
    chmod +x /app/e2e/start.sh && \
    bash -c 'source kit-common/kit-utils.sh && \
      for fn in $(grep -oE "^[a-z_]+" start.sh | sort -u); do \
        type -t "$fn" >/dev/null || { echo "start.sh calls ${fn}, which kit-utils.sh does not define" >&2; exit 1; }; \
      done && \
      redact_self_test || { echo "the credential mask in kit-utils.sh is not working" >&2; exit 1; }' && \
    ! grep -nE '^[^#]*(npx |/root/\.kube)' start.sh kit-common/kit-utils.sh

# Create a non-root runtime user and writable output/configuration directories.
# Runs last so the chown covers the suite and the promoted entry point.
RUN mkdir -p \
        /app/allure-results \
        /app/allure-report \
        /app/test-results \
        /app/e2e/results \
        /home/runai/.kube \
    && printf 'runai:x:65532:\n' >> /etc/group \
    && printf 'runai:x:65532:65532:RunAI workloads kit:/home/runai:/usr/sbin/nologin\n' >> /etc/passwd \
    && chown -R 65532:65532 /app /home/runai

# Expose port for Allure report serving (optional)
EXPOSE 8080

USER 65532:65532

ENTRYPOINT ["/app/e2e/start.sh"]
