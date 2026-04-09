# Missing query should return 400
GET {{base_url}}/api/v1/search
X-API-Key: {{api_key}}
HTTP 400

# Valid-but-missing email should return 404
GET {{base_url}}/api/v1/email/nonexistent-hurl-message-id@example.invalid
X-API-Key: {{api_key}}
HTTP 404

# Missing trip should return 404
GET {{base_url}}/api/v1/trips/nonexistent-trip-id/emails
X-API-Key: {{api_key}}
HTTP 404
