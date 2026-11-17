# Pinned: the floating "debian" tag changed the default user from root to "user".
FROM qmcgaw/godevcontainer:v0.23.0-debian

# The base image may ship an older Go; this makes the go command fetch and
# use exactly this release for the workspace and for the tools below.
ENV GOTOOLCHAIN=go1.27.4

# Install Go development tools (build-time, cached in image)
# These don't need the mounted workspace, so install them at build time
RUN go install github.com/air-verse/air@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/jmattheis/goverter/cmd/goverter@v1.9.1 && \
    go install github.com/cweill/gotests/...@latest && \ 
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 && \
    go install mvdan.cc/gofumpt@latest && \
    go install golang.org/x/tools/cmd/goimports@latest

# Copy the entrypoint script into the image, executable. The base image runs
# as the non-root "user", so a later chmod on this root-owned file would fail.
COPY --chmod=755 .devcontainer/entrypoint.sh /usr/local/bin/entrypoint.sh

# Set the entrypoint (runs every time container starts)
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Default command (can be overridden by docker-compose.yml)
CMD ["sleep", "infinity"]

