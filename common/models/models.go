// models/types.go
package models

import "time"

// JobStatus represents the current state of a grep job
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
)

// Job represents a grep search job
type Job struct {
	ID          string     `json:"id"`              // Unique identifier for the job
	RequestID   string     `json:"request_id"`      // Original request ID from API
	Pattern     string     `json:"pattern"`         // Grep pattern to search
	Files       []string   `json:"files"`           // List of files/patterns to search
	Status      JobStatus  `json:"status"`          // Current job status
	CreatedAt   time.Time  `json:"created_at"`      // Job creation timestamp
	StartedAt   *time.Time `json:"started_at"`      // When job processing began
	CompletedAt *time.Time `json:"completed_at"`    // When job finished (success/failure)
	Error       string     `json:"error,omitempty"` // Error message if job failed
	Progress    float64    `json:"progress"`        // Progress percentage (0-100)
	IsCancelled bool       `json:"is_cancelled"`    // Whether job was cancelled

	// Search options
	CaseSensitive bool `json:"case_sensitive"` // Whether search is case-sensitive
	Regex         bool `json:"regex"`          // Whether pattern is regex
	ContextLines  int  `json:"context_lines"`  // Number of context lines
}

// Chunk represents a portion of a file to be processed
type Chunk struct {
	ID        string    `json:"id"`         // Unique identifier for the chunk
	JobID     string    `json:"job_id"`     // Parent job ID
	FileName  string    `json:"file_name"`  // Original file name
	StartByte int64     `json:"start_byte"` // Starting byte position
	EndByte   int64     `json:"end_byte"`   // Ending byte position
	Size      int64     `json:"size"`       // Chunk size in bytes
	CreatedAt time.Time `json:"created_at"` // When chunk was created

	// Optional: for context-aware splitting
	StartLine int `json:"start_line"` // Starting line number
	EndLine   int `json:"end_line"`   // Ending line number
}

// Match represents a single grep match
type Match struct {
	LineNumber int     `json:"line_number"` // Line number in original file
	Content    string  `json:"content"`     // The matching line content
	FileName   string  `json:"file_name"`   // Source file name
	Context    Context `json:"context"`     // Surrounding context lines
}

// Context holds lines before and after a match
type Context struct {
	Before []string `json:"before"` // Lines before match
	After  []string `json:"after"`  // Lines after match
}

// Result represents processed results from a mapper
type Result struct {
	ID        string    `json:"id"`         // Unique identifier for this result
	JobID     string    `json:"job_id"`     // Parent job ID
	ChunkID   string    `json:"chunk_id"`   // Source chunk ID
	Matches   []Match   `json:"matches"`    // Grep matches found
	CreatedAt time.Time `json:"created_at"` // When result was created

	// Statistics
	ProcessedBytes int64 `json:"processed_bytes"` // Number of bytes processed
	ProcessedLines int   `json:"processed_lines"` // Number of lines processed
	MatchCount     int   `json:"match_count"`     // Number of matches found
}

// LogFile represents a file available for grepping
type LogFile struct {
	Name      string    `json:"name"`       // File name
	Path      string    `json:"path"`       // Full path in storage
	Size      int64     `json:"size"`       // File size in bytes
	UpdatedAt time.Time `json:"updated_at"` // Last modification time
}

// JobStats holds statistical information about a job
type JobStats struct {
	TotalFiles      int   `json:"total_files"`      // Total number of files to process
	ProcessedFiles  int   `json:"processed_files"`  // Number of files processed
	TotalChunks     int   `json:"total_chunks"`     // Total number of chunks created
	ProcessedChunks int   `json:"processed_chunks"` // Number of chunks processed
	TotalMatches    int   `json:"total_matches"`    // Total matches found
	BytesProcessed  int64 `json:"bytes_processed"`  // Total bytes processed
}

// Message types for NATS
type ChunkMessage struct {
	Chunk
	Pattern       string `json:"pattern"` // Search pattern
	CaseSensitive bool   `json:"case_sensitive"`
	Regex         bool   `json:"regex"`
	ContextLines  int    `json:"context_lines"`
}

type ResultMessage struct {
	Result
	Stats JobStats `json:"stats"` // Partial stats from this chunk
}

// Redis key types (for consistent key formatting)
type RedisKeys struct{}

func (k RedisKeys) JobKey(jobID string) string {
	return "job:" + jobID
}

func (k RedisKeys) JobStatsKey(jobID string) string {
	return "job:" + jobID + ":stats"
}

func (k RedisKeys) ChunkKey(jobID, chunkID string) string {
	return "job:" + jobID + ":chunk:" + chunkID
}
