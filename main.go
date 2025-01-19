package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// TrackingData stores information about email tracking
type TrackingData struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Subject   string     `json:"subject"`
	CreatedAt time.Time  `json:"created_at"`
	OpenedAt  *time.Time `json:"opened_at,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	UserAgent string     `json:"user_agent,omitempty"`
}

// Config holds application configuration
type Config struct {
	Port        string
	DatabaseURL string
	Environment string
}

var (
	db     *sql.DB
	config Config
)

func main() {
	// Initialize configuration
	initConfig()

	// Initialize database
	initDB()
	defer db.Close()

	// Create database tables
	createTables()

	// Set up Gin router
	router := setupRouter()

	// Start server
	port := config.Port
	log.Printf("Server starting in %s mode on port %s", config.Environment, port)
	log.Fatal(router.Run(":" + port))
}

func initConfig() {
	// Load environment variables from .env file if present
	_ = godotenv.Load()

	// Set default values
	config = Config{
		Port:        "8080",
		DatabaseURL: "postgres://postgres:postgres@localhost/mailping?sslmode=disable",
		Environment: "development",
	}

	// Override with environment variables if present
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}

	if env := os.Getenv("APP_ENV"); env != "" {
		config.Environment = env
	}

	// Configure Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func initDB() {
	var err error
	db, err = sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database successfully")
}

func createTables() {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tracking (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		subject TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		opened_at TIMESTAMP,
		ip_address TEXT,
		user_agent TEXT
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	// Middleware for request logging
	router.Use(gin.Logger())

	// Recovery middleware to handle panics
	router.Use(gin.Recovery())

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Static files
	router.Static("/static", "./static")

	// Routes
	router.GET("/", homeHandler)
	router.POST("/create", createTrackingPixelHandler)
	router.GET("/pixel/:id", pixelHandler)
	router.GET("/stats/:id", statsHandler)
	router.GET("/api/tracking/:id", apiStatsHandler)

	return router
}

func homeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", gin.H{
		"title": "Email Tracker",
	})
}

func createTrackingPixelHandler(c *gin.Context) {
	email := c.PostForm("email")
	subject := c.PostForm("subject")

	if email == "" || subject == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and subject are required"})
		return
	}

	// Generate unique ID for tracking
	trackingID := uuid.New().String()
	createdAt := time.Now()

	// Store tracking data in database
	_, err := db.Exec(
		"INSERT INTO tracking (id, email, subject, created_at) VALUES ($1, $2, $3, $4)",
		trackingID, email, subject, createdAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tracking record"})
		log.Println("Database error:", err)
		return
	}

	// Get base URL (consider protocol and host)
	protocol := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		protocol = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", protocol, c.Request.Host)
	pixelURL := fmt.Sprintf("%s/pixel/%s", baseURL, trackingID)
	htmlCode := fmt.Sprintf(`<img src="%s" width="1" height="1" alt="" />`, pixelURL)
	statsURL := fmt.Sprintf("%s/stats/%s", baseURL, trackingID)

	c.HTML(http.StatusOK, "result.html", gin.H{
		"title":      "Tracking Pixel Generated",
		"htmlCode":   htmlCode,
		"email":      email,
		"subject":    subject,
		"statsURL":   statsURL,
		"baseURL":    baseURL,
		"trackingID": trackingID,
	})
}

func pixelHandler(c *gin.Context) {
	trackingID := c.Param("id")
	if trackingID == "" {
		c.Status(http.StatusNotFound)
		return
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Log the email open event
	now := time.Now()
	_, err := db.Exec(
		"UPDATE tracking SET opened_at = $1, ip_address = $2, user_agent = $3 WHERE id = $4 AND opened_at IS NULL",
		now, ipAddress, userAgent, trackingID,
	)
	if err != nil {
		log.Println("Failed to update tracking record:", err)
	}

	// Return a transparent 1x1 pixel
	c.Data(http.StatusOK, "image/gif", transparentPixel())
}

func statsHandler(c *gin.Context) {
	trackingID := c.Param("id")
	if trackingID == "" {
		c.Status(http.StatusNotFound)
		return
	}

	// Get tracking data
	data, err := getTrackingData(trackingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(http.StatusNotFound)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			log.Println("Database error:", err)
		}
		return
	}

	// Determine status class for styling
	statusClass := "pending"
	status := "Email has not been opened yet."
	if data.OpenedAt != nil {
		statusClass = "opened"
		status = "Email has been opened!"
	}

	c.HTML(http.StatusOK, "stats.html", gin.H{
		"title":       "Email Tracking Statistics",
		"data":        data,
		"statusClass": statusClass,
		"status":      status,
	})
}

func apiStatsHandler(c *gin.Context) {
	trackingID := c.Param("id")
	if trackingID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tracking ID not found"})
		return
	}

	// Get tracking data
	data, err := getTrackingData(trackingID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tracking ID not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			log.Println("Database error:", err)
		}
		return
	}

	c.JSON(http.StatusOK, data)
}

func getTrackingData(trackingID string) (TrackingData, error) {
	var data TrackingData
	var openedAt sql.NullTime

	err := db.QueryRow(
		"SELECT id, email, subject, created_at, opened_at, ip_address, user_agent FROM tracking WHERE id = $1",
		trackingID,
	).Scan(&data.ID, &data.Email, &data.Subject, &data.CreatedAt, &openedAt, &data.IPAddress, &data.UserAgent)

	if err != nil {
		return data, err
	}

	if openedAt.Valid {
		data.OpenedAt = &openedAt.Time
	}

	return data, nil
}

// Helper function to generate a transparent 1x1 GIF pixel
func transparentPixel() []byte {
	// This is a raw representation of a transparent 1x1 GIF
	return []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00,
		0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x21, 0xF9, 0x04, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02,
		0x44, 0x01, 0x00, 0x3B,
	}
}
