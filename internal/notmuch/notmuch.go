package notmuch

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/zachatrocity/voyage/notmuch"
)

// EmailResult represents a single email search result
// @Description Email search result
type EmailResult struct {
	MessageID string    `json:"message_id" example:"<12345@example.com>"`
	ThreadID  string    `json:"thread_id" example:"thread123"`
	Date      time.Time `json:"date" example:"2023-01-01T12:00:00Z"`
	From      string    `json:"from" example:"sender@example.com"`
	Subject   string    `json:"subject" example:"Flight Confirmation"`
	Tags      []string  `json:"tags" example:"travel,flight"`
	Filename  string    `json:"filename" example:"/path/to/email.eml"`
}

// SearchResults represents the results of a search query
// @Description Search results containing matching emails
type SearchResults struct {
	Query   string        `json:"query" example:"subject:flight"`
	Count   int           `json:"count" example:"42"`
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

// SortType represents the sort order for search results
type SortType int

const (
	// SortOldestFirst sorts messages with oldest first
	SortOldestFirst SortType = iota
	// SortNewestFirst sorts messages with newest first
	SortNewestFirst
	// SortMessageID sorts messages by message ID
	SortMessageID
	// SortUnsorted does not apply any sorting
	SortUnsorted
)

// Search performs a search against the notmuch database
func Search(query string, limitStr string, sortType SortType) (*SearchResults, error) {
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

	// Map our SortType to notmuch.Sort
	var notmuchSort notmuch.Sort
	switch sortType {
	case SortOldestFirst:
		notmuchSort = 0
	case SortNewestFirst:
		notmuchSort = 1
	default:
		notmuchSort = 1 // Default to newest first
	}

	// Set the sort order
	q.SetSort(notmuchSort)

	// Log the sort order for debugging
	fmt.Printf("Search with query: %s notmuch sort: %d\n", query, notmuchSort)

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

		// Create email result using helper function and add to results
		emailResult := createEmailResultFromMessage(msg)
		results.Results = append(results.Results, *emailResult)

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

	// Create result using helper function
	result := createEmailResultFromMessage(msg)

	return result, nil
}

// TagEmail sets a tag on a particular messageID email
func TagEmail(messageID string, tag string) (*EmailResult, error) {
	// Open the database
	db, status := notmuch.OpenDatabase(GetDatabasePath(), notmuch.DATABASE_MODE_READ_WRITE)
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

	tagStatus := msg.AddTag(tag)
	if tagStatus != notmuch.STATUS_SUCCESS {
		return nil, fmt.Errorf("failed to add tag: %s", status)
	}

	result := createEmailResultFromMessage(msg)

	return result, nil
}

// createEmailResultFromMessage creates an EmailResult from a notmuch Message
func createEmailResultFromMessage(msg *notmuch.Message) *EmailResult {
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
	return &EmailResult{
		MessageID: msg.GetMessageId(),
		ThreadID:  msg.GetThreadId(),
		Date:      date,
		From:      msg.GetHeader("from"),
		Subject:   msg.GetHeader("subject"),
		Tags:      tags,
		Filename:  msg.GetFileName(),
	}
}
