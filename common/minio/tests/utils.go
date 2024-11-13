package tests

import (
	"context"
	"fmt"

	testminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

func SetupMinio(ctx context.Context) (*testminio.MinioContainer, error) {
	minioContainer, err := testminio.Run(ctx, "minio/minio:RELEASE.2024-11-07T00-52-20Z.fips")
	if err != nil {
		return nil, fmt.Errorf("failed to setup minio test container")
	}

	return minioContainer, nil
}
