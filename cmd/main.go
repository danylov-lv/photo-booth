package main

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"photo-booth.com/controllers"
	"photo-booth.com/internal"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	if _, err := os.Stat("data"); os.IsNotExist(err) {
		if err := os.Mkdir("data", os.ModePerm); err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
	}

	internal.InitDB("data/photo-booth.db")
	defer internal.DB.Close()

	mux := http.NewServeMux()

	wrappedMux := internal.AuthMiddleware(mux)

	sfs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", sfs))

	ufs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", ufs))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authenticated, _ := r.Context().Value(internal.AuthenticatedKey).(bool)

		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Unable to load index page", http.StatusInternalServerError)
			return
		}

		data := struct {
			Authenticated bool
		}{
			Authenticated: authenticated,
		}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Unable to render template", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/register", controllers.RegisterHandler)
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/gallery", controllers.GalleryHandler)
	mux.HandleFunc("/camera", internal.RequireAuth(controllers.CameraHandler))
	mux.HandleFunc("/comments/add", internal.RequireAuth(controllers.AddComment))
	mux.HandleFunc("/like", internal.RequireAuth(controllers.LikeImageHandler))
	mux.HandleFunc("/password/reset", controllers.ResetPasswordHandler)
	mux.HandleFunc("/password/change", controllers.ChangePasswordHandler)
	mux.HandleFunc("/confirm", controllers.ConfirmAccountHandler)
	mux.HandleFunc("/confirm_account", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/confirm_account.html")
		if err != nil {
			http.Error(w, "Unable to load confirmation page", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	})
	mux.HandleFunc("/logout", internal.RequireAuth(controllers.LogoutHandler))
	mux.HandleFunc("/images/delete", internal.RequireAuth(controllers.DeleteImageHandler))
	mux.HandleFunc("/settings", internal.RequireAuth(controllers.SettingsHandler))

	log.Printf("Starting server on %s", port)
	if err := http.ListenAndServe(port, wrappedMux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
