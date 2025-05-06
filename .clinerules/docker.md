## Brief overview
This set of guidelines is specific to the Voyage project, a self-hosted travel plan aggregator that searches through emails to identify and organize travel-related information.

## Technology stack
- Backend should be implemented in Go
- Use Docker and Docker Compose for containerization and deployment
- Prefer notmuch for email indexing and searching
- Use mbsync (isync) for email fetching
- Implement justfile for command simplification and developer experience

## Configuration approach
- Use environment variables for configurable settings (email credentials, sync frequency)
- Store configuration files in a dedicated config directory
- Provide example configuration files for first-time setup
- Support both new notmuch database creation and pointing to existing databases

## Development workflow
- Use shell scripts for container initialization and recurring tasks
- Implement a simple command interface via justfile
- Ensure all commands are intuitive and self-explanatory
- Provide clear error messages and logging

## Container design
- Use Alpine Linux as the base image for smaller container size
- Include only necessary packages to minimize attack surface
- Set up proper volume mounts for persistent data
- Configure automatic restart unless explicitly stopped

## Email processing
- Configure mbsync to fetch emails at configurable intervals
- Use notmuch for efficient email indexing and searching
- Implement tagging system for organizing emails into travel plans
- Support searching by various criteria (subject, sender, content)
