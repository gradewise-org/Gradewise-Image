# Gradewise-Image Dockerfile
# Similar structure to Gradescope but with Go-based harness

FROM fedora:38

# Install required packages
RUN dnf install -y \
    openssh-server \
    openssh-clients \
    tar \
    gzip \
    bash \
    python3 \
    curl \
    && dnf clean all

# Create necessary directories
RUN mkdir -p /var/run/sshd /autograder /gradescope /root/.ssh

# Install dumb-init for proper signal handling
# Note: You'll need to download dumb-init binary or build it
# For now, we'll use a simple approach or install via package manager
RUN dnf install -y dumb-init || ( \
    curl -L https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64 -o /usr/bin/dumb-init && \
    chmod +x /usr/bin/dumb-init \
)

# Install Go (for building the harness)
RUN dnf install -y golang || ( \
    curl -L https://go.dev/dl/go1.21.5.linux-amd64.tar.gz -o /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz \
)

# Set up Go environment
ENV PATH=/usr/local/go/bin:$PATH
ENV GOPATH=/go
ENV GOCACHE=/tmp/go-cache

# Copy Go module files
WORKDIR /build
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build harness binary
RUN go build -o /usr/local/bin/harness ./cmd/harness

# Build sshd-setup binary
RUN go build -o /usr/local/bin/sshd-setup ./cmd/sshd-setup

# Copy SSH configuration
COPY sshd_config /etc/ssh/sshd_config

# Create wrapper script for harness update and run
RUN echo '#!/bin/bash\n\
set -e\n\
/usr/local/bin/harness "$@"\n\
' > /autograder/update_and_run_harness.sh && \
    chmod +x /autograder/update_and_run_harness.sh

# Create SSH wrapper script (kills container on session exit)
RUN echo '#!/bin/bash\n\
exec "$@"\n\
# Kill container when SSH session exits\n\
kill 1\n\
' > /usr/local/sbin/ssh_wrapper.sh && \
    chmod +x /usr/local/sbin/ssh_wrapper.sh

# Set up environment
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8

# Set entrypoint
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Default command runs the harness
CMD ["/autograder/update_and_run_harness.sh"]
