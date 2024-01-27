package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthcheckResponse struct {
	Status      string `json:"status"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

func (app *application) healthcheckHandler(c *gin.Context) {
	resp := envelope{
		"data": healthcheckResponse{
			Status:      "available",
			Version:     version,
			Environment: app.config.env,
		}}
	if err := app.writeJSON(c, http.StatusOK, resp, nil); err != nil {
		app.serverErrorResponse(c, err)
	}
}
