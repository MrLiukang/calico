# Copyright (c) 2015-2024 Tigera, Inc. All rights reserved.
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

# Note: the felix-test image is really a "test environment" image.  It doesn't include Felix itself and we don't ship it.

ARG BPFTOOL_IMAGE

FROM ${BPFTOOL_IMAGE} AS bpftool

FROM debian:12-slim AS source

LABEL maintainer="Shaun Crampton <shaun@tigera.io>"

# Install remaining runtime deps required for felix from the global repository
RUN apt-get update && apt-get install -y \
    ipset \
    iptables \
    iproute2 \
    iputils-arping \
    iputils-ping \
    iputils-tracepath \
    # Need arp
    net-tools \
    conntrack \
    runit \
    # Need kmod to ensure ip6tables-save works correctly
    kmod \
    # Need netbase in order for ipset to work correctly
    # See https://github.com/kubernetes/kubernetes/issues/68703
    netbase \
    libnftables-dev \
    # Also needed (provides utilities for browsing procfs like ps)
    procps \
    ca-certificates \
    # Felix FV tests require these
    wget \
    ethtool \
    hping3 \
    libpcap0.8 \
    tcpdump \
    iperf3 \
    iperf \
    jq

RUN apt-get install --no-install-recommends -y wireguard-tools

RUN apt-get clean

# Default to the xtables backend of iptables.
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy
RUN update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

# Felix is built with RHEL/UBI and links against libpcap.so.1. We need this symbolic link
# until Debian changes the soname from .0.8 to .1.
RUN ln -s $(readlink /usr/lib/*-linux-gnu/libpcap.so.0.8) $(ldconfig -v 2>/dev/null | grep '^/lib/' | head -n 1 | cut -f1 -d:)/libpcap.so.1

COPY felix.cfg /etc/calico/felix.cfg
COPY calico-felix-wrapper /usr/bin

COPY --from=bpftool /bpftool /usr/bin

FROM scratch

COPY --from=source / /

WORKDIR /code

# Run felix (via the wrapper script) by default
CMD ["/usr/bin/calico-felix-wrapper"]
