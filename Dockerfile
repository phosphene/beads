# Base stage for building
FROM golang:1.23-bullseye AS builder

WORKDIR /app
ENV GOTOOLCHAIN=auto

# Install build dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    make \
    libicu-dev \
    gcc \
    g++ \
    && rm -rf /var/lib/apt/lists/*

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the bd binary
RUN make build

# Runtime stage for the application
FROM debian:bullseye-slim AS runtime

WORKDIR /app

# Install runtime dependencies (dolt + ICU)
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    libicu67 \
    && rm -rf /var/lib/apt/lists/*

# Install Dolt
RUN curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | bash

# Copy the built binary
COPY --from=builder /app/bd /usr/local/bin/bd

# Verify installation
RUN bd --version && dolt version

VOLUME /root/.beads
CMD ["bd"]

# Test stage for BDD runner
FROM golang:1.23-bullseye AS tester

WORKDIR /app
ENV GOTOOLCHAIN=auto

# Install test dependencies (dolt + ICU + Docker client)
RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    libicu-dev \
    docker.io \
    && rm -rf /var/lib/apt/lists/*

# Install Dolt
RUN curl -L https://github.com/dolthub/dolt/releases/latest/download/install.sh | bash

# Copy the built binary and source code for tests
COPY --from=builder /app/bd /usr/local/bin/bd
COPY . .

# Ensure dependencies are available for testing
RUN go mod download

VOLUME /root/.beads
CMD ["go", "test", "-v", "./tests/bdd/..."]
