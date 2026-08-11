# Dockerfile for PocketBase
FROM alpine:3.23.3

ARG PB_VERSION=0.36.6
ARG PB_ARCH=linux_amd64

RUN apk add --no-cache \
    unzip \
    ca-certificates \
    curl

# Download PocketBase
WORKDIR /pb
RUN PB_ZIP="pocketbase_${PB_VERSION}_${PB_ARCH}.zip" && \
    curl -L -o "${PB_ZIP}" "https://github.com/pocketbase/pocketbase/releases/download/v${PB_VERSION}/${PB_ZIP}" && \
    curl -L -o checksums.txt "https://github.com/pocketbase/pocketbase/releases/download/v${PB_VERSION}/checksums.txt" && \
    grep " ${PB_ZIP}$" checksums.txt | sha256sum -c - && \
    unzip "${PB_ZIP}" && \
    rm "${PB_ZIP}" checksums.txt

# Copy migrations (These should be committed in your repo)
COPY pb_migrations /pb/pb_migrations

# Copy hooks (JavaScript server-side logic)
COPY pb_hooks /pb/pb_hooks

# Expose port
EXPOSE 8090

# Run with migrations
CMD ["./pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb/pb_data", "--migrationsDir=/pb/pb_migrations"]
