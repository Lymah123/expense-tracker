package storage

import (
	"expense-tracker/firebase"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func UploadReceipt(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("receipt")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	uploadDir := "uploads/"
	os.MkdirAll(uploadDir, os.ModePerm)

	filePath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	fmt.Fprintf(w, "File uploaded successfully: %s", filePath)
}

// SaveReceipt uploads the receipt file to Firebase Storage
func SaveReceipt(localPath string) error {
	fileName := filepath.Base(localPath)
	err := firebase.UploadFile(localPath, "/receipts"+fileName)
	if nil != err {
		return fmt.Errorf("error uploading file: %v", err)
	}
	fmt.Println("Receipt uploaded:", fileName)
	return nil
}
