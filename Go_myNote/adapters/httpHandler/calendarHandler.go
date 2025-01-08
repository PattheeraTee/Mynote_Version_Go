package httpHandler

import (
	"context"
	"fmt"
	"log"

	"miw/entities"
	"miw/usecases/service"

	// "strings"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
)

type CalendarHandler struct {
	calendarService service.CalendarService
	store           *session.Store
}

func NewCalendarHandler(service service.CalendarService, store *session.Store) *CalendarHandler {
	return &CalendarHandler{
		calendarService: service,
		store:           store,
	}
}

func (h *CalendarHandler) HandleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		log.Println("Code not found in callback")
		return c.Status(fiber.StatusBadRequest).SendString("Code not found")
	}

	token, err := h.calendarService.ExchangeCode(context.Background(), code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to retrieve token")
	}

	log.Printf("Successfully retrieved token: AccessToken=%s, Expiry=%s", token.AccessToken, token.Expiry)

	sess, err := h.store.Get(c)
	if err != nil {
        log.Printf("Session save error: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to create session")
	}

	sess.Set("token", token)
	if err := sess.Save(); err != nil {
		log.Printf("Session save error: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to save session")
	}

	// Save token in a cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		Expires:  token.Expiry,
		HTTPOnly: true,
		Secure:   false, // Set to true if using HTTPS
	})

	log.Printf("Token saved to session: AccessToken=%s, RefreshToken=%s, Expiry=%s",
		token.AccessToken, token.RefreshToken, token.Expiry)

	// return c.Redirect("http://localhost:3000/form")
	// return c.SendString("Token saved successfully")
	return c.Redirect("http://localhost:3000/note")
}

func (h *CalendarHandler) ServeCreateForm(c *fiber.Ctx) error {
	sess, err := h.store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve session")
	}

	token := sess.Get("token")
	if token == nil {
		return c.Redirect("/")
	}

	return c.SendFile("form.html")
}
func (h *CalendarHandler) CreateEvent(c *fiber.Ctx) error {
	// Retrieve token from cookie
	cookie := c.Cookies("access_token")
	if cookie == "" {
		log.Println("Access token not found in cookies")
		return c.Status(fiber.StatusUnauthorized).SendString("User not authenticated")
	}

	token := &oauth2.Token{
		AccessToken: cookie,
		TokenType:   "Bearer",
	}

	log.Printf("Using token from cookie: AccessToken=%s, TokenType=%s", token.AccessToken, token.TokenType)

    // Parse JSON body
    var eventData struct {
        Summary     string `json:"summary"`
        Location    string `json:"location"`
        Description string `json:"description"`
        Start       string `json:"start"`
        End         string `json:"end"`
    }

    if err := c.BodyParser(&eventData); err != nil {
        log.Printf("Error parsing JSON body: %v", err)
        return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON body")
    }

    // Validate fields
    if eventData.Summary == "" || eventData.Start == "" || eventData.End == "" {
        return c.Status(fiber.StatusBadRequest).SendString("Missing required fields: summary, start, or end")
    }

    // Create an Event entity
    event := &entities.EventGoogle{
        Summary:     eventData.Summary,
        Location:    eventData.Location,
        Description: eventData.Description,
        Start:       eventData.Start,
        End:         eventData.End,
    }

    // Call the service layer
    createdEvent, err := h.calendarService.CreateEvent(token, event)
    if err != nil {
        log.Printf("Google Calendar API error: %v", err)
        return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Unable to create event: %v", err))
    }

    // Respond with success
    return c.JSON(fiber.Map{
        "message": "Event created successfully",
        "event":   createdEvent,
    })
}

// func (h *CalendarHandler) CreateEvent(c *fiber.Ctx) error {
//     var token *oauth2.Token

//     // Check for Bearer token in the Authorization header
//     authHeader := c.Get("Authorization")
//     if authHeader != "" {
//         // Extract the Bearer token
// 		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {            // accessToken := authHeader[7:]
//             token = &oauth2.Token{
// 				AccessToken: authHeader[7:], // Correctly extract the token without the "Bearer " prefix
// 				TokenType:   "Bearer",
// 			}
//         } else {
//             return c.Status(fiber.StatusBadRequest).SendString("Invalid Authorization header format")
//         }
//     } else {
//         // Fallback to session-based token retrieval
//         sess, err := h.store.Get(c)
//         if err != nil {
//             return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve session")
//         }

//         var ok bool
//         token, ok = sess.Get("token").(*oauth2.Token)
//         if !ok || token == nil {
//             return c.Status(fiber.StatusUnauthorized).SendString("User not authenticated")
//         }
//     }

//     log.Printf("Using token: AccessToken=%s, TokenType=%s", token.AccessToken, token.TokenType)

//     // Parse JSON body
//     var eventData struct {
//         Summary     string `json:"summary"`
//         Location    string `json:"location"`
//         Description string `json:"description"`
//         Start       string `json:"start"`
//         End         string `json:"end"`
//     }

//     if err := c.BodyParser(&eventData); err != nil {
//         log.Printf("Error parsing JSON body: %v", err)
//         return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON body")
//     }

//     // Validate fields
//     if eventData.Summary == "" || eventData.Start == "" || eventData.End == "" {
//         return c.Status(fiber.StatusBadRequest).SendString("Missing required fields: summary, start, or end")
//     }

//     // Create an Event entity
//     event := &entities.EventGoogle{
//         Summary:     eventData.Summary,
//         Location:    eventData.Location,
//         Description: eventData.Description,
//         Start:       eventData.Start,
//         End:         eventData.End,
//     }

//     // Call the service layer
//     createdEvent, err := h.calendarService.CreateEvent(token, event)
//     if err != nil {
//         log.Printf("Google Calendar API error: %v", err)
//         return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Unable to create event: %v", err))
//     }

//     // Respond with success
//     return c.JSON(fiber.Map{
//         "message": "Event created successfully",
//         "event":   createdEvent,
//     })
// }



// func (h *CalendarHandler) CreateEvent(c *fiber.Ctx) error {
// 	// ตรวจสอบ Authorization Header หรือ Body
// 	authHeader := c.Get("Authorization")
// 	var token *oauth2.Token

// 	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
// 		// ใช้ Authorization Header
// 		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
// 		token = &oauth2.Token{
// 			AccessToken: accessToken,
// 			TokenType:   "Bearer",
// 		}
// 	} else {
// 		// ตรวจสอบ JSON Body
// 		var body struct {
// 			AccessToken string `json:"access_token"`
// 			TokenType   string `json:"token_type"`
// 			Expiry      string `json:"expiry"`
// 		}
// 		if err := c.BodyParser(&body); err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 				"error": "Invalid JSON body",
// 			})
// 		}

// 		token = &oauth2.Token{
// 			AccessToken: body.AccessToken,
// 			TokenType:   body.TokenType,
// 		}
// 	}

// 	// ตรวจสอบว่า token ถูกต้อง
// 	if token == nil || token.AccessToken == "" {
// 		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 			"error": "Token missing or invalid",
// 		})
// 	}

// 	// Parse Event Data
// 	var event entities.EventGoogle
// 	if err := c.BodyParser(&event); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid event data",
// 		})
// 	}

// 	// สร้าง Event
// 	createdEvent, err := h.calendarService.CreateEvent(token, &event)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": fmt.Sprintf("Failed to create event: %v", err),
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "Event created successfully",
// 		"event":   createdEvent,
// 	})
// }
