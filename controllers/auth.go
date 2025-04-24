package controllers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/crypto/bcrypt"
	"photo-booth.com/internal"
	"photo-booth.com/internal/models"
	"photo-booth.com/internal/utils"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		authenticated := r.Context().Value(internal.AuthenticatedKey).(bool)

		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			http.Error(w, "Unable to load register page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			Authenticated bool
		}{Authenticated: authenticated})
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		user := models.User{
			Username:          username,
			Email:             email,
			Password:          string(hashedPassword),
			ConfirmationToken: utils.GenerateToken(),
			IsConfirmed:       false,
		}

		if err := internal.CreateUser(&user); err != nil {
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		go utils.SendConfirmationEmail(user.Email, user.ConfirmationToken)

		http.Redirect(w, r, "/confirm_account", http.StatusSeeOther)
	}
}

func ConfirmAccountHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	if err := internal.ConfirmUser(token); err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		authenticated := r.Context().Value(internal.AuthenticatedKey).(bool)

		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, "Unable to load login page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			Authenticated bool
		}{Authenticated: authenticated})
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		storedUser, err := internal.GetUserByUsername(username)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(password)) != nil {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		if !storedUser.IsConfirmed {
			http.Error(w, "Account not confirmed", http.StatusForbidden)
			return
		}

		session, _ := internal.Store.Get(r, "session")
		session.Values["authenticated"] = true
		session.Values["user_id"] = storedUser.ID
		if err := session.Save(r, w); err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Unable to save session", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/gallery", http.StatusSeeOther)
	}
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("templates/reset_password.html")
		if err != nil {
			http.Error(w, "Unable to load reset password page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	} else if r.Method == http.MethodPost {
		email := r.FormValue("email")
		if email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		user, err := internal.GetUserByEmail(email)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		token := utils.GenerateToken()
		expiryStr := os.Getenv("RESET_TOKEN_EXPIRY")
		expirySeconds, err := strconv.Atoi(expiryStr)
		if err != nil || expirySeconds <= 0 {
			expirySeconds = 3600
		}
		expiry := time.Now().Add(time.Duration(expirySeconds) * time.Second)

		if err := internal.SavePasswordResetToken(user.ID, token, expiry); err != nil {
			http.Error(w, "Failed to save reset token", http.StatusInternalServerError)
			return
		}

		go utils.SendPasswordResetEmail(user.Email, token)

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Invalid or missing token", http.StatusBadRequest)
			return
		}

		authenticated := r.Context().Value(internal.AuthenticatedKey).(bool)

		tmpl, err := template.ParseFiles("templates/change_password.html")
		if err != nil {
			http.Error(w, "Unable to load change password page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, struct {
			Token         string
			Authenticated bool
		}{
			Token:         token,
			Authenticated: authenticated,
		})
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		token := r.FormValue("token")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		if newPassword != confirmPassword {
			http.Error(w, "Passwords do not match", http.StatusBadRequest)
			return
		}

		user, err := internal.GetUserByResetToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusBadRequest)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		if err := internal.UpdateUserPassword(user.ID, string(hashedPassword)); err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := internal.Store.Get(r, "session")
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
