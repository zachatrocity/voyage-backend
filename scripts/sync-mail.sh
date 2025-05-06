#!/bin/bash
set -e

# Log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Check for required configurations
if [ ! -f "/config/.mbsyncrc" ]; then
    log "ERROR: mbsync configuration file not found at /config/.mbsyncrc"
    log "Please mount your mbsync configuration file to /config/.mbsyncrc"
    exit 1
fi

if [ ! -f "/config/notmuch/config" ]; then
    log "ERROR: notmuch configuration file not found at /config/notmuch/config"
    log "Please mount your notmuch configuration file to /config/notmuch/config"
    exit 1
fi

# Set NOTMUCH_CONFIG environment variable
export NOTMUCH_CONFIG=/config/notmuch/config

# Run mbsync
log "Starting email synchronization with mbsync..."
mbsync -c /config/.mbsyncrc -a
sync_status=$?

if [ $sync_status -eq 0 ]; then
    log "Email synchronization completed successfully."
else
    log "Email synchronization failed with status $sync_status."
    # Don't exit on error, continue with notmuch indexing
    log "Continuing with notmuch indexing..."
fi

# Run notmuch new to index new emails
log "Indexing new emails with notmuch..."
notmuch new
index_status=$?

if [ $index_status -eq 0 ]; then
    log "Email indexing completed successfully."
else
    log "Email indexing failed with status $index_status."
    exit $index_status
fi

log "Mail sync and index process completed."
