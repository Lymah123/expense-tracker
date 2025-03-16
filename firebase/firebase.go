package firebase

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	Bucket     *storage.BucketHandle
	AuthClient *auth.Client
	App        *firebase.App
)

// Initialize Firebase app, Auth client, and Storage
func InitFirebase() error {
	// Load Firebase service account key from environment variable
	firebaseKey := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if firebaseKey == "" {
		return fmt.Errorf("FIREBASE_SERVICE_ACCOUNT_KEY environment variable is not set")
	}

	opt := option.WithCredentialsJSON([]byte(firebaseKey))
	var err error
	App, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing firebase app: %v", err)
	}

	// Initialize Auth client
	AuthClient, err = App.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("error initializing auth client: %v", err)
	}

	// Initialize Storage client
	client, err := App.Storage(context.Background())
	if err != nil {
		return fmt.Errorf("error getting storage client: %v", err)
	}

	Bucket, err = client.Bucket("expense-tracker-43fbe.appspot.com")
	if err != nil {
		return fmt.Errorf("error getting bucket handle: %v", err)
	}

	return nil
}

// UploadFile uploads a file to Firebase Storage
func UploadFile(filePath string, objectName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer f.Close()

	obj := Bucket.Object(objectName)
	wc := obj.NewWriter(ctx)

	if _, err := io.Copy(wc, f); err != nil {
		return fmt.Errorf("error copying file to storage: %v", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("error closing writer: %v", err)
	}

	fmt.Println("File uploaded successfully:", objectName)
	return nil
}
