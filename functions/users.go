package functions

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/azikazi/azikazi/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	User_id          int        `json:"user_id"`
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	Password         string     `json:"password"`
	Role             string     `json:"role"`
	Created          string     `json:"created"`
	Updated          string     `json:"updated"`
	ResetToken       *string    `json:"resetToken"` // Pointer to string to handle NULL
	ResetTokenExpiry *time.Time `json:"resetTokenExpiry"`
}

// Function to validate email format using regex
func isValidEmail(email string) bool {
	const emailRegexPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailRegexPattern)
	return regex.MatchString(email)
}
func CreateUser(ctx *gin.Context) {
	var req Users

	// Attempt to bind the JSON request body to req struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Log the exact error
		// fmt.Println("Error binding JSON: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Please ensure all fields are provided correctly."})
		return
	}

	// Normalize the email to lowercase
	req.Email = strings.ToLower(req.Email)

	// Validate email format
	if !isValidEmail(req.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format. Please provide a valid email like 'example@domain.com'."})
		return
	}

	// Check if the email already exists
	var existingEmail string
	err := database.DB.QueryRow("SELECT email FROM users WHERE email = $1", req.Email).Scan(&existingEmail)
	if err != nil && err != sql.ErrNoRows {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		fmt.Println("there is an error with the database:", err)
		return
	}
	if existingEmail != "" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Email is already in use"})
		return
	} 

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	req.Password = string(hashedPassword)
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	req.Created = currentTime

	// Insert the new user into the database
	_, err = database.DB.Exec("INSERT INTO users (name, email, password, role, created) VALUES ($1, $2, $3, $4, $5)", req.Name, req.Email, req.Password, req.Role, req.Created)
	if err != nil {
		fmt.Println("Database error during insert: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create a new user"})
		return
	}

	// Send welcome email to user
	// err = sendWelcomEmail(req.Email)
	// if err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending welcome email"})
	// 	return
	// }

	// Successfully created user
	ctx.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

// Define your secret key (store in .env later)
var jwtSecret = []byte("your_secret_key") // 🔐 use a secure random strin

// //// login function ////
func LoginUser(c *gin.Context) {
	var req Users
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Normalize the email to lowercase
	req.Email = strings.ToLower(req.Email)

	var storedEmployee Users
	err := database.DB.QueryRow(
		"SELECT user_id, name, email, password, role, created, resetToken, resetTokenExpiry FROM users WHERE email = $1",
		req.Email).
		Scan(
			&storedEmployee.User_id, &storedEmployee.Name, &storedEmployee.Email, &storedEmployee.Password, &storedEmployee.Role,
			&storedEmployee.Created, &storedEmployee.ResetToken, &storedEmployee.ResetTokenExpiry,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			fmt.Println("database error", err)
		}
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(storedEmployee.Password), []byte(req.Password))
	if err != nil {
		//fmt.Println("Stored hash:", storedEmployee.Password)
		//fmt.Println("Input password:", req.Password)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	fmt.Println("StoredEmployee:", storedEmployee)
    fmt.Println("StoredEmployee.User_id:", storedEmployee.User_id)

	// ✅ Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": storedEmployee.User_id,
		"email":   storedEmployee.Email,
		"role":    storedEmployee.Role,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // expires in 7 days
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// ✅ Send token + user info to frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   tokenString,
		"user": gin.H{
			"id":    storedEmployee.User_id,
			"name":  storedEmployee.Name,
			"email": storedEmployee.Email,
			"role":  storedEmployee.Role,
		},
	})
}


func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	  authHeader := c.GetHeader("Authorization")
	  if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		c.Abort()
		return
	  }
  
	  tokenString := strings.TrimPrefix(authHeader, "Bearer ")
  
	  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	  }) 
  
	  if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	  }
  
	  claims := token.Claims.(jwt.MapClaims)

	  fmt.Println("Claims:", claims)

	  userIDRaw := claims["user_id"]
	  
	  var userID int
	
	  switch v := userIDRaw.(type) {
	case float64:
	  userID = int(v)
	case int:
	  userID = v
	case string:
	  parsed, err := strconv.Atoi(v)
	  if err != nil {
		userID = 0
	  } else {
		userID = parsed
	  }
	default:
	  userID = 0
	}
		  
	  if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token payload"})
		c.Abort()
		return
	  }
	  
	  c.Set("user_id", userID)	  
  
	  c.Next()
	}
  }
  