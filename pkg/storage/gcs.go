package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type GCSUploader struct {
	client     *storage.Client
	bucketName string
}

func NewGCSUploader(bucketName string) (*GCSUploader, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCSUploader{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (g *GCSUploader) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	wc := g.client.Bucket(g.bucketName).Object(filename).NewWriter(ctx)
	wc.ContentType = "image/jpeg"
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	if _, err := io.Copy(wc, file); err != nil {
		return "", err
	}

	if err := wc.Close(); err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, filename)
	return url, nil
}

func (g *GCSUploader) Delete(ctx context.Context, filename string) error {
	return g.client.Bucket(g.bucketName).Object(filename).Delete(ctx)
}

func (g *GCSUploader) Close() error {
	return g.client.Close()
}
