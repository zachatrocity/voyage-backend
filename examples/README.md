# Configuration Examples

This directory contains example configuration files for notmuch and mbsync that can be used with the Voyage Docker setup.

## Setup Instructions

1. Create the necessary directories:
   ```bash
   mkdir -p config/notmuch
   ```

2. Copy the example configuration files to the appropriate locations:
   ```bash
   cp examples/mbsyncrc.example config/.mbsyncrc
   cp examples/notmuch-config.example config/notmuch/config
   ```

3. Edit the configuration files to match your email settings:
   ```bash
   # Edit mbsync configuration
   nano config/.mbsyncrc
   
   # Edit notmuch configuration
   nano config/notmuch/config
   ```

4. Start the Docker container:
   ```bash
   docker-compose up -d
   ```

## Configuration Details

### mbsync Configuration

The `mbsyncrc.example` file contains a basic configuration for syncing with Gmail. You'll need to:

1. Replace `your.email@gmail.com` with your Gmail address
2. Replace `your_app_specific_password` with an app-specific password generated from your Google account
   (See: https://support.google.com/accounts/answer/185833)

For other email providers, you'll need to adjust the host, port, and SSL settings accordingly.

### notmuch Configuration

The `notmuch-config.example` file contains a basic configuration for notmuch. You'll need to:

1. Replace `Your Name` with your name
2. Replace `your.email@example.com` with your email address

The configuration sets up basic tags for new emails and configures the database path to match the Docker volume mount.
