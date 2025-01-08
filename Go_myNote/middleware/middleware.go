package middleware

import (
	"os"
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	// "fmt"
)

// AuthMiddleware ตรวจสอบว่าโทเค็น JWT ถูกต้องและยังไม่หมดอายุ
func AuthMiddleware(c *fiber.Ctx) error {
	// รับโทเค็นจาก Cookie หรือ Header
	tokenString := c.Cookies("jwt")
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization token not provided"})
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	// ตรวจสอบโทเค็น
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	// ดึงข้อมูล user_id จากโทเค็น
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token data"})
	}

	// แปลง user_id จาก claims เป็น uint
	userIDFloat := claims["user_id"].(float64)
	userID := uint(userIDFloat)

	// เพิ่ม user_id ใน Context เพื่อให้ handler ใช้ได้
	c.Locals("user_id", userID)
	// fmt.Printf("Middleware: user_id = %v\n", c.Locals("user_id"))


	// ตรวจสอบ `id` ใน URL (ถ้ามี)
	if c.Params("userid") != "" {
		requestedID, err := strconv.Atoi(c.Params("userid"))
		if err != nil || userID != uint(requestedID) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to access this resource"})
		}
	}
	
	return c.Next()
}

