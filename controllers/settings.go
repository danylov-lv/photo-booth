package controllers

import (
	"net/http"
	"text/template"

	"golang.org/x/crypto/bcrypt"
	"photo-booth.com/internal"
	"photo-booth.com/internal/models"
)

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(internal.UserIDKey).(int)
	if !ok || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authenticated := r.Context().Value(internal.AuthenticatedKey).(bool)

	if r.Method == http.MethodGet {
		user, err := internal.GetUserByID(userID)
		if err != nil {
			http.Error(w, "Unable to load user data", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles("templates/settings.html")
		if err != nil {
			http.Error(w, "Unable to load settings page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			User          *models.User
			Authenticated bool
		}{
			User:          user,
			Authenticated: authenticated,
		})
		return
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		if username != "" || email != "" {
			err := internal.UpdateUserProfile(userID, username, email)
			if err != nil {
				http.Error(w, "Failed to update profile", http.StatusInternalServerError)
				return
			}
		}

		if currentPassword != "" && newPassword != "" && confirmPassword != "" {
			if newPassword != confirmPassword {
				http.Error(w, "Passwords do not match", http.StatusBadRequest)
				return
			}

			user, err := internal.GetUserByID(userID)
			if err != nil {
				http.Error(w, "Unable to load user data", http.StatusInternalServerError)
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
			if err != nil {
				http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
				return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				return
			}

			err = internal.UpdateUserPassword(userID, string(hashedPassword))
			if err != nil {
				http.Error(w, "Failed to update password", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/settings", http.StatusSeeOther)
	}
}
