package handlers

import (
	"context"
	"expense-tracker/firebase"
	"io"
	"net/http"
	"os"
)

func BackupData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if firebase.Bucket == nil {
		http.Error(w, "Failed to access storage", http.StatusInternalServerError)
		return
	}

	backupFile, err := os.Open("backup.json")
	if err != nil {
		http.Error(w, "Backup file not found", http.StatusNotFound)
		return
	}
	defer backupFile.Close()

	obj := firebase.Bucket.Object("backups/backup.json")
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	_, err = io.Copy(writer, backupFile)
	if err != nil {
		http.Error(w, "Backup failed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Backup successful"))
}
