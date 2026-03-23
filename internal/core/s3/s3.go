package s3

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/radni/soapbox/internal/core/config"
)

// S3Client implements Client using the AWS SDK v2 (compatible with MinIO).
type S3Client struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
	baseURL string
}

// New creates an S3Client from the application's S3 configuration.
func New(cfg config.S3Config) *S3Client {
	awsCfg := aws.Config{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	baseURL := strings.TrimRight(cfg.Endpoint, "/") + "/" + cfg.Bucket

	return &S3Client{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
		baseURL: baseURL,
	}
}

func (c *S3Client) PresignPutObject(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	req, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", fmt.Errorf("s3: presign put: %w", err)
	}

	return req.URL, nil
}

func (c *S3Client) ObjectURL(key string) string {
	return c.baseURL + "/" + key
}

func (c *S3Client) EnsureBucket(ctx context.Context) error {
	_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err == nil {
		return nil
	}

	var notFound *s3types.NotFound
	if !errors.As(err, &notFound) {
		return fmt.Errorf("s3: head bucket: %w", err)
	}

	_, err = c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		return fmt.Errorf("s3: create bucket: %w", err)
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetObject",
			"Resource": "arn:aws:s3:::%s/*"
		}]
	}`, c.bucket)

	_, err = c.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(c.bucket),
		Policy: aws.String(policy),
	})
	if err != nil {
		return fmt.Errorf("s3: set bucket policy: %w", err)
	}

	return nil
}
