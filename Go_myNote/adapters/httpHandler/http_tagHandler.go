package httpHandler

import (
	"miw/entities"
	"miw/usecases/service"
	"strconv"
	"github.com/gofiber/fiber/v2"
)

type TagResponse struct {
	TagID   uint   `json:"tag_id"`
	TagName string `json:"tag_name"`
	UserID  uint   `json:"user_id"` 
	Notes   []uint `json:"notes"`
}

type HttpTagHandler struct {
	tagUseCase service.TagUseCase
}

func NewHttpTagHandler(useCase service.TagUseCase) *HttpTagHandler {
	return &HttpTagHandler{tagUseCase: useCase}
}

func (h *HttpTagHandler) CreateTagHandler(c *fiber.Ctx) error {
	tag := new(entities.Tag)

	// ดึง UserID จาก Context (Middleware ควรตั้งค่า UserID)
	userID := c.Locals("user_id").(uint)

	// รับข้อมูลแท็กจาก Body ของ request
	if err := c.BodyParser(tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// กำหนด UserID ให้กับแท็ก
	tag.UserID = userID

	// เรียกใช้ฟังก์ชันสร้างแท็ก
	if err := h.tagUseCase.CreateTag(tag); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Tag created successfully",
		"tag":     tag,
	})
}

func (h *HttpTagHandler) GetAllTagsHandler(c *fiber.Ctx) error {
	// ดึง UserID จาก Context
	userID := c.Locals("user_id").(uint)

	// ดึงแท็กทั้งหมดของผู้ใช้
	tags, err := h.tagUseCase.GetAllTagsByUserId(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch tags"})
	}

	// สร้าง JSON Response
	var response []TagResponse
	for _, tag := range tags {
		var noteIDs []uint
		for _, note := range tag.Notes {
			noteIDs = append(noteIDs, note.NoteID)
		}
		response = append(response, TagResponse{
			TagID:   tag.TagID,
			TagName: tag.TagName,
			UserID:  tag.UserID,
			Notes:   noteIDs,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *HttpTagHandler) GetTagHandler(c *fiber.Ctx) error {
	tagID, err := strconv.Atoi(c.Params("tagid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tag ID"})
	}

	// ดึง UserID จาก Context
	userID := c.Locals("user_id").(uint)

	// ดึง Tag ตาม ID และ UserID
	tag, err := h.tagUseCase.GetTagById(uint(tagID), userID)
	if err != nil {
		switch err.Error() {
		case "tag not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Tag not found"})
		case "you are not authorized to view this tag":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You are not authorized to view this tag"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch tag"})
		}
	}

	// สร้าง JSON Response
	var noteIDs []uint
	for _, note := range tag.Notes {
		noteIDs = append(noteIDs, note.NoteID)
	}

	response := TagResponse{
		TagID:   tag.TagID,
		TagName: tag.TagName,
		UserID:  tag.UserID,
		Notes:   noteIDs,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}


func (h *HttpTagHandler) UpdateTagNameHandler(c *fiber.Ctx) error {
	var request struct {
		NewTagname string `json:"new_tagname"`
	}

	// รับ tag ID จากพารามิเตอร์
	tagID, err := strconv.Atoi(c.Params("tagid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tag ID"})
	}

	// ดึง UserID จาก Context
	userID := c.Locals("user_id").(uint)

	// รับค่า new_name จาก body
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// เรียกใช้ service เพื่อแก้ไขชื่อแท็ก
	if err := h.tagUseCase.UpdateTagName(uint(tagID), userID, request.NewTagname); err != nil {
		// ใช้ข้อความเปรียบเทียบโดยตรงแทน fmt.Sprintf
		if err.Error() == "tag not found or does not belong to this user" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Tag not found"})
		}
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}
	

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Tag name updated successfully"})
}


func (h *HttpTagHandler) DeleteTagHandler(c *fiber.Ctx) error {
	tagID, err := strconv.Atoi(c.Params("tagid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid tag ID"})
	}

	// ดึง UserID จาก Context
	userID := c.Locals("user_id").(uint)

	// เรียกใช้ service เพื่อลบแท็ก
	if err := h.tagUseCase.DeleteTag(uint(tagID), userID); err != nil {
		// ใช้ข้อความเปรียบเทียบโดยตรงแทน fmt.Sprintf
		if err.Error() == "tag not found or does not belong to this user" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Tag not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Tag deleted successfully"})
}
