package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"cloud.google.com/go/storage"
)

type Service interface {
	UploadToGCS(ctx context.Context, r io.Reader, objectName string, public bool, contentType string) (string, error)
	GetPublicURL(objectName string) string
	DeleteImage(objectName string) error
}

type Client struct {
	projectID  string
	bucketName string
	client     *storage.Client
}

func (gcs *Client) UploadToGCS(ctx context.Context, r io.Reader, objectName string, public bool, contentType string) (string, error) {
	bh := gcs.client.Bucket(gcs.bucketName)
	// Next check if the bucket exists
	if _, err := bh.Attrs(ctx); err != nil {
		return "", err
	}

	obj := bh.Object(objectName)
	w := obj.NewWriter(ctx)

	// Set the Content-Type header
	if contentType != "" {
		w.ContentType = contentType
	}

	if _, err := io.Copy(w, r); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	if public {
		if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			return "", err
		}
	}
	// Return the public URL
	return gcs.GetPublicURL(objectName), nil
}

func (gcs *Client) GetPublicURL(objectName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", gcs.bucketName, objectName)
}

// readFile reads the named file in Google Cloud Storage.
func (gcs *Client) ReadFile(ctx context.Context, fileName string) {
	bh := gcs.client.Bucket(gcs.bucketName)
	// Next check if the bucket exists
	if _, err := bh.Attrs(ctx); err != nil {
		// return nil, nil, err
	}

	obj := bh.Object(fileName)
	w := obj.NewWriter(ctx)
	io.WriteString(w, "\nAbbreviated file content (first line and last 1K):\n")

	rc, err := bh.Object(fileName).NewReader(ctx)
	if err != nil {
		log.Println(err)
	}

	defer rc.Close()
	slurp, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Println(err)
	}
	log.Println(w, "%s\n", bytes.SplitN(slurp, []byte("\n"), 2)[0])
	if len(slurp) > 1024 {
		log.Println(w, "...%s\n", slurp[len(slurp)-1024:])
	} else {
		log.Println(w, "%s\n", slurp)
	}
}

func InitialGCSClient(projectID, bucketName string, client *storage.Client) *Client {

	return &Client{
		projectID:  projectID,
		bucketName: bucketName,
		client:     client,
	}
}

func (gcs *Client) DeleteImage(objectName string) error {
	// Create a context
	ctx := context.Background()

	// Create a storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create a handle to the object (image)
	object := client.Bucket(gcs.bucketName).Object(objectName)

	// Delete the object
	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object %q: %v", objectName, err)
	}

	fmt.Printf("Successfully deleted object %s from bucket %s\n", objectName, gcs.bucketName)
	return nil
}
