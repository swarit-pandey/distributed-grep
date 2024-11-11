package minio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	gominio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/swarit-pandey/distributed-grep/common/logger"
	"github.com/swarit-pandey/distributed-grep/common/models"
)

// Storage represents the MinIO client wrapper
type Storage struct {
	client         *MinOptions
	minioClient    *gominio.Client
	storageOptions *StorageOptions
}

// New returns a new Storage with access to storage APIs for MinIO
func New(endpoint, accessKeyID, secretAccessKey string, ssl bool, log *logger.Logger, opts *StorageOptions) (*Storage, error) {
	if log == nil {
		InitLogger(nil)
	}

	minioOptions := MinOptions{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SSL:             ssl,
	}

	c := &Storage{
		client:         &minioOptions,
		storageOptions: opts,
	}

	return c, opts.Validate()
}

// Instantiate initializes a new minio instance and ensures buckets exist
func (s *Storage) Instantiate(ctx context.Context) error {
	minioClient, err := gominio.New(s.client.Endpoint, &gominio.Options{
		Creds:  credentials.NewStaticV4(s.client.AccessKeyID, s.client.SecretAccessKey, ""),
		Secure: s.client.SSL,
	})
	if err != nil {
		log.Error("failed to init minio client", "err", err)
		return fmt.Errorf("failed to initialize minio client: %w", err)
	}

	s.minioClient = minioClient

	// Ensure all required buckets exist
	for _, bucket := range s.storageOptions.Buckets {
		exists, err := s.minioClient.BucketExists(ctx, bucket.Name)
		if err != nil {
			log.Error("failed to check bucket existence", "bucket", bucket.Name, "err", err)
			return fmt.Errorf("failed to check bucket %s: %w", bucket.Name, err)
		}

		if !exists {
			err = s.minioClient.MakeBucket(ctx, bucket.Name, gominio.MakeBucketOptions{})
			if err != nil {
				log.Error("failed to create bucket", "bucket", bucket.Name, "err", err)
				return fmt.Errorf("failed to create bucket %s: %w", bucket.Name, err)
			}
			log.Info("created bucket", "bucket", bucket.Name)
		}
	}

	log.Info("minio initialized successfully", "endpoint", s.client.Endpoint)
	return nil
}

// UploadLogFile uploads a single log file
func (s *Storage) UploadLogFile(ctx context.Context, logFile models.LogFile, reader io.Reader) error {
	bucket := s.storageOptions.GetBucketByCategory(LogStorage)

	_, err := s.minioClient.PutObject(ctx, bucket.Name, logFile.Path, reader, -1,
		gominio.PutObjectOptions{
			ContentType: string(bucket.Type),
			UserMetadata: map[string]string{
				"uploaded_at": time.Now().Format(time.RFC3339),
			},
		})
	if err != nil {
		log.Error("failed to upload log file", "file", logFile.Name, "err", err)
		return fmt.Errorf("failed to upload log file %s: %w", logFile.Name, err)
	}

	log.Info("uploaded log file", "file", logFile.Name)
	return nil
}

// ListLogFiles returns all available log files
func (s *Storage) ListLogFiles(ctx context.Context) ([]models.LogFile, error) {
	bucket := s.storageOptions.GetBucketByCategory(LogStorage)
	var logFiles []models.LogFile

	for object := range s.minioClient.ListObjects(ctx, bucket.Name, gominio.ListObjectsOptions{
		Recursive: true,
	}) {
		if object.Err != nil {
			log.Error("error listing objects", "err", object.Err)
			continue
		}

		logFiles = append(logFiles, models.LogFile{
			Name:      filepath.Base(object.Key),
			Path:      object.Key,
			Size:      object.Size,
			UpdatedAt: object.LastModified,
		})
	}

	return logFiles, nil
}

// StoreChunk stores a file chunk with metadata
func (s *Storage) StoreChunk(ctx context.Context, chunk models.Chunk, reader io.Reader) error {
	bucket := s.storageOptions.GetBucketByCategory(ChunkStorage)
	objectName := fmt.Sprintf("%s/chunk_%s", chunk.JobID, chunk.ID)

	_, err := s.minioClient.PutObject(ctx, bucket.Name, objectName, reader, chunk.Size,
		gominio.PutObjectOptions{
			ContentType: string(bucket.Type),
			UserMetadata: map[string]string{
				"job_id":     chunk.JobID,
				"chunk_id":   chunk.ID,
				"file_name":  chunk.FileName,
				"start_byte": fmt.Sprintf("%d", chunk.StartByte),
				"end_byte":   fmt.Sprintf("%d", chunk.EndByte),
			},
		})
	if err != nil {
		log.Error("failed to store chunk", "chunk_id", chunk.ID, "err", err)
		return fmt.Errorf("failed to store chunk %s: %w", chunk.ID, err)
	}

	log.Info("stored chunk", "chunk_id", chunk.ID)
	return nil
}

// GetChunk retrieves a specific chunk
func (s *Storage) GetChunk(ctx context.Context, jobID, chunkID string) (*models.Chunk, io.ReadCloser, error) {
	bucket := s.storageOptions.GetBucketByCategory(ChunkStorage)
	objectName := fmt.Sprintf("%s/chunk_%s", jobID, chunkID)

	object, err := s.minioClient.GetObject(ctx, bucket.Name, objectName, gominio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chunk %s: %w", chunkID, err)
	}

	info, err := object.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chunk stats: %w", err)
	}

	chunk := &models.Chunk{
		ID:        chunkID,
		JobID:     jobID,
		FileName:  info.UserMetadata["file_name"],
		Size:      info.Size,
		CreatedAt: info.LastModified,
	}

	return chunk, object, nil
}

// StoreResult stores the grep results for a job
func (s *Storage) StoreResult(ctx context.Context, result models.Result) error {
	bucket := s.storageOptions.GetBucketByCategory(ResultStorage)
	objectName := fmt.Sprintf("%s/result_%s.json", result.JobID, result.ID)

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	_, err = s.minioClient.PutObject(ctx, bucket.Name, objectName, bytes.NewReader(data), int64(len(data)),
		gominio.PutObjectOptions{
			ContentType: "application/json",
			UserMetadata: map[string]string{
				"job_id":   result.JobID,
				"chunk_id": result.ChunkID,
			},
		})
	if err != nil {
		log.Error("failed to store result", "result_id", result.ID, "err", err)
		return fmt.Errorf("failed to store result %s: %w", result.ID, err)
	}

	log.Info("stored result", "result_id", result.ID)
	return nil
}

// GetJobResults retrieves all results for a specific job
func (s *Storage) GetJobResults(ctx context.Context, jobID string) ([]models.Result, error) {
	bucket := s.storageOptions.GetBucketByCategory(ResultStorage)
	prefix := fmt.Sprintf("%s/", jobID)

	var results []models.Result
	for object := range s.minioClient.ListObjects(ctx, bucket.Name, gominio.ListObjectsOptions{
		Prefix: prefix,
	}) {
		if object.Err != nil {
			log.Error("error listing results", "err", object.Err)
			continue
		}

		obj, err := s.minioClient.GetObject(ctx, bucket.Name, object.Key, gominio.GetObjectOptions{})
		if err != nil {
			log.Error("failed to get result object", "key", object.Key, "err", err)
			continue
		}

		var result models.Result
		if err := json.NewDecoder(obj).Decode(&result); err != nil {
			log.Error("failed to decode result", "key", object.Key, "err", err)
			continue
		}

		results = append(results, result)
	}

	return results, nil
}

// DeleteJob removes all data associated with a job
func (s *Storage) DeleteJob(ctx context.Context, jobID string) error {
	if err := s.deleteJobObjects(ctx, ChunkStorage, jobID); err != nil {
		return fmt.Errorf("failed to delete job chunks: %w", err)
	}

	if err := s.deleteJobObjects(ctx, ResultStorage, jobID); err != nil {
		return fmt.Errorf("failed to delete job results: %w", err)
	}

	return nil
}

// deleteJobObjects removes all objects with the given job ID prefix from the specified bucket
func (s *Storage) deleteJobObjects(ctx context.Context, category StorageCategory, jobID string) error {
	bucket := s.storageOptions.GetBucketByCategory(category)
	prefix := fmt.Sprintf("%s/", jobID)

	objectsCh := make(chan gominio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range s.minioClient.ListObjects(ctx, bucket.Name, gominio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Error("error listing objects for deletion", "err", object.Err)
				continue
			}
			objectsCh <- object
		}
	}()

	errorCh := s.minioClient.RemoveObjects(ctx, bucket.Name, objectsCh, gominio.RemoveObjectsOptions{})

	var deleteErrors []error
	for err := range errorCh {
		if err.Err != nil {
			log.Error("error deleting object", "object", err.ObjectName, "err", err.Err)
			deleteErrors = append(deleteErrors, fmt.Errorf("failed to delete %s: %w", err.ObjectName, err.Err))
		}
	}

	if len(deleteErrors) > 0 {
		return fmt.Errorf("failed to delete some objects: %v", deleteErrors)
	}

	return nil
}
