package oss

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Client represents an OSS client
type Client struct {
	client *oss.Client
	config *Config
}

// Config holds the OSS client configuration
type Config struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	SecurityToken   string // Optional
}

// NewClient creates a new OSS client
func NewClient(config *Config) (*Client, error) {
	// Create OSS client options
	var options []oss.ClientOption
	if config.SecurityToken != "" {
		options = append(options, oss.SecurityToken(config.SecurityToken))
	}

	// Create OSS client
	client, err := oss.New(
		config.Endpoint,
		config.AccessKeyID,
		config.AccessKeySecret,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// ListBuckets lists all buckets
func (c *Client) ListBuckets() ([]string, error) {
	// List buckets
	result, err := c.client.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	// Extract bucket names
	buckets := make([]string, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		buckets = append(buckets, bucket.Name)
	}

	return buckets, nil
}

// GetBucket gets a bucket by name
func (c *Client) GetBucket(name string) (*Bucket, error) {
	// Get bucket
	bucket, err := c.client.Bucket(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket %s: %w", name, err)
	}

	return &Bucket{
		bucket: bucket,
		name:   name,
	}, nil
}

// Bucket represents an OSS bucket
type Bucket struct {
	bucket *oss.Bucket
	name   string
}

// Name returns the bucket name
func (b *Bucket) Name() string {
	return b.name
}

// ListObjects lists objects in the bucket
func (b *Bucket) ListObjects(prefix string) ([]string, error) {
	// List objects
	result, err := b.bucket.ListObjects(oss.Prefix(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
	}

	// Extract object keys
	objects := make([]string, 0, len(result.Objects))
	for _, object := range result.Objects {
		objects = append(objects, object.Key)
	}

	return objects, nil
}

// PutObject uploads an object
func (b *Bucket) PutObject(key string, data []byte) error {
	// This is a placeholder implementation
	// The full implementation will be done in Task 9
	return fmt.Errorf("not implemented")
}

// GetObject downloads an object
func (b *Bucket) GetObject(key string) ([]byte, error) {
	// This is a placeholder implementation
	// The full implementation will be done in Task 9
	return nil, fmt.Errorf("not implemented")
}

// DeleteObject deletes an object
func (b *Bucket) DeleteObject(key string) error {
	// This is a placeholder implementation
	// The full implementation will be done in Task 9
	return fmt.Errorf("not implemented")
}
