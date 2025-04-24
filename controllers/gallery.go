package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"photo-booth.com/internal"
	"photo-booth.com/internal/models"
)

func GalleryHandler(w http.ResponseWriter, r *http.Request) {
	authenticated, _ := r.Context().Value(internal.AuthenticatedKey).(bool)
	userID, _ := r.Context().Value(internal.UserIDKey).(int)

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 20
	offset := (page - 1) * limit

	images, err := internal.GetImagesPaginated(userID, limit, offset)
	if err != nil {
		http.Error(w, "Unable to retrieve images", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(images)
		return
	}

	tmpl, err := template.ParseFiles("templates/gallery.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Images        []models.Image
		Authenticated bool
	}{
		Images:        images,
		Authenticated: authenticated,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}
