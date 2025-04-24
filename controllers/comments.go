package controllers

import (
	"net/http"
	"strconv"

	"photo-booth.com/internal"
	"photo-booth.com/internal/utils"
)

func AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	imageIDStr := r.FormValue("image_id")
	content := r.FormValue("content")
	userID, ok := r.Context().Value(internal.UserIDKey).(int)

	if !ok || userID == 0 || imageIDStr == "" || content == "" {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	err = internal.AddComment(imageID, userID, content)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	author, err := internal.GetImageAuthor(imageID)
	if err != nil {
		http.Error(w, "Failed to get image author", http.StatusInternalServerError)
		return
	}

	if author.NotifyOnComment {
		go utils.SendCommentNotification(author.Email, content)
	}

	http.Redirect(w, r, "/gallery", http.StatusSeeOther)
}
