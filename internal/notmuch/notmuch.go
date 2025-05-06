package notmuch

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/zachatrocity/voyage/notmuch"
)

// EmailResult represents a single email search result
type EmailResult struct {
	MessageID string    `json:"message_id"`
	ThreadID  string    `json:"thread_id"`
	Date      time.Time `json:"date"`
	From      string    `json:"from"`
	Subject   string    `json:"subject"`
	Tags      []string  `json:"tags"`
	Filename  string    `json:"filename"`
}

// SearchResults represents the results of a search query
type SearchResults struct {
	Query   string        `json:"query"`
	Count   int           `json:"count"`
	Results []EmailResult `json:"results"`
}

// GetDatabasePath returns the path to the notmuch database
func GetDatabasePath() string {
	// Check environment variable first
	path := os.Getenv("NOTMUCH_DATABASE")
	if path != "" {
		return path
	}

	// Default to /mail
	return "/mail"
}

// CheckDatabaseConnection checks if the notmuch database is accessible
func CheckDatabaseConnection() error {
	db, status := notmuch.OpenDatabase(GetDatabasePath(), notmuch.DATABASE_MODE_READ_ONLY)
	if status != notmuch.STATUS_SUCCESS {
		return fmt.Errorf("failed to open notmuch database: %s", status)
	}
	defer db.Close()
	return nil
}

// Search performs a search against the notmuch database
func Search(query string, limitStr string) (*SearchResults, error) {
	// Convert limit to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50 // Default limit
	}

	// Open the database
	db, status := notmuch.OpenDatabase(GetDatabasePath(), notmuch.DATABASE_MODE_READ_ONLY)
	if status != notmuch.STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to open notmuch database: %s", status)
	}
	defer db.Close()

	// Create a query
	q := db.CreateQuery(query)
	if q == nil {
		return nil, fmt.Errorf("failed to create query")
	}
	defer q.Destroy()

	// Set sort order to newest first
	q.SetSort(notmuch.SORT_NEWEST_FIRST)

	// Execute the query
	var messages *notmuch.Messages
	status = notmuch.STATUS_SUCCESS
	messages, status = q.SearchMessages()
	if status != notmuch.STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to execute query: %s", status)
	}

	// Get the count of messages
	var count uint
	count, status = q.CountMessages()
	if status != notmuch.STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to count messages: %s", status)
	}

	// Create results
	results := &SearchResults{
		Query:   query,
		Count:   int(count),
		Results: []EmailResult{},
	}

	// Iterate through messages
	i := 0
	for messages.Valid() && i < limit {
		msg := messages.Get()
		if msg == nil {
			messages.MoveToNext()
			continue
		}

		// Get message date
		timestamp, _ := msg.GetDate()
		date := time.Unix(timestamp, 0)

		// Get tags
		tags := []string{}
		msgTags := msg.GetTags()
		for msgTags.Valid() {
			tags = append(tags, msgTags.Get())
			msgTags.MoveToNext()
		}

		// Add to results
		results.Results = append(results.Results, EmailResult{
			MessageID: msg.GetMessageId(),
			ThreadID:  msg.GetThreadId(),
			Date:      date,
			From:      msg.GetHeader("from"),
			Subject:   msg.GetHeader("subject"),
			Tags:      tags,
			Filename:  msg.GetFileName(),
		})

		messages.MoveToNext()
		i++
	}

	return results, nil
}

// GetEmail retrieves a single email by its message ID
func GetEmail(messageID string) (*EmailResult, error) {
	// Open the database
	db, status := notmuch.OpenDatabase(GetDatabasePath(), notmuch.DATABASE_MODE_READ_ONLY)
	if status != notmuch.STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to open notmuch database: %s", status)
	}
	defer db.Close()

	// Find the message
	msg, status := db.FindMessage(messageID)
	if status != notmuch.STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to find message: %s", status)
	}
	if msg == nil {
		return nil, nil // Message not found
	}
	defer msg.Destroy()

	// Get message date
	timestamp, _ := msg.GetDate()
	date := time.Unix(timestamp, 0)

	// Get tags
	tags := []string{}
	msgTags := msg.GetTags()
	for msgTags.Valid() {
		tags = append(tags, msgTags.Get())
		msgTags.MoveToNext()
	}

	// Create result
	result := &EmailResult{
		MessageID: msg.GetMessageId(),
		ThreadID:  msg.GetThreadId(),
		Date:      date,
		From:      msg.GetHeader("from"),
		Subject:   msg.GetHeader("subject"),
		Tags:      tags,
		Filename:  msg.GetFileName(),
	}

	return result, nil
}
