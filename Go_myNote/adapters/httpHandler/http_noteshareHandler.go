package httpHandler

import (
	"miw/usecases/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ShareNoteHandler struct {
	shareNoteUseCase service.ShareNoteUseCase
}

func NewShareNoteHandler(useCase service.ShareNoteUseCase) *ShareNoteHandler {
	return &ShareNoteHandler{shareNoteUseCase: useCase}
}

func (h *ShareNoteHandler) GetSharedEmailsHandler(c *fiber.Ctx) error {
	noteIDStr := c.Params("noteid")

	// แปลง noteID จาก string เป็น uint
	noteID, err := strconv.ParseUint(noteIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid note ID",
		})
	}

	emails, err := h.shareNoteUseCase.GetSharedEmailsByNoteID(uint(noteID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve shared emails",
		})
	}

	return c.JSON(fiber.Map{
		"shared_emails": emails,
	})
}

func (h *ShareNoteHandler) ShareNoteHandler(c *fiber.Ctx) error {
	var request struct {
		NoteID uint   `json:"note_id"`
		Email  string `json:"email"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ownerID := c.Locals("user_id").(uint)

	// แชร์โน้ตและรับรายการอีเมลที่แชร์
	sharedEmails, err := h.shareNoteUseCase.ShareNoteWithEmail(request.NoteID, ownerID, request.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":      "Note shared successfully",
		"share_emails": sharedEmails,
	})
}

func (h *ShareNoteHandler) RemoveShareHandler(c *fiber.Ctx) error {
	var request struct {
		NoteID uint   `json:"note_id"`
		Email  string `json:"email"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ownerID := c.Locals("user_id").(uint)

	// ลบแชร์โน้ต
	if err := h.shareNoteUseCase.RemoveShareByEmail(request.NoteID, ownerID, request.Email); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// ดึง shared_emails ที่เหลืออยู่
	sharedEmails, err := h.shareNoteUseCase.GetSharedEmailsByNoteID(request.NoteID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":       "Shared note removed successfully",
		"shared_emails": sharedEmails,
	})
}
