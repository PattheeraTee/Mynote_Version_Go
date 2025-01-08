package httpHandler

import (
	"fmt"
	"miw/entities"
	"miw/usecases/service"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HttpUserHandler struct {
	userUseCase service.UserUseCase
}

func NewHttpUserHandler(useCase service.UserUseCase) *HttpUserHandler {
	return &HttpUserHandler{userUseCase: useCase}
}

func (h *HttpUserHandler) Register(c *fiber.Ctx) error {
	user := new(entities.User)

	// รับข้อมูลผู้ใช้จาก Body และกรอง UserID
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ตรวจสอบให้แน่ใจว่า UserID เป็นค่าเริ่มต้น (ไม่กำหนดเอง)
	user.UserID = 0

	// เรียกใช้ฟังก์ชันสร้างผู้ใช้
	if err := h.userUseCase.Register(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "This email has already been registered."})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

func (h *HttpUserHandler) Login(c *fiber.Ctx) error {
	data := new(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	token, err := h.userUseCase.Login(data.Email, data.Password)
	fmt.Println("Token:", token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Email or password is incorrect")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(time.Hour * 72),
		HTTPOnly: true,  // ไม่อนุญาตเข้าถึงผ่าน JavaScript
		Secure:   false,  // ตั้งเป็น false ใน localhost
		SameSite: "None", // อนุญาตสำหรับ same-origin requests
		Path:     "/",
	})
	token = c.Cookies("jwt")
	fmt.Println("Cookie Value:", token)

	return c.JSON(fiber.Map{"message": "Login successful"})
}

func (h *HttpUserHandler) ForgotPassword(c *fiber.Ctx) error {
	data := new(struct {
		Email string `json:"email"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if err := h.userUseCase.SendResetPasswordEmail(data.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not send reset email")
	}

	return c.JSON(fiber.Map{"message": "Reset password email sent"})
}

func (h *HttpUserHandler) ChangePassword(c *fiber.Ctx) error {
	data := new(struct {
		Token           string `json:"token"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	})

	// อ่านค่า JSON จาก body
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// ตรวจสอบว่ารหัสผ่านใหม่และยืนยันรหัสผ่านตรงกันและไม่ว่าง
	if data.NewPassword == "" || data.ConfirmPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password fields cannot be empty"})
	}
	if data.NewPassword != data.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	// เรียกใช้ฟังก์ชัน ResetPassword ของ use case
	if err := h.userUseCase.ResetPassword(data.Token, data.NewPassword); err != nil {
		if err.Error() == "user not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Could not reset password"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password updated successfully",
	})
}

func (h *HttpUserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("userid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	user, err := h.userUseCase.GetUser(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("User not found")
	}

	return c.JSON(user)
}

func (h *HttpUserHandler) ChangeUsername(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("userid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	var requestBody map[string]interface{}
	if err := c.BodyParser(&requestBody); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	if newUsername, exists := requestBody["username"]; exists && len(requestBody) == 1 {
		newUsernameStr := newUsername.(string)

		if err := h.userUseCase.ChangeUsername(uint(id), newUsernameStr); err != nil {
			if err.Error() == "user not found" {
				return c.Status(fiber.StatusNotFound).SendString("user not found")
			}
			return c.Status(fiber.StatusInternalServerError).SendString("could not change username")
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "username updated successfully",
		})
	}

	return c.Status(fiber.StatusBadRequest).SendString("Only 'username' field is allowed")
}
