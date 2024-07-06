package http

import (
	"github.com/labstack/echo/v4"
	"go-db/domain"
	"net/http"
)

type godbHandler struct {
	godbUsecase domain.GoDBUsecase
}

func NewGoDBHandler(e *echo.Echo, godbUsecase domain.GoDBUsecase) {
	handler := &godbHandler{
		godbUsecase: godbUsecase,
	}

	e.GET("/health", handler.HealthCheck)

	e.POST("/execute", handler.Execute)

}

// HealthCheck is the handler for health check endpoint
func (h *godbHandler) HealthCheck(c echo.Context) error {

	var response = make(map[string]string)

	healthCheckStatus, err := h.godbUsecase.HealthCheck()

	response["status"] = healthCheckStatus

	if err != nil {
		return c.JSON(503, response)
	}
	return c.JSON(200, response)
}

func (h *godbHandler) Execute(c echo.Context) error {
	err := h.godbUsecase.Load("default")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	query := c.QueryParam("query")
	if query == "" {
		return c.JSON(http.StatusBadRequest, "Query parameter is required")
	}

	result, err := h.godbUsecase.Execute("default", query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	// Save the database state to a file
	err = h.godbUsecase.Save("default")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	switch v := result.(type) {
	case []domain.Row:
		c.Response().Header().Set("Content-Type", "application/json")

		// List of combined data
		var data []map[string]interface{}

		// Assuming 'v' is a slice of Row objects
		for _, row := range v {
			// Create a new map to hold combined data
			combined := make(map[string]interface{})

			// Add non-Data fields
			combined["id"] = row.ID
			combined["createdAt"] = row.CreatedAt
			combined["updatedAt"] = row.UpdatedAt

			// Merge Data fields
			for key, value := range row.Data {
				combined[key] = value
			}

			// Append the combined map to the list
			data = append(data, combined)
		}

		if len(data) == 0 {
			return c.JSON(http.StatusNotFound, "No data found")
		}
		// Encode the combined map
		return c.JSON(http.StatusOK, data)
	default:
		return c.JSON(http.StatusOK, result)
	}
}
