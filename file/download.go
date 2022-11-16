package file

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/minpeter/tempfiles-backend/database"
)

func DownloadHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message":  "Please provide a file id",
			"error":    nil,
			"download": false,
		})
	}

	FileTracking := database.FileTracking{
		FileId: id,
	}

	has, err := database.Engine.Get(&FileTracking)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "db query error",
			"error":   err.Error(),
		})
	}

	if !has {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "file not found",
			"error":   nil,
		})
	}

	c.Response().Header.Set("Content-Disposition", "attachment; filename="+strings.ReplaceAll(url.PathEscape(FileTracking.FileName), "+", "%20"))
	return c.SendFile("tmp/" + FileTracking.FileId + "/" + FileTracking.FileName)
}
