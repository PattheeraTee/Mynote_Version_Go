package httpHandler

import (
	"miw/entities"
	"miw/usecases/service"
	"strconv"
	"strings"
	"github.com/gofiber/fiber/v2"
)

type HttpReminderHandler struct {
	reminderUseCase service.ReminderUseCase
}

func NewHttpReminderHandler(useCase service.ReminderUseCase) *HttpReminderHandler {
	return &HttpReminderHandler{reminderUseCase: useCase}
}

// เพิ่ม Reminder
func (h *HttpReminderHandler) AddReminderHandler(c *fiber.Ctx) error {
	noteID, _ := strconv.Atoi(c.Params("noteid"))

	data := new(entities.Reminder)
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เรียกใช้ Service Layer เพื่อเพิ่ม Reminder
	createdReminder, err := h.reminderUseCase.AddReminder(uint(noteID), userID, data)
	if err != nil {
		if strings.Contains(err.Error(), "does not belong to the user") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to add reminders to this note"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// ส่ง JSON ตอบกลับพร้อม Reminder ที่สร้างใหม่
	return c.JSON(fiber.Map{
		"message":  "Reminder added successfully",
		"reminder": createdReminder,
	})
}

// แสดง Reminder ทั้งหมดของ Note
func (h *HttpReminderHandler) GetRemindersHandler(c *fiber.Ctx) error {
	// ดึง noteID จาก URL พารามิเตอร์
	noteID, err := strconv.Atoi(c.Params("noteid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid note ID"})
	}

	// ดึง userID จาก Context (Middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// เรียกใช้ Service Layer เพื่อดึง Reminder
	reminders, err := h.reminderUseCase.GetReminderByNoteID(userID, uint(noteID))
	if err != nil {
		if strings.Contains(err.Error(), "not belong to the user") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to access this note's reminders"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(reminders)
}


// อัปเดต Reminder
func (h *HttpReminderHandler) UpdateReminderHandler(c *fiber.Ctx) error {
	// ดึง reminderID จาก URL พารามิเตอร์
	reminderID, err := strconv.Atoi(c.Params("reminderid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reminder ID"})
	}

	// ดึง userID จาก Context (Middleware)
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// รับข้อมูลจาก Body
	data := new(struct {
		ReminderTime string `json:"reminder_time"`
		Recurring    *bool  `json:"recurring"`  // ใช้ *bool เพื่อรองรับ nil
		Frequency    string `json:"frequency"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// แปลง frequency ให้เป็น *string
	var frequencyPointer *string
	if data.Frequency != "" {
		frequencyPointer = &data.Frequency
	}

	// เรียก Service Layer เพื่ออัปเดต Reminder
	err = h.reminderUseCase.UpdateReminder(userID, uint(reminderID), &data.ReminderTime, data.Recurring, frequencyPointer)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reminder not found"})
		}
		if strings.Contains(err.Error(), "does not belong to the user") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to update this reminder"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึง Reminder ที่อัปเดตแล้ว
	updatedReminder, err := h.reminderUseCase.GetReminderByID(uint(reminderID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated reminder"})
	}

	return c.JSON(fiber.Map{
		"message":  "Reminder updated successfully",
		"reminder": updatedReminder,
	})
}

// ลบ Reminder
func (h *HttpReminderHandler) DeleteReminderHandler(c *fiber.Ctx) error {
    reminderID, err := strconv.Atoi(c.Params("reminderid"))
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid reminder ID"})
    }
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

    // เรียกใช้ Service เพื่อดำเนินการลบ Reminder
    if err := h.reminderUseCase.DeleteReminder(userID,uint(reminderID)); err != nil {
        if strings.Contains(err.Error(), "not found") {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Reminder not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Reminder deleted successfully"})
}

