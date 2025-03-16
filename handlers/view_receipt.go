package handlers

import (
	"log"
	"net/http"
)

// View the receipts serve the receipt image
func ViewReceipt(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	log.Println("Serving file:", "uploads/"+filePath)
	http.ServeFile(w, r, "uploads/"+filePath)
}
