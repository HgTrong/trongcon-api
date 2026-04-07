package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"trongcon-api/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var ErrS3NotConfigured = errors.New("s3 not configured: set AWS_S3_BUCKET and AWS_REGION")

type UploadService struct {
	cfg    config.S3Config
	client *s3.Client
}

func NewUploadService(cfg config.S3Config) *UploadService {
	s := &UploadService{cfg: cfg}
	if cfg.Bucket == "" || cfg.Region == "" {
		return s
	}
	ctx := context.Background()
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return s
	}
	s.client = s3.NewFromConfig(awsCfg)
	return s
}

// Upload đưa file lên S3; folder là subpath dưới prefix (vd: categories, articles, common).
func (s *UploadService) Upload(ctx context.Context, folder, filename string, body io.Reader, contentType string) (string, error) {
	if s.client == nil {
		return "", ErrS3NotConfigured
	}
	ext := path.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	folder = strings.Trim(folder, "/")
	prefix := strings.Trim(s.cfg.Prefix, "/")
	key := folder + "/" + uuid.New().String() + ext
	if prefix != "" {
		key = prefix + "/" + key
	}
	ct := contentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(ct),
	})
	if err != nil {
		return "", err
	}
	return s.publicURL(key), nil
}

func (s *UploadService) publicURL(key string) string {
	if s.cfg.PublicBaseURL != "" {
		return strings.TrimRight(s.cfg.PublicBaseURL, "/") + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.cfg.Bucket, s.cfg.Region, key)
}
