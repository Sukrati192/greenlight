package main

import (
	"github.com/Sukrati192/greenlight/internal/data"
	"github.com/gin-gonic/gin"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(c *gin.Context, user *data.User) {
	c.Set(string(userContextKey), user)
}

func (app *application) contextGetUser(c *gin.Context) *data.User {
	user, ok := c.Get(string(userContextKey))
	if !ok {
		panic("missing user value in request context")
	}
	dataUser, ok := user.(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return dataUser
}
