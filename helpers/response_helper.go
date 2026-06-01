package helpers

import (
	fiber "github.com/gofiber/fiber/v2"
	models "github.com/heru-oktafian/fiber-apotek/models"
)

// getStatusText mengembalikan "success" atau "error" berdasarkan kode status HTTP
func getStatusText(status int) string {
	if status >= 200 && status < 300 {
		return "success"
	}
	return "error"
}

// JSONResponse mengirimkan format / struktur respons JSON standar
func JSONResponse(c *fiber.Ctx, status int, message string, data interface{}) error {
	resp := models.Response{
		Status:  getStatusText(status),
		Message: message,
		Data:    data,
	}
	return c.Status(status).JSON(resp)
}

// JSONResponseGetAll mengirimkan format / struktur respons JSON standar
func JSONResponseGetAll(c *fiber.Ctx, status int, message string, search string, total_items int, current_page int, total_pages int, per_page int, data interface{}) error {
	resp := models.ResponseGetAll{
		Status:      getStatusText(status),
		Message:     message,
		Search:      search,
		TotalItems:  total_items,
		CurrentPage: current_page,
		TotalPages:  total_pages,
		PerPage:     per_page,
		Data:        data,
	}
	return c.Status(status).JSON(resp)
}

func JSONResponseFlat(c *fiber.Ctx, status int, message string, payload map[string]interface{}) error {
	payload["status"] = "success"
	payload["message"] = message
	return c.Status(status).JSON(payload)
}
