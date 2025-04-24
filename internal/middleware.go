package internal

import (
	"context"
	"log"
	"net/http"

	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var Store *sessions.CookieStore

type contextKey string

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET is not set in .env file")
	}

	Store = sessions.NewCookieStore([]byte(os.Getenv("JWT_SECRET")))
}

const UserIDKey contextKey = "user_id"
const AuthenticatedKey contextKey = "authenticated"

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session")
		if session.Values["authenticated"] != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, ok := session.Values["user_id"].(int)
		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "session")
		authenticated := session.Values["authenticated"] == true
		userID, _ := session.Values["user_id"].(int)

		ctx := context.WithValue(r.Context(), AuthenticatedKey, authenticated)
		ctx = context.WithValue(ctx, UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
