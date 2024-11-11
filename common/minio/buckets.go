package minio

import "fmt"

// For now there will be only two types of storage JSON and plain text,
// maybe later on I can add binary too
type StorageType string

const (
	TextType StorageType = "text/plain"
	JSONType StorageType = "application/json"
)

// StorageCategory has either Logs, Chunks or Results
type StorageCategory string

const (
	LogStorage    StorageCategory = "LOGS"
	ChunkStorage  StorageCategory = "CHUNKS"
	ResultStorage StorageCategory = "RESULTS"
)

// BucketOptions configures individual bucket properties
type BucketOptions struct {
	Name     string          `mapstructure:"name"`
	Type     StorageType     `mapstructure:"type"`
	Category StorageCategory `mapstructure:"category"`
}

// StorageOptions keeps bucket options for storage related utilities
type StorageOptions struct {
	Buckets []BucketOptions `mapstructure:"buckets"`
}

func (so *StorageOptions) Validate() error {
	for _, bucket := range so.Buckets {
		switch bucket.Type {
		case TextType, JSONType:
			// cool
		default:
			return fmt.Errorf("data type to be stored in not valid")
		}

		switch bucket.Category {
		case LogStorage, ChunkStorage, ResultStorage:
			// cool
		default:
			return fmt.Errorf("category type is not valid, only Logs, Chunks and Results are supported")
		}
	}

	return nil
}

// GetBucketByCategory returns the bucket configuration for a given category
func (so *StorageOptions) GetBucketByCategory(category StorageCategory) *BucketOptions {
	for _, bucket := range so.Buckets {
		if bucket.Category == category {
			return &bucket
		}
	}
	return nil
}
