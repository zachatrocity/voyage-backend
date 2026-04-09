package notmuch

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zachatrocity/voyage/internal/email"
	"github.com/zachatrocity/voyage/notmuch"
)

const (
	dbOpenMaxAttempts = 3
	dbOpenRetryDelay  = 150 * time.Millisecond
)

// OperationError is a structured backend error for notmuch operations.
type OperationError struct {
	Operation    string
	Code         string
	Retryable    bool
	DatabasePath string
	Mode         string
	Cause        error
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("notmuch %s failed (%s): %v", e.Operation, e.Code, e.Cause)
}

func (e *OperationError) Unwrap() error {
	return e.Cause
}

func (e *OperationError) Details() map[string]any {
	return map[string]any{
		"operation":     e.Operation,
		"code":          e.Code,
		"retryable":     e.Retryable,
		"database_path": e.DatabasePath,
		"mode":          e.Mode,
	}
}

func AsOperationError(err error) (*OperationError, bool) {
	var opErr *OperationError
	if errors.As(err, &opErr) {
		return opErr, true
	}
	return nil, false
}

// EmailResult represents a single email search result
// @Description Email search result
type EmailResult struct {
	MessageID   string    `json:"message_id" example:"<12345@example.com>"`
	ThreadID    string    `json:"thread_id" example:"thread123"`
	Date        time.Time `json:"date" example:"2023-01-01T12:00:00Z"`
	From        string    `json:"from" example:"sender@example.com"`
	Subject     string    `json:"subject" example:"Flight Confirmation"`
	Tags        []string  `json:"tags" example:"travel,flight"`
	Filename    string    `json:"filename" example:"/path/to/email.eml"`
	BodyPreview string    `json:"body_preview"`
	Category    string    `json:"category"`
	TripID      string    `json:"trip_id,omitempty"`
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

func openDatabase(mode notmuch.DatabaseMode, operation string) (*notmuch.Database, error) {
	dbPath := GetDatabasePath()
	var lastStatus notmuch.Status

	for attempt := 1; attempt <= dbOpenMaxAttempts; attempt++ {
		db, status := notmuch.OpenDatabase(dbPath, mode)
		if status == notmuch.STATUS_SUCCESS {
			return db, nil
		}

		lastStatus = status
		if status != notmuch.STATUS_XAPIAN_EXCEPTION {
			break
		}

		if attempt < dbOpenMaxAttempts {
			time.Sleep(time.Duration(attempt) * dbOpenRetryDelay)
		}
	}

	return nil, &OperationError{
		Operation:    operation,
		Code:         codeFromStatus(lastStatus),
		Retryable:    lastStatus == notmuch.STATUS_XAPIAN_EXCEPTION,
		DatabasePath: dbPath,
		Mode:         modeToString(mode),
		Cause:        fmt.Errorf("failed to open notmuch database: %s", lastStatus),
	}
}

func modeToString(mode notmuch.DatabaseMode) string {
	if mode == notmuch.DATABASE_MODE_READ_WRITE {
		return "read_write"
	}
	return "read_only"
}

func codeFromStatus(status notmuch.Status) string {
	switch status {
	case notmuch.STATUS_XAPIAN_EXCEPTION:
		return "notmuch_db_open_xapian_exception"
	case notmuch.STATUS_READ_ONLY_DATABASE:
		return "notmuch_db_read_only"
	case notmuch.STATUS_FILE_ERROR:
		return "notmuch_db_file_error"
	default:
		return "notmuch_db_open_failed"
	}
}

func operationError(operation, code string, retryable bool, cause error) error {
	return &OperationError{
		Operation:    operation,
		Code:         code,
		Retryable:    retryable,
		DatabasePath: GetDatabasePath(),
		Cause:        cause,
	}
}

// CheckDatabaseConnection checks if the notmuch database is accessible
func CheckDatabaseConnection() error {
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_ONLY, "check_connection")
	if err != nil {
		return err
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
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_ONLY, "search_open")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Create a query
	q := db.CreateQuery(query)
	if q == nil {
		return nil, operationError("search_create_query", "notmuch_query_create_failed", false, fmt.Errorf("failed to create query"))
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

	// Execute the query
	messages, status := q.SearchMessages()
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("search_execute", "notmuch_query_execute_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to execute query: %s", status))
	}

	// Get the count of messages
	count, status := q.CountMessages()
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("search_count", "notmuch_query_count_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to count messages: %s", status))
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

		emailResult := createEmailResultFromMessage(msg)
		results.Results = append(results.Results, *emailResult)

		messages.MoveToNext()
		i++
	}

	return results, nil
}

// GetEmail retrieves a single email by its message ID
func GetEmail(messageID string) (*EmailResult, error) {
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_ONLY, "get_email_open")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	msg, status := db.FindMessage(messageID)
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("get_email_find_message", "notmuch_find_message_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to find message: %s", status))
	}
	if msg == nil {
		return nil, nil
	}
	defer msg.Destroy()

	result := createEmailResultFromMessage(msg)
	result.BodyPreview = email.ExtractBodyPreview(result.Filename, 300)

	return result, nil
}

// TagEmail sets a tag on a particular messageID email
func TagEmail(messageID string, tag string) (*EmailResult, error) {
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_WRITE, "tag_email_open")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	msg, status := db.FindMessage(messageID)
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("tag_email_find_message", "notmuch_find_message_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to find message: %s", status))
	}
	if msg == nil {
		return nil, nil
	}
	defer msg.Destroy()

	tagStatus := msg.AddTag(tag)
	if tagStatus != notmuch.STATUS_SUCCESS {
		return nil, operationError("tag_email_add_tag", "notmuch_add_tag_failed", tagStatus == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to add tag: %s", tagStatus))
	}

	result := createEmailResultFromMessage(msg)
	return result, nil
}

// RemoveTag removes a tag from a particular messageID email
func RemoveTag(messageID string, tag string) (*EmailResult, error) {
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_WRITE, "remove_tag_open")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	msg, status := db.FindMessage(messageID)
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("remove_tag_find_message", "notmuch_find_message_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to find message: %s", status))
	}
	if msg == nil {
		return nil, nil
	}
	defer msg.Destroy()

	tagStatus := msg.RemoveTag(tag)
	if tagStatus != notmuch.STATUS_SUCCESS {
		return nil, operationError("remove_tag_remove_tag", "notmuch_remove_tag_failed", tagStatus == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to remove tag: %s", tagStatus))
	}

	result := createEmailResultFromMessage(msg)
	return result, nil
}

// GetEmailTags retrieves the tags for a particular messageID email without modifying them
func GetEmailTags(messageID string) ([]string, error) {
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_ONLY, "get_email_tags_open")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	msg, status := db.FindMessage(messageID)
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("get_email_tags_find_message", "notmuch_find_message_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to find message: %s", status))
	}
	if msg == nil {
		return nil, nil
	}
	defer msg.Destroy()

	tags := []string{}
	msgTags := msg.GetTags()
	for msgTags.Valid() {
		tags = append(tags, msgTags.Get())
		msgTags.MoveToNext()
	}

	return tags, nil
}

// FilterTripTags returns only tags that have the "trip:" prefix from a tag list.
// This is exported for unit testing the tag-filtering logic without a live DB.
func FilterTripTags(tags []string) []string {
	var tripTags []string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "trip:") {
			tripTags = append(tripTags, tag)
		}
	}
	return tripTags
}

// ReplaceTripTag removes any existing trip:* tags from the email and applies trip:{newTripID}.
// Returns the updated EmailResult.
func ReplaceTripTag(messageID, newTripID string) (*EmailResult, error) {
	db, err := openDatabase(notmuch.DATABASE_MODE_READ_WRITE, "replace_trip_tag_open")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	msg, status := db.FindMessage(messageID)
	if status != notmuch.STATUS_SUCCESS {
		return nil, operationError("replace_trip_tag_find_message", "notmuch_find_message_failed", status == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to find message: %s", status))
	}
	if msg == nil {
		return nil, nil
	}
	defer msg.Destroy()

	tags := []string{}
	msgTags := msg.GetTags()
	for msgTags.Valid() {
		tags = append(tags, msgTags.Get())
		msgTags.MoveToNext()
	}

	if st := msg.Freeze(); st != notmuch.STATUS_SUCCESS {
		return nil, operationError("replace_trip_tag_freeze", "notmuch_freeze_failed", st == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to freeze message for tag update: %s", st))
	}

	for _, tag := range FilterTripTags(tags) {
		if st := msg.RemoveTag(tag); st != notmuch.STATUS_SUCCESS {
			cause := fmt.Errorf("failed to remove tag %q: %s", tag, st)
			if thawStatus := msg.Thaw(); thawStatus != notmuch.STATUS_SUCCESS {
				cause = fmt.Errorf("%w (also failed to thaw after remove: %s)", cause, thawStatus)
			}
			return nil, operationError("replace_trip_tag_remove_existing", "notmuch_remove_tag_failed", st == notmuch.STATUS_XAPIAN_EXCEPTION, cause)
		}
	}

	if st := msg.AddTag("trip:" + newTripID); st != notmuch.STATUS_SUCCESS {
		cause := fmt.Errorf("failed to add new trip tag: %s", st)
		if thawStatus := msg.Thaw(); thawStatus != notmuch.STATUS_SUCCESS {
			cause = fmt.Errorf("%w (also failed to thaw after add: %s)", cause, thawStatus)
		}
		return nil, operationError("replace_trip_tag_add_new", "notmuch_add_tag_failed", st == notmuch.STATUS_XAPIAN_EXCEPTION, cause)
	}

	if st := msg.Thaw(); st != notmuch.STATUS_SUCCESS {
		return nil, operationError("replace_trip_tag_thaw", "notmuch_thaw_failed", st == notmuch.STATUS_XAPIAN_EXCEPTION, fmt.Errorf("failed to commit trip tag changes: %s", st))
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
	if msgTags != nil {
		for msgTags.Valid() {
			tags = append(tags, msgTags.Get())
			msgTags.MoveToNext()
		}
	}

	// Extract TripID from trip:* tags
	var tripID string
	for _, tag := range tags {
		if strings.HasPrefix(tag, "trip:") {
			tripID = strings.TrimPrefix(tag, "trip:")
			break
		}
	}

	return &EmailResult{
		MessageID: msg.GetMessageId(),
		ThreadID:  msg.GetThreadId(),
		Date:      date,
		From:      msg.GetHeader("from"),
		Subject:   msg.GetHeader("subject"),
		Tags:      tags,
		Filename:  msg.GetFileName(),
		TripID:    tripID,
	}
}
