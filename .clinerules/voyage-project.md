## Brief overview
This set of guidelines is specific to the Voyage project, a self-hosted travel plan aggregator that searches through emails to identify and organize travel-related information.

## Technology stack
- Backend should be implemented in Go
- Use Docker and Docker Compose for containerization and deployment
- Prefer notmuch for email indexing and searching
- Use mbsync for email fetching

## Principles
- Keep implementation simple but extensible
- Avoid adding additional persistence layers when possible
- Use notmuch's tagging system for organizing emails into "plans"
- Ensure the system can be self-hosted with minimal configuration


