package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Secret key để xác thực token JWT
var jwtSecret = []byte("nguyen-secret-key") // Đổi thành secret key thật trong môi trường sản xuất

// Middleware xác thực JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Kiểm tra xem session đã tồn tại chưa
		session, _ := c.Cookie("session_token")
		if session != "" {
			log.Println("Session exists, skipping token check.")
			c.Next()
			return
		}

		// Lấy token từ header Authorization
		authHeader := c.GetHeader("Authorization")
		log.Print("Authorization header:", authHeader)

		if authHeader == "" {
			log.Println("Authorization header missing")
			// Chuyển hướng về trang login nếu không có token
			c.Redirect(http.StatusFound, "/auth/login?message=You need to login!")
			c.Abort()
			return
		}

		// Tách token từ chuỗi Authorization
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Println("Invalid authorization format")
			c.Redirect(http.StatusFound, "/auth/login?message=token split error!")
			c.Abort()
			return
		}

		// Parse và xác minh token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.NewValidationError("Unexpected signing method", jwt.ValidationErrorSignatureInvalid)
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Println("Unauthorized access or invalid token")
			c.Redirect(http.StatusFound, "/auth/login?message=Unauthorized access&target!")
			c.Abort()
			return
		}

		// Extract claims and store in session
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			role := claims["role"]
			if role != "admin" {
				log.Println("Access denied: Admin role required")
				c.Redirect(http.StatusFound, "/auth/login?message=You are not admin!")
				c.Abort()
				return
			}

			userInfo := map[string]interface{}{
				"userID":   claims["sub"],
				"email":    claims["email"],
				"username": claims["username"],
				"role":     claims["role"],
				"status":   claims["status"],
			}

			// Serialize user information as JSON for cookie storage
			userInfoJSON, err := json.Marshal(userInfo)
			if err != nil {
				log.Println("Error serializing user info:", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "session_token",
				Value:    string(userInfoJSON),
				Expires:  time.Now().Add(1 * time.Hour), // 1 hour session
				HttpOnly: true,
			})
		}

		// Cho phép tiếp tục xử lý
		c.Next()
	}
}

// Hàm tạo JWT token
func CreateToken(userID string, userEmail string, userUsername string, userPassword string, userRole string, userStatus int) (string, error) {
	// Khởi tạo các claims của token
	claims := jwt.MapClaims{
		"sub":      userID,                                // ID của người dùng
		"email":    userEmail,                             // Email của người dùng
		"username": userUsername,                          // Tên người dùng
		"role":     userRole,                              // Vai trò của người dùng
		"status":   userStatus,                            // Trạng thái của người dùng
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Thời hạn token
		"iat":      time.Now().Unix(),                     // Thời gian phát hành token
	}

	// Tạo token với phương thức ký HMAC và claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ký token với secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
