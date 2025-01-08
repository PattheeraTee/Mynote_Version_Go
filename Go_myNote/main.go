package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"miw/adapters/gormRepository"
	"miw/adapters/httpHandler"
	"miw/database"
	"miw/entities"
	"miw/middleware"
	"miw/usecases/repository"
	"miw/usecases/service"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	store       *session.Store
	oauthConfig *oauth2.Config
)

func init() {
	// Register oauth2.Token type with gob
	gob.Register(&oauth2.Token{})
}

func main() {

	cfg := database.LoadConfig()
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable search_path=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSchema,
	)

	database, err := database.NewDatabaseConnection(dsn)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// สร้างตารางอัตโนมัติโดยใช้ AutoMigrate
	err = database.AutoMigrate(
		&entities.User{},
		&entities.Note{},
		&entities.Reminder{},
		&entities.Tag{},
		&entities.ShareNote{},
		&entities.Event{},
		&entities.ToDo{},
	)

	if err != nil {
		log.Fatal("Failed to migrate tables:", err)
	}

	// สร้าง Repository และ Service
	userRepo := gormRepository.NewGormUserRepository(database)
	noteRepo := gormRepository.NewGormNoteRepository(database)
	tagRepo := gormRepository.NewGormTagRepository(database)
	reminderRepo := gormRepository.NewGormReminderRepository(database)
	sharenoteRepo := gormRepository.NewGormShareNoteRepository(database)

	userService := service.NewUserService(userRepo)
	noteService := service.NewNoteService(noteRepo, sharenoteRepo)
	tagService := service.NewTagService(tagRepo, noteRepo)
	reminderService := service.NewReminderService(reminderRepo, noteRepo, userRepo)
	sharenoteService := service.NewShareNoteService(sharenoteRepo, noteRepo)

	// สร้าง Handlers สำหรับ HTTP
	userHandler := httpHandler.NewHttpUserHandler(userService)
	noteHandler := httpHandler.NewHttpNoteHandler(noteService)
	tagHandler := httpHandler.NewHttpTagHandler(tagService)
	reminderHandler := httpHandler.NewHttpReminderHandler(reminderService)
	sharenoteHandler := httpHandler.NewShareNoteHandler(sharenoteService)

	// สร้าง Fiber App และเพิ่ม Middleware
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000", // Allow your frontend origin
		AllowMethods:     "GET,POST,PUT,DELETE",   // Allowed HTTP methods
		AllowHeaders:     "Content-Type,Authorization", // Allowed headers
		AllowCredentials: true,                   // Allow cookies to be sent with requests
	}))
	

	store = session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSecure:   false,       // ใช้ true ถ้าเป็น HTTPS
		CookieSameSite: "None",      // รองรับ cross-site cookies
	})
	
		b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	oauthConfig, err = google.ConfigFromJSON(b, "https://www.googleapis.com/auth/calendar")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	calendarRepo := repository.NewGoogleCalendarRepository(oauthConfig)
	calendarService := service.NewCalendarService(calendarRepo)
	calendarHandler := httpHandler.NewCalendarHandler(calendarService, store)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("login.html")
	})

	app.Post("/set-token", func(c *fiber.Ctx) error {
		// Define the expected request body structure
		var body struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			RefreshToken string `json:"refresh_token"`
			Expiry      string `json:"expiry"` // Optional: For demonstration
		}
	
		// Parse the incoming request body
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
	
		// Create an OAuth2 token from the received data
		token := &oauth2.Token{
			AccessToken:  body.AccessToken,
			TokenType:    body.TokenType,
			RefreshToken: body.RefreshToken,
		}
	
		// Store the token in the session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve session")
		}
	
		sess.Set("token", token)
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to save token in session")
		}
	
		return c.JSON(fiber.Map{
			"message": "Token saved successfully",
			"token":   token,
		})
	})

	app.Get("/authorize", func(c *fiber.Ctx) error {
		authURL := oauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		return c.Redirect(authURL)
	})
	app.Get("/callback", calendarHandler.HandleCallback)
	app.Post("/callback", func(c *fiber.Ctx) error {
		var body struct {
			Code string `json:"code"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		token, err := oauthConfig.Exchange(context.Background(), body.Code)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to exchange token",
			})
		}

		// Save token to session or database
		session, _ := store.Get(c)
		session.Set("google_token", token)
		session.Save()

		return c.JSON(fiber.Map{
			"message": "OAuth successful",
			"token":   token,
		})
	})

	app.Get("/create", calendarHandler.ServeCreateForm)

	app.Get("/form", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve session")
		}
	
		token := sess.Get("token")
		if token == nil {
			return c.Redirect("/")
		}
		// return c.SendFile("form.html")
	    return c.Redirect("http://localhost:3000/form")

	})
	app.Post("/create", calendarHandler.CreateEvent)

	app.Get("/logout", func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve session")
		}

		// Destroy the session
		sess.Destroy()

		return c.SendStatus(fiber.StatusOK)
	})

	//********************************************
	// User
	//********************************************
	app.Post("/register", userHandler.Register)
	app.Post("/login", userHandler.Login)

	app.Post("/forgot-password", userHandler.ForgotPassword)
	app.Post("/reset-password", userHandler.ChangePassword)

	app.Get("/user/:userid", middleware.AuthMiddleware, userHandler.GetUser)        // ดูข้อมูล user
	app.Put("/user/:userid", middleware.AuthMiddleware, userHandler.ChangeUsername) // แก้ไข username

	//********************************************
	// Note
	//********************************************
	app.Post("/note", middleware.AuthMiddleware, noteHandler.CreateNoteHandler)        // สร้าง note
	app.Get("/note/:userid", middleware.AuthMiddleware, noteHandler.GetAllNoteHandler) // ดู note
	app.Put("/note/color/:noteid", middleware.AuthMiddleware, noteHandler.UpdateColorHandler)
	app.Put("/note/priority/:noteid", middleware.AuthMiddleware, noteHandler.UpdatePriorityHandler)
	app.Put("/note/title-content/:noteid", middleware.AuthMiddleware, noteHandler.UpdateTitleAndContentHandler)
	app.Put("/note/status/:noteid", middleware.AuthMiddleware, noteHandler.UpdateStatusHandler)
	app.Delete("/note/:noteid", middleware.AuthMiddleware, noteHandler.DeleteNoteHandler) // ลบ note
	app.Put("/note/restore/:noteid", middleware.AuthMiddleware, noteHandler.RestoreNoteHandler)
	app.Get("/note/deleted/:userid", middleware.AuthMiddleware, noteHandler.GetDeletedNotesHandler)
	//********************************************
	// Add Tag to Note And Remove Tag from Note
	//********************************************
	app.Post("/note/add-tag", middleware.AuthMiddleware, noteHandler.AddTagToNoteHandler)
	app.Post("/note/remove-tag", middleware.AuthMiddleware, noteHandler.RemoveTagFromNoteHandler)
	//********************************************
	// Reminder
	//********************************************
	app.Post("/note/reminder/:noteid", middleware.AuthMiddleware, reminderHandler.AddReminderHandler)
	app.Get("/note/reminder/:noteid", middleware.AuthMiddleware, reminderHandler.GetRemindersHandler)
	app.Put("/reminder/:reminderid", middleware.AuthMiddleware, reminderHandler.UpdateReminderHandler)
	app.Delete("/reminder/:reminderid", middleware.AuthMiddleware, reminderHandler.DeleteReminderHandler)

	//********************************************
	// Tag
	//********************************************
	app.Get("/tag", middleware.AuthMiddleware, tagHandler.GetAllTagsHandler)           // ดู tag ทั้งหมด
	app.Post("/tag", middleware.AuthMiddleware, tagHandler.CreateTagHandler)           // สร้าง tag
	app.Get("/tag/:tagid", middleware.AuthMiddleware, tagHandler.GetTagHandler)        // ดู tag
	app.Put("/tag/:tagid", middleware.AuthMiddleware, tagHandler.UpdateTagNameHandler) // แก้ไขชื่อ tag
	app.Delete("/tag/:tagid", middleware.AuthMiddleware, tagHandler.DeleteTagHandler)  // ลบ tag

	//********************************************
	// sharenote
	//********************************************
	app.Post("/note/share", middleware.AuthMiddleware, sharenoteHandler.ShareNoteHandler)
	app.Get("/note/:noteid/shared-emails", middleware.AuthMiddleware, sharenoteHandler.GetSharedEmailsHandler)
	app.Post("/note/remove-share", middleware.AuthMiddleware, sharenoteHandler.RemoveShareHandler)

	// เริ่มเซิร์ฟเวอร์
	if err := app.Listen(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}

}
