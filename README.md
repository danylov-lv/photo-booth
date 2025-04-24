# Photo Booth Project

## Overview
The Photo Booth project is a web application built with Go that allows users to take snapshots with camera overlays, upload images, view a gallery of saved images, and interact with them through likes and comments. The application supports user authentication, password reset, and infinite scrolling in the gallery.

## Project Structure
```
photo-booth
├── cmd
│   └── main.go               # Entry point of the application
├── controllers
│   ├── auth.go               # User authentication handling (registration, login, password reset)
│   ├── gallery.go            # Gallery handling for viewing and interacting with images
│   ├── camera.go             # Logic for taking snapshots, uploading images, and applying overlays
│   ├── comments.go           # Handling comments for images
│   ├── likes.go              # Handling likes for images
│   └── settings.go           # User settings management
├── internal
│   ├── db.go                 # Database initialization and operations
│   ├── middleware.go         # Middleware for user authentication and route protection
│   ├── utils
│   │   ├── email.go          # Utility functions for sending emails
│   │   └── token.go          # Utility functions for generating tokens
│   └── models
│       ├── user.go           # User data structure
│       ├── image.go          # Image data structure
│       └── comment.go        # Comment data structure
├── static
│   └── css
│       ├── img          # Directory for image assets
│       │   └── overlays   # Directory for overlay images
│       └── styles.css        # Styles for the web application
├── uploads               # Directory for user-uploaded images
├── templates
│   ├── index.html            # Template for the main page
│   ├── gallery.html          # Template for the gallery page
│   ├── camera.html           # Template for the camera page
│   ├── login.html            # Template for the login page
│   ├── register.html         # Template for the registration page
│   ├── settings.html         # Template for user settings
│   ├── reset_password.html   # Template for password reset
│   ├── change_password.html  # Template for changing the password
│   └── confirm_account.html  # Template for account confirmation
├── .env                      # Environment variables for configuration
├── go.mod                    # Module and dependency information
├── .gitignore                # Ignored files and directories
└── README.md                 # Project documentation
```

## Features
- **User Authentication**: Users can register, log in, and reset their passwords.
- **Image Capture and Upload**: Users can take snapshots using their camera with overlays or upload images directly.
- **Gallery**: Users can view a gallery of saved images with infinite scrolling.
- **Likes and Comments**: Users can like images and add comments to them.
- **User Settings**: Users can update their username, email, and password.
- **Password Reset**: Users can reset their password via email.
- **Responsive Design**: The application is optimized for both desktop and mobile devices.

## Getting Started

### Prerequisites
- Go 1.20 or later
- SQLite installed on your system

### Installation
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd photo-booth
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Create a `.env` file in the root directory and configure the following variables:
   ```env
   PORT=8080
   JWT_SECRET=your-secret-key
   RESET_TOKEN_EXPIRY=3600
   ```

4. Initialize the database:
   ```bash
   go run cmd/main.go
   ```

5. Open your browser and navigate to `http://localhost:8080`.

## Usage
- **Register**: Create a new account.
- **Log In**: Log in to access the gallery and camera features.
- **Take Photos**: Use your camera to take snapshots with overlays or upload images.
- **Interact with Images**: Like images, add comments, or delete your own images.
- **Manage Profile**: Update your username, email, or password in the settings.

## Acknowledgments
- Built with Go for backend development.
- Frontend styled with custom CSS.
- SQLite used for lightweight database storage.
