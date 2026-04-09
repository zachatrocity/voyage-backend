package handlers

import "strings"

// isLikelyMessageID performs lightweight validation for message-id-like values
// before calling notmuch APIs. Voyage stores message IDs mostly without angle
// brackets, so we accept either style as long as they contain an @ and no spaces.
func isLikelyMessageID(messageID string) bool {
	id := strings.TrimSpace(messageID)
	if id == "" {
		return false
	}
	if strings.ContainsAny(id, " \t\n\r") {
		return false
	}
	return strings.Contains(id, "@")
}
