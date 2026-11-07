package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// sha256MetaKey carries the application SHA-256 checksum as S3 object metadata
// (x-amz-meta-sha256). S3 ETags are MD5-based and cannot serve this purpose.
const sha256MetaKey = "sha256"

// S3Config selects the object-store endpoint. An empty Endpoint targets real
// AWS S3; a custom Endpoint (plus UsePathStyle) targets S3-compatible stores
// such as LocalStack. Empty credentials select the SDK default chain
// (env, shared config, EC2/ECS IAM role), which is the production path.
type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
	HTTPClient      *http.Client
}

// S3ObjectStorage implements ObjectStorage against AWS S3 or any
// S3-compatible store (LocalStack, MinIO) via aws-sdk-go-v2.
type S3ObjectStorage struct {
	client *s3.Client
}

func NewS3ObjectStorage(ctx context.Context, cfg S3Config) (*S3ObjectStorage, error) {
	region := cfg.Region
	if region == "" {
		region = "ap-northeast-2"
	}
	opts := []func(*config.LoadOptions) error{config.WithRegion(region)}
	if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	if cfg.HTTPClient != nil {
		opts = append(opts, config.WithHTTPClient(cfg.HTTPClient))
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})
	return &S3ObjectStorage{client: client}, nil
}

func isS3NotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "NoSuchBucket":
			return true
		}
	}
	return false
}

func strptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strval(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetObject fetches bytes, system headers, the sha256 metadata, and tags.
func (s *S3ObjectStorage) GetObject(ctx context.Context, bucket, key string) (*StorageObject, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
	if err != nil {
		if isS3NotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("s3 GetObject failed: %w", err)
	}
	defer out.Body.Close()
	content, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("s3 GetObject body read failed: %w", err)
	}
	tags := map[string]string{}
	tagOut, err := s.client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{Bucket: &bucket, Key: &key})
	if err != nil {
		if isS3NotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("s3 GetObjectTagging failed: %w", err)
	}
	for _, t := range tagOut.TagSet {
		tags[strval(t.Key)] = strval(t.Value)
	}
	obj := &StorageObject{
		Key:                key,
		Content:            content,
		ContentType:        strval(out.ContentType),
		CacheControl:       strval(out.CacheControl),
		ContentDisposition: strval(out.ContentDisposition),
		Tags:               tags,
	}
	if out.Metadata != nil {
		obj.SHA256 = out.Metadata[sha256MetaKey]
	}
	if out.LastModified != nil {
		obj.CreatedAt = *out.LastModified
	}
	return obj, nil
}

// PutObject stores bytes with system headers, sha256 metadata, and tags.
func (s *S3ObjectStorage) PutObject(ctx context.Context, bucket string, obj *StorageObject) error {
	if obj == nil {
		return fmt.Errorf("s3 PutObject requires a non-nil object")
	}
	in := &s3.PutObjectInput{
		Bucket:             &bucket,
		Key:                &obj.Key,
		Body:               bytes.NewReader(obj.Content),
		ContentType:        strptr(obj.ContentType),
		CacheControl:       strptr(obj.CacheControl),
		ContentDisposition: strptr(obj.ContentDisposition),
	}
	if obj.SHA256 != "" {
		in.Metadata = map[string]string{sha256MetaKey: obj.SHA256}
	}
	if len(obj.Tags) > 0 {
		values := url.Values{}
		for k, v := range obj.Tags {
			values.Set(k, v)
		}
		in.Tagging = aws.String(values.Encode())
	}
	if _, err := s.client.PutObject(ctx, in); err != nil {
		return fmt.Errorf("s3 PutObject failed: %w", err)
	}
	return nil
}

// DeleteObject removes an object. Like S3 itself, deleting a missing key is
// a no-op success, matching InMemoryObjectStorage semantics.
func (s *S3ObjectStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &bucket, Key: &key}); err != nil {
		return fmt.Errorf("s3 DeleteObject failed: %w", err)
	}
	return nil
}

// TagObject replaces the tag set of an existing object.
func (s *S3ObjectStorage) TagObject(ctx context.Context, bucket, key string, tags map[string]string) error {
	tagSet := make([]types.Tag, 0, len(tags))
	for k, v := range tags {
		k, v := k, v
		tagSet = append(tagSet, types.Tag{Key: &k, Value: &v})
	}
	if _, err := s.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket: &bucket, Key: &key, Tagging: &types.Tagging{TagSet: tagSet},
	}); err != nil {
		if isS3NotFound(err) {
			return ErrObjectNotFound
		}
		return fmt.Errorf("s3 PutObjectTagging failed: %w", err)
	}
	return nil
}
