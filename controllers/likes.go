package controllers

import (
	"net/http"

	"photo-booth.com/internal"
)

func LikeImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	imageID := r.FormValue("image_id")
	userID, ok := r.Context().Value(internal.UserIDKey).(int)
	if imageID == "" || !ok || userID == 0 {
		http.Error(w, "Image ID and User ID are required", http.StatusBadRequest)
		return
	}

	err := internal.AddLike(userID, imageID)
	if err != nil {
		if err.Error() == "like already exists" {
			http.Error(w, "You have already liked this image", http.StatusConflict)
			return
		}
		http.Error(w, "Unable to like image", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/gallery", http.StatusSeeOther)
}
