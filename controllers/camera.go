package controllers

import (
	"html/template"
	"net/http"
	"os"

	"photo-booth.com/internal"
	"photo-booth.com/internal/models"
)

func CameraHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		files, err := os.ReadDir("static/img/overlays")
		if err != nil {
			http.Error(w, "Unable to load overlays", http.StatusInternalServerError)
			return
		}

		var overlays []string
		for _, file := range files {
			if !file.IsDir() {
				overlays = append(overlays, file.Name())
			}
		}

		authenticated := r.Context().Value(internal.AuthenticatedKey).(bool)
		userID, ok := r.Context().Value(internal.UserIDKey).(int)

		if !ok || userID == 0 {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}

		recentImages, err := internal.GetRecentImagesByUser(userID, 5)
		if err != nil {
			http.Error(w, "Unable to load recent images", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/camera.html")
		if err != nil {
			http.Error(w, "Unable to load camera page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			Overlays      []string
			Authenticated bool
			RecentImages  []models.Image
		}{Overlays: overlays, Authenticated: authenticated, RecentImages: recentImages})
		return
	}

	if r.Method == http.MethodPost {
		imageData := r.FormValue("image")
		if imageData == "" {
			http.Error(w, "No image data provided", http.StatusBadRequest)
			return
		}

		filePath, err := internal.SaveImageFromBase64(imageData)
		if err != nil {
			http.Error(w, "Unable to save image", http.StatusInternalServerError)
			return
		}

		session, _ := internal.Store.Get(r, "session")
		userID := session.Values["user_id"].(int)
		if err := internal.SaveImageInfo(filePath, userID); err != nil {
			http.Error(w, "Unable to save image info", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	}
}
