package minio

import "github.com/swarit-pandey/distributed-grep/common/logger"

// Setting up logger
var log *logger.Logger

func InitLogger(l *logger.Logger) {
	if l != nil {
		log = l
	} else {
		log = logger.New()
	}
}

// MinIOClient represents core MinIO client
type MinOptions struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyId"`
	SecretAccessKey string `mapstructure:"secretAccessKey"`
	SSL             bool   `mapstructure:"ssl"`
}
