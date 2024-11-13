package tests

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swarit-pandey/distributed-grep/common/minio"
	"github.com/swarit-pandey/distributed-grep/common/models"
)

func TestMinIOIntegration(t *testing.T) {
	ctx := context.Background()
	container, err := SetupMinio(ctx)
	require.NoError(t, err)
	defer container.Terminate(ctx)

	opts := minio.StorageOptions{
		Buckets: []minio.BucketOptions{
			{
				Name:     "logs",
				Type:     minio.TextType,
				Category: minio.LogStorage,
			},
			{
				Name:     "chunks",
				Type:     minio.TextType,
				Category: minio.ChunkStorage,
			},
			{
				Name:     "results",
				Type:     minio.JSONType,
				Category: minio.ResultStorage,
			},
		},
	}

	endpoint, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	storage, err := minio.New(endpoint, container.Username, container.Password, false, nil, &opts)
	require.NoError(t, err)

	err = storage.Instantiate(ctx)
	require.NoError(t, err)
}

func TestLogFileOperations(t *testing.T) {
	ctx := context.Background()

	opts := &minio.StorageOptions{
		Buckets: []minio.BucketOptions{
			{
				Name:     "test-logs",
				Type:     minio.TextType,
				Category: minio.LogStorage,
			},
		},
	}

	storage, err := initTestStorage(t, opts)
	require.NoError(t, err)

	t.Run("UploadAndListLogFiles", func(t *testing.T) {
		sampleLogs := map[string]string{
			"service1/app.log": createSampleLogs("service1", 2000),
			"service2/app.log": createSampleLogs("service2", 1000),
			"service3/app.log": createSampleLogs("service3", 500),
		}

		expectedFiles := make(map[string]int64)

		for path, content := range sampleLogs {
			logModel := models.LogFile{
				Name:      path,
				Path:      path,
				Size:      int64(len(content)),
				UpdatedAt: time.Now(),
			}

			err := storage.UploadLogFile(ctx, logModel, strings.NewReader(content))
			require.NoError(t, err)

			expectedFiles[path] = int64(len(content))
			t.Logf("log file uploaded %s", path)
		}

		results, err := storage.ListLogFiles(ctx)
		require.NoError(t, err)

		assert.Equal(t, len(sampleLogs), len(results), "number of files should match")

		resultMap := make(map[string]models.LogFile)
		for _, file := range results {
			resultMap[file.Path] = file
		}

		for expectedPath, expectedSize := range expectedFiles {
			result, exists := resultMap[expectedPath]

			assert.True(t, exists, "file %s should exist", expectedPath)

			if exists {
				assert.Equal(t, expectedPath, result.Path, "file path should match")
				assert.Equal(t, filepath.Base(expectedPath), result.Name, "file name should match")
				assert.Equal(t, expectedSize, result.Size, "file size should match for %s", expectedPath)
				assert.False(t, result.UpdatedAt.IsZero(), "UpdatedAt should be set")
			}
		}
	})
}

func createSampleLogs(service string, lines int) string {
	var sb strings.Builder
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	messages := []string{"message1", "message2", "message3", "message4", "message5"}

	for i := 0; i <= lines; i++ {
		level := levels[i%len(levels)]
		message := messages[i%len(messages)]
		timestamp := time.Now()
		sb.WriteString(fmt.Sprintf("%s [%s] %s: %s\n", timestamp, service, level, message))
	}

	return sb.String()
}

func TestChunkOperations(t *testing.T) {
	ctx := context.Background()

	opts := &minio.StorageOptions{
		Buckets: []minio.BucketOptions{
			{
				Name:     "test-chunks",
				Type:     minio.TextType,
				Category: minio.ChunkStorage,
			},
		},
	}

	storage, err := initTestStorage(t, opts)
	require.NoError(t, err)

	t.Run("StoreAndRetrieveChunk", func(t *testing.T) {
		content := "This is chunk content for testing"
		contentSize := int64(len(content))

		chunk := models.Chunk{
			ID:        "chunk1",
			JobID:     "job123",
			FileName:  "service1/app.log",
			StartByte: 0,
			EndByte:   contentSize,
			Size:      contentSize,
			CreatedAt: time.Now(),
		}

		err := storage.StoreChunk(ctx, chunk, strings.NewReader(content))
		require.NoError(t, err, "should store chunk without error")

		retrievedChunk, reader, err := storage.GetChunk(ctx, chunk.JobID, chunk.ID)
		require.NoError(t, err, "should retrieve chunk without error")
		defer reader.Close()

		assert.Equal(t, chunk.ID, retrievedChunk.ID, "chunk ID should match")
		assert.Equal(t, chunk.JobID, retrievedChunk.JobID, "job ID should match")
		assert.Equal(t, chunk.FileName, retrievedChunk.FileName, "filename should match")
		assert.Equal(t, contentSize, retrievedChunk.Size, "size should match")

		retrievedContent, err := io.ReadAll(reader)
		require.NoError(t, err, "should read chunk content")
		assert.Equal(t, content, string(retrievedContent), "chunk content should match")
	})

	t.Run("GetAllChunks", func(t *testing.T) {
		chunks := []struct {
			chunk   models.Chunk
			content string
		}{
			{
				content: "First chunk content",
				chunk: models.Chunk{
					ID:        "chunk1",
					JobID:     "job456",
					FileName:  "service1/app.log",
					CreatedAt: time.Now(),
				},
			},
			{
				content: "Second chunk content",
				chunk: models.Chunk{
					ID:        "chunk2",
					JobID:     "job456",
					FileName:  "service1/app.log",
					CreatedAt: time.Now(),
				},
			},
		}

		for i := range chunks {
			chunks[i].chunk.Size = int64(len(chunks[i].content))
			chunks[i].chunk.EndByte = chunks[i].chunk.Size

			err := storage.StoreChunk(ctx, chunks[i].chunk, strings.NewReader(chunks[i].content))
			require.NoError(t, err, "should store chunk without error")
		}

		for i := range chunks {
			_, reader, err := storage.GetChunk(ctx, chunks[i].chunk.JobID, chunks[i].chunk.ID)
			require.NoError(t, err)
			retrievedContent, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, chunks[i].content, string(retrievedContent))
			t.Logf("actual content got: %s", string(retrievedContent))
		}
	})

	t.Run("NonExistentChunk", func(t *testing.T) {
		_, _, err := storage.GetChunk(ctx, "nonexistent", "chunk")
		assert.Error(t, err, "should return error for non-existent chunk")
		t.Logf("Expected error for non-existent chunk: %v", err)
	})
}

func initTestStorage(t *testing.T, opts *minio.StorageOptions) (*minio.Storage, error) {
	minioContainer, err := SetupMinio(context.Background())
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		if err := minioContainer.Container.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	connection, err := minioContainer.ConnectionString(context.Background())
	if err != nil {
		t.Logf("failed to setup minio test container: %v", err)
		t.FailNow()
	}

	storage, err := minio.New(
		connection,
		minioContainer.Username,
		minioContainer.Password,
		false,
		nil,
		opts,
	)
	if err != nil {
		return nil, err
	}

	err = storage.Instantiate(context.Background())
	if err != nil {
		return nil, err
	}

	return storage, nil
}
