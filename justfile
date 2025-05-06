# Voyage justfile
# Run commands with: just <command>

# Build the Docker image
build:
    docker-compose build

# Start the Docker container
up:
    docker-compose up -d

# Stop the Docker container
down:
    docker-compose down

# View the container logs
logs:
    docker-compose logs -f

# Open a shell in the container
shell:
    docker-compose exec voyage bash

# Run a notmuch search
search query:
    docker-compose exec voyage notmuch search {{query}}

# Show all emails in the database
count:
    docker-compose exec voyage notmuch count --output=messages '*'

# Run notmuch new to index new emails
index:
    docker-compose exec voyage notmuch new

# Rebuild and restart the container
restart: down build up
    just logs
