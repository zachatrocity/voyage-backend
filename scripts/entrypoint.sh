#!/bin/bash
set -e

# Function to check required configurations
check_configs() {
    local configs_ok=0  # 0 = success, 1 = failure

    # Check for mbsync configuration
    if [ ! -f "/config/.mbsyncrc" ]; then
        echo "ERROR: mbsync configuration file not found at /config/.mbsyncrc"
        echo "Please mount your mbsync configuration file to /config/.mbsyncrc"
        configs_ok=1
    else
        echo "mbsync configuration found."
    fi

    # Check for notmuch configuration
    if [ ! -f "/config/notmuch/config" ]; then
        echo "ERROR: notmuch configuration file not found at /config/notmuch/config"
        echo "Please mount your notmuch configuration file to /config/notmuch/config"
        configs_ok=1
    else
        echo "notmuch configuration found."
        # Set NOTMUCH_CONFIG environment variable
        export NOTMUCH_CONFIG=/config/notmuch/config
    fi

    return $configs_ok
}

# Function to setup cron job
setup_cron() {
    local frequency=${SYNC_FREQUENCY:-15m}
    
    # Convert frequency to cron format
    case $frequency in
        *m)
            # Minutes
            min=${frequency%m}
            cron_expr="*/${min} * * * *"
            ;;
        *h)
            # Hours
            hour=${frequency%h}
            cron_expr="0 */${hour} * * *"
            ;;
        *d)
            # Days
            day=${frequency%d}
            cron_expr="0 0 */${day} * *"
            ;;
        *)
            # Default to every 15 minutes
            cron_expr="*/15 * * * *"
            ;;
    esac
    
    echo "Setting up cron job with frequency: $frequency (cron: $cron_expr)"
    echo "$cron_expr /scripts/sync-mail.sh >> /var/log/cron.log 2>&1" > /etc/crontabs/root
    echo "Cron job set up."
}

# Main execution
echo "Starting mail service..."

# Check for required configurations
if check_configs; then
    echo "All required configurations found."
    
    # Setup cron job
    setup_cron
    
    # Run initial sync
    echo "Running initial mail sync..."
    /scripts/sync-mail.sh
    
    # Start cron daemon
    echo "Starting cron daemon..."
    crond -f -l 8
else
    echo "ERROR: Missing required configurations. The service will not start."
    echo "Please provide the necessary configuration files and restart the container."
    exit 1
fi
