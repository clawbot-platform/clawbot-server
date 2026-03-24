FROM ubuntu:24.04

ARG ZEROCLAW_VERSION=0.1.7
ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gettext-base \
    tar \
    bash \
    util-linux \
 && rm -rf /var/lib/apt/lists/*

RUN case "${TARGETARCH}" in \
      amd64) asset="zeroclaw-x86_64-unknown-linux-gnu.tar.gz" ;; \
      arm64) asset="zeroclaw-aarch64-unknown-linux-gnu.tar.gz" ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    curl -fsSL "https://github.com/zeroclaw-labs/zeroclaw/releases/download/v${ZEROCLAW_VERSION}/${asset}" -o /tmp/zeroclaw.tar.gz && \
    tar -xzf /tmp/zeroclaw.tar.gz -C /usr/local/bin zeroclaw && \
    chmod +x /usr/local/bin/zeroclaw && \
    /usr/local/bin/zeroclaw --help >/dev/null

RUN groupadd --system zeroclaw && \
    useradd --system --gid zeroclaw --create-home --home-dir /home/zeroclaw zeroclaw && \
    mkdir -p /etc/zeroclaw /workspace && \
    chown -R zeroclaw:zeroclaw /etc/zeroclaw /workspace /home/zeroclaw

COPY deploy/docker/zeroclaw/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["gateway", "--host", "0.0.0.0", "--port", "3000"]