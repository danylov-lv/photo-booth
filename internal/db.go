package internal

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"photo-booth.com/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTables()
}

func createTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		confirmation_token TEXT,
		is_confirmed BOOLEAN DEFAULT FALSE,
		reset_token TEXT,
		reset_token_expiry DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		notify_on_comment BOOLEAN DEFAULT TRUE
	);`

	imagesTable := `CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		file_path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	commentsTable := `CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		image_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (image_id) REFERENCES images(id),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	likesTable := `CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		image_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (image_id) REFERENCES images(id)
	);`

	uniqueLikeIndex := `CREATE UNIQUE INDEX IF NOT EXISTS unique_like ON likes (user_id, image_id);`

	_, err := DB.Exec(usersTable)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	_, err = DB.Exec(imagesTable)
	if err != nil {
		log.Fatalf("Failed to create images table: %v", err)
	}

	_, err = DB.Exec(commentsTable)
	if err != nil {
		log.Fatalf("Failed to create comments table: %v", err)
	}

	_, err = DB.Exec(likesTable)
	if err != nil {
		log.Fatalf("Failed to create likes table: %v", err)
	}

	_, err = DB.Exec(uniqueLikeIndex)
	if err != nil {
		log.Fatalf("Failed to create unique index on likes table: %v", err)
	}
}

func GetImages(userID int) ([]models.Image, error) {
	query := `
        SELECT 
            images.id, 
            images.user_id, 
            images.file_path, 
            images.created_at,
            (SELECT COUNT(*) FROM likes WHERE likes.image_id = images.id) AS likes_count,
			images.user_id = ? AS is_owner
        FROM images
        ORDER BY images.created_at DESC
    `

	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("Error fetching images: %v", err)
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(&image.ID, &image.UserID, &image.FilePath, &image.CreatedAt, &image.Likes, &image.IsOwner); err != nil {
			log.Printf("Error scanning image row: %v", err)
			return nil, err
		}

		comments, err := GetCommentsByImageID(image.ID)
		if err != nil {
			log.Printf("Error fetching comments for image %d: %v", image.ID, err)
			return nil, err
		}
		image.Comments = comments

		images = append(images, image)
	}

	return images, nil
}

func GetCommentsByImageID(imageID int) ([]models.Comment, error) {
	query := `
        SELECT 
            comments.id, 
            comments.image_id, 
            comments.user_id, 
            users.username, 
            comments.content, 
            comments.created_at
        FROM comments
        JOIN users ON comments.user_id = users.id
        WHERE comments.image_id = ?
        ORDER BY comments.created_at ASC
    `

	rows, err := DB.Query(query, imageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []models.Comment{}
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.ImageID, &comment.UserID, &comment.Username, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (username, email, password, confirmation_token, is_confirmed) VALUES (?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, user.Username, user.Email, user.Password, user.ConfirmationToken, user.IsConfirmed)
	return err
}

func GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, password, is_confirmed FROM users WHERE username = ?`
	row := DB.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsConfirmed)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password, is_confirmed FROM users WHERE email = ?`
	row := DB.QueryRow(query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsConfirmed)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByID(userID int) (*models.User, error) {
	query := `SELECT id, username, email, password, is_confirmed FROM users WHERE id = ?`
	row := DB.QueryRow(query, userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsConfirmed)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(userID int, username, email, password string) error {
	query := `UPDATE users SET username = ?, email = ?, password = ? WHERE id = ?`
	_, err := DB.Exec(query, username, email, password, userID)
	return err
}

func UpdateUserProfile(userID int, username, email string) error {
	query := `UPDATE users SET username = COALESCE(NULLIF(?, ''), username), email = COALESCE(NULLIF(?, ''), email) WHERE id = ?`
	_, err := DB.Exec(query, username, email, userID)
	return err
}

func SaveImage(file io.Reader) (string, error) {
	fileName := fmt.Sprintf("image_%d.jpg", time.Now().UnixNano())
	filePath := filepath.Join("static", "images", fileName)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	return filePath, nil
}

func SaveImageFromBase64(data string) (string, error) {
	parts := strings.Split(data, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid image data")
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("uploads/photo_%d.png", time.Now().UnixNano())
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write(decoded); err != nil {
		return "", err
	}

	return fileName, nil
}

func SaveImageInfo(filePath string, userID int) error {
	query := `INSERT INTO images (user_id, file_path, created_at) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, userID, filePath, time.Now())
	return err
}

func ConfirmUser(token string) error {
	query := `UPDATE users SET is_confirmed = 1, confirmation_token = NULL WHERE confirmation_token = ?`
	result, err := DB.Exec(query, token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("invalid or expired token")
	}

	return nil
}

func SavePasswordResetToken(userID int, token string, expiry time.Time) error {
	query := `UPDATE users SET reset_token = ?, reset_token_expiry = ? WHERE id = ?`
	_, err := DB.Exec(query, token, expiry, userID)
	return err
}

func GetUserByResetToken(token string) (*models.User, error) {
	query := `SELECT id, username, email, password, is_confirmed FROM users WHERE reset_token = ? AND reset_token_expiry > ?`
	row := DB.QueryRow(query, token, time.Now())

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsConfirmed)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUserPassword(userID int, newPassword string) error {
	query := `UPDATE users SET password = ?, reset_token = NULL, reset_token_expiry = NULL WHERE id = ?`
	_, err := DB.Exec(query, newPassword, userID)
	return err
}

func AddLike(userID int, imageID string) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND image_id = ?)`
	err := DB.QueryRow(query, userID, imageID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("like already exists")
	}

	query = `INSERT INTO likes (user_id, image_id, created_at) VALUES (?, ?, ?)`
	_, err = DB.Exec(query, userID, imageID, time.Now())
	return err
}

func AddComment(imageID int, userID int, content string) error {
	query := `INSERT INTO comments (image_id, user_id, content, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`
	_, err := DB.Exec(query, imageID, userID, content)
	return err
}

func GetImageAuthor(imageID int) (*models.User, error) {
	query := `
        SELECT users.id, users.username, users.email, users.notify_on_comment
        FROM images
        JOIN users ON images.user_id = users.id
        WHERE images.id = ?
    `
	row := DB.QueryRow(query, imageID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.NotifyOnComment)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetRecentImagesByUser(userID int, limit int) ([]models.Image, error) {
	query := `
		SELECT images.id, images.user_id, images.file_path, images.created_at
		FROM images
		WHERE images.user_id = ?
		ORDER BY images.created_at DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(&image.ID, &image.UserID, &image.FilePath, &image.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func GetImageByID(imageID int) (*models.Image, error) {
	query := `SELECT id, file_path, user_id FROM images WHERE id = ?`
	row := DB.QueryRow(query, imageID)

	var image models.Image
	err := row.Scan(&image.ID, &image.FilePath, &image.UserID)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func DeleteImageByID(imageID int) error {
	query := `DELETE FROM images WHERE id = ?`
	_, err := DB.Exec(query, imageID)
	return err
}

func GetImagesPaginated(userID, limit, offset int) ([]models.Image, error) {
	query := `
        SELECT 
            images.id, 
            images.user_id, 
            images.file_path, 
            images.created_at,
            (SELECT COUNT(*) FROM likes WHERE likes.image_id = images.id) AS likes_count,
			images.user_id = ? AS is_owner
        FROM images
        ORDER BY images.created_at DESC
		LIMIT ? OFFSET ?
    `

	rows, err := DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	images := []models.Image{}
	for rows.Next() {
		var image models.Image
		if err := rows.Scan(&image.ID, &image.UserID, &image.FilePath, &image.CreatedAt, &image.Likes, &image.IsOwner); err != nil {
			log.Printf("Error scanning image row: %v", err)
			return nil, err
		}

		comments, err := GetCommentsByImageID(image.ID)
		if err != nil {
			log.Printf("Error fetching comments for image %d: %v", image.ID, err)
			return nil, err
		}
		image.Comments = comments

		images = append(images, image)
	}

	return images, nil
}
