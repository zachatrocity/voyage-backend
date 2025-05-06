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

# View only the API logs
api-logs:
    docker-compose logs -f voyage-api

# View only the mail logs
mail-logs:
    docker-compose logs -f voyage-mail

# Open a shell in the mail container
mail-shell:
    docker-compose exec voyage-mail bash

# Open a shell in the API container
api-shell:
    docker-compose exec voyage-api sh

# Run a notmuch search via CLI
search query:
    docker-compose exec voyage-mail notmuch search {{query}}

# Run a notmuch search via API
api-search query:
    curl -s "http://localhost:${API_PORT:-8080}/api/v1/search?q={{query}}" | jq

# Show all emails in the database
count:
    docker-compose exec voyage-mail notmuch count --output=messages '*'

# Run notmuch new to index new emails
index:
    docker-compose exec voyage-mail notmuch new

# Rebuild and restart the containers
restart: down build up
    just logs
