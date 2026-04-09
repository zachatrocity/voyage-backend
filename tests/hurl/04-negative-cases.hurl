# Missing query should return 400
GET {{base_url}}/api/v1/search
X-API-Key: {{api_key}}
HTTP 400

# Invalid message-id shape should return 400
GET {{base_url}}/api/v1/email/not-a-message-id
X-API-Key: {{api_key}}
HTTP 400

# Valid-but-missing email should return 404
GET {{base_url}}/api/v1/email/nonexistent-hurl-message-id@example.invalid
X-API-Key: {{api_key}}
HTTP 404

# Tagging with invalid message-id should return 400
POST {{base_url}}/api/v1/email/not-a-message-id/tags/inbox
X-API-Key: {{api_key}}
HTTP 400

# Missing trip should return 404
GET {{base_url}}/api/v1/trips/nonexistent-trip-id/emails
X-API-Key: {{api_key}}
HTTP 404
