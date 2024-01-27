package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Sukrati192/greenlight/internal/data"
	"github.com/Sukrati192/greenlight/internal/validator"
	"github.com/gin-gonic/gin"
)

func (app *application) createAuthentication(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := app.readJSON(c, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(c, err)
	}
	if !match {
		app.invalidCredentialsResponse(c)
		return
	}
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	if err = app.writeJSON(c, http.StatusOK, envelope{"authentication_token": token}, nil); err != nil {
		app.serverErrorResponse(c, err)
	}
}
