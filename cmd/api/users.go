package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Sukrati192/greenlight/internal/data"
	"github.com/Sukrati192/greenlight/internal/validator"
	"github.com/gin-gonic/gin"
)

func (app *application) registerUserHandler(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := app.readJSON(c, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	if err := user.Password.Set(input.Password); err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	if err := app.models.Users.Insert(user); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(c, v.Errors)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	if err := app.models.Permissions.AddForUser(user.ID, "movies:read"); err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		if err := app.mailer.Send(user.Email, "user_welcome.html", data); err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	if err := app.writeJSON(c, http.StatusAccepted, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(c, err)
	}
}

func (app *application) activateUserHandler(c *gin.Context) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	if err := app.readJSON(c, &input); err != nil {
		app.badRequestResponse(c, err)
		return
	}
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(c, v.Errors)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	user.Activated = true
	if err := app.models.Users.Update(user); err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return
	}
	if err := app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID); err != nil {
		app.serverErrorResponse(c, err)
		return
	}
	if err := app.writeJSON(c, http.StatusOK, envelope{"user": user}, nil); err != nil {
		app.serverErrorResponse(c, err)
	}
}
