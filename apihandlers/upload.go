package apihandlers

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"os"
	"time"
)

func getFirebaseBucket() (*storage.BucketHandle, error) {
	credentials := []byte(os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY"))
	opt := option.WithCredentialsJSON(credentials)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	client, err := storage.NewClient(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}
	bucketName := "expense-tracker-43fbe.appspot.com"
	bucket := client.Bucket(bucketName)
	return bucket, nil
}

// UploadFileHandler handles file upload from the web
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload File to Firebase Storage
	bucket, err := getFirebaseBucket()
	if err != nil {
		http.Error(w, "Firebase error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new file in database storage
	objectName := "receipts/" + handler.Filename
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	wc := bucket.Object(objectName).NewWriter(ctx)

	// Copy file to firebase
	if _, err := io.Copy(wc, file); err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Close Writer
	if err := wc.Close(); err != nil {
		http.Error(w, "Error finalizing upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Generate public URL
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket.Object(objectName).BucketName(), objectName)

	// Return success response with the file upload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"url": "` + url + `"}`))
}
