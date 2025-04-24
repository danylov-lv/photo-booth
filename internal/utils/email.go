package utils

import (
	"fmt"
)

func SendConfirmationEmail(email, token string) {
	fmt.Printf("[DEBUG] Confirmation email to %s: Click the link to confirm your account: http://localhost:8080/confirm?token=%s\n", email, token)
}

func SendPasswordResetEmail(email, token string) {
	fmt.Printf("[DEBUG] Password reset email to %s: Click the link to reset your password: http://localhost:8080/password/change?token=%s\n", email, token)
}

func SendCommentNotification(email, comment string) {
	fmt.Printf("[DEBUG]  Comment notification email to %s: New comment: %s\n", email, comment)
}
