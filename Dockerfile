FROM alpine:3.18

# Install required packages
RUN apk add --no-cache \
    notmuch \
    isync \
    ca-certificates \
    dcron \
    bash \
    tzdata

# Create directories
RUN mkdir -p /mail /config /scripts

# Copy scripts
COPY scripts/entrypoint.sh /scripts/
COPY scripts/sync-mail.sh /scripts/

# Make scripts executable
RUN chmod +x /scripts/*.sh

# Set working directory
WORKDIR /mail

# Set entrypoint
ENTRYPOINT ["/scripts/entrypoint.sh"]
