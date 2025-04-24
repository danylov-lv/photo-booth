package controllers

import (
	"net/http"
	"os"
	"strconv"

	"photo-booth.com/internal"
)

func DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(internal.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	imageIDStr := r.FormValue("image_id")
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	image, err := internal.GetImageByID(imageID)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	if image.UserID != userID {
		http.Error(w, "You are not authorized to delete this image", http.StatusForbidden)
		return
	}

	err = os.Remove(image.FilePath)
	if err != nil {
		http.Error(w, "Failed to delete image file", http.StatusInternalServerError)
		return
	}

	err = internal.DeleteImageByID(imageID)
	if err != nil {
		http.Error(w, "Failed to delete image record", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/gallery", http.StatusSeeOther)
}
