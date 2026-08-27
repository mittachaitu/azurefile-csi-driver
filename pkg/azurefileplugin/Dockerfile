# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG ARCH=amd64
ARG BUILDPLATFORM=linux/amd64
ARG BASE_IMAGE=registry.k8s.io/build-image/debian-base:bookworm-v1.0.8

FROM ${BASE_IMAGE} AS base

FROM --platform=$BUILDPLATFORM ${BASE_IMAGE} AS builder

ARG TARGETARCH=amd64

RUN apt update \
    && apt install -y curl \
    && curl -Ls https://github.com/Azure/azure-storage-azcopy/releases/download/v10.32.7/azcopy_linux_${TARGETARCH}_10.32.7.tar.gz \
        | tar xvzf - --strip-components=1 -C /usr/local/bin/ --wildcards "*/azcopy"

FROM ${BASE_IMAGE} AS azfilesauth-builder

ARG AZFILESAUTH_REPOSITORY=https://github.com/Azure/AzFilesAuthenticator.git
ARG AZFILESAUTH_REF=personal/sprasad/azsdk-identity

RUN apt update \
    && apt install -y git autoconf automake libtool build-essential python3 libcurl4-openssl-dev libkrb5-dev debhelper-compat \
    && git clone --depth 1 --branch "${AZFILESAUTH_REF}" "${AZFILESAUTH_REPOSITORY}" /src/azfilesauth \
    && cd /src/azfilesauth \
    && autoreconf -fi \
    && dpkg-buildpackage -us -uc \
    && cp /src/azfilesauth_*.deb /tmp/azfilesauth.deb

FROM base

ARG ARCH
ARG binary=./_output/${ARCH}/azurefileplugin

COPY ${binary} /azurefileplugin
COPY --from=builder --chown=root:root /usr/local/bin/azcopy /usr/local/bin/azcopy
COPY --from=azfilesauth-builder --chown=root:root /tmp/azfilesauth.deb /azfilesauth.deb

RUN apt update && apt upgrade -y && apt-mark unhold libcap2 && clean-install ca-certificates cifs-utils util-linux e2fsprogs mount udev xfsprogs nfs-common netbase libcurl4 libkrb5-3 python3 python3-pip \
    && dpkg -i /azfilesauth.deb \
    && rm -f /azfilesauth.deb

COPY ./pkg/azurefile-proxy/*.sh /azurefile-proxy/
COPY ./pkg/azurefile-proxy/azurefile-proxy.service /azurefile-proxy/
COPY ./_output/${ARCH}/azurefile-proxy /azurefile-proxy/

RUN chmod +x /azurefile-proxy/*.sh && \
  chmod +x /azurefile-proxy/azurefile-proxy.service && \
  chmod +x /azurefile-proxy/azurefile-proxy

LABEL maintainers="andyzhangx"
LABEL description="AzureFile CSI Driver"

ENTRYPOINT ["/azurefileplugin"]
