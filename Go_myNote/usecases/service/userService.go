package service

import (
	"errors"
	"miw/entities"
	"miw/usecases/repository"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

type UserUseCase interface {
	Register(user *entities.User) error
	Login(email, password string) (string, error)
	ChangeUsername(userid uint, newUsername string) error
	SendResetPasswordEmail(email string) error
	ResetPassword(tokenString string, newPassword string) error 
	GetUser(userID uint) (*entities.User, error)
}

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register a new user
func (s *UserService) Register(user *entities.User) error {
	// ตรวจสอบให้แน่ใจว่า UserID ถูกรีเซ็ตเพื่อไม่ให้ผู้ใช้กำหนดเอง
	user.UserID = 0

	// แฮชรหัสผ่านก่อนบันทึก
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// บันทึกข้อมูลผู้ใช้
	return s.repo.CreateUser(user)
}

// Login a user and return JWT token
func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.generateToken(user.UserID)
}

// Generate JWT token for user
func (s *UserService) generateToken(userID uint) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	return token.SignedString([]byte(jwtSecret))
}

// Send reset password email
func (s *UserService) SendResetPasswordEmail(email string) error {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	resetToken, err := s.generateToken(user.UserID)
	if err != nil {
		return err
	}

	resetURL := "http://localhost:3000/change_password/" + resetToken
	return s.sendEmail(user.Email, resetURL)
}

func (s *UserService) sendEmail(email, resetURL string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "your-email@example.com")
	mailer.SetHeader("To", email)
	mailer.SetHeader("Subject", "Password Reset Request")
	mailer.SetBody("text/plain", "Click here to reset your password: "+resetURL)

	dialer := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("MAIL_EMAIL"), os.Getenv("MAIL_PASSWORD"))
	return dialer.DialAndSend(mailer)
}

// ChangeUsername changes the username of a user given their ID
func (s *UserService) ChangeUsername(userID uint, newUsername string) error {
	user, err := s.repo.GetUserById(userID)
	if err != nil {
		return err
	}
	user.Username = newUsername
	return s.repo.UpdateUser(user)
}

func (s *UserService) ResetPassword(tokenString string, newPassword string) error {
	jwtSecret := os.Getenv("JWT_SECRET")

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return errors.New("invalid token data")
	}

	// Extract user_id from token
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return errors.New("invalid token: user_id not found or invalid")
	}
	userID := uint(userIDFloat)

	// Retrieve user by user_id
	user, err := s.repo.GetUserById(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user's password
	user.Password = string(hashedPassword)
	return s.repo.UpdateUser(user)
}



func (s *UserService) GetUser(userID uint) (*entities.User, error) {
	return s.repo.GetUserById(userID)
}
