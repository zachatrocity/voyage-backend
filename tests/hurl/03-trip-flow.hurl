# Create a throwaway trip for integration testing
POST {{base_url}}/api/v1/trips
X-API-Key: {{api_key}}
Content-Type: application/json
{
  "name": "Hurl Smoke {{run_id}}",
  "date_range": "2026-01-01 to 2026-01-02"
}
HTTP 201
[Captures]
trip_id: jsonpath "$.id"
[Asserts]
jsonpath "$.id" exists
jsonpath "$.name" == "Hurl Smoke {{run_id}}"

# Get one message id for association
GET {{base_url}}/api/v1/search?q={{search_query}}&limit=1
X-API-Key: {{api_key}}
HTTP 200
[Captures]
message_id: jsonpath "$.results[0].message_id"
[Asserts]
jsonpath "$.results[0].message_id" exists

# Associate the message to the new trip
POST {{base_url}}/api/v1/email/{{message_id}}/trip/{{trip_id}}
X-API-Key: {{api_key}}
HTTP 200
[Asserts]
jsonpath "$.trip_id" == "{{trip_id}}"

# Verify trip email listing includes at least one item
GET {{base_url}}/api/v1/trips/{{trip_id}}/emails
X-API-Key: {{api_key}}
HTTP 200
[Asserts]
jsonpath "$.emails[0].message_id" exists

# Cleanup
DELETE {{base_url}}/api/v1/trips/{{trip_id}}
X-API-Key: {{api_key}}
HTTP 204
