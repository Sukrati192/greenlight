package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sukrati192/greenlight/internal/data"
	"github.com/Sukrati192/greenlight/internal/validator"
	"github.com/felixge/httpsnoop"
	"github.com/gin-gonic/gin"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			c.Header("Connection", "close")
			app.serverErrorResponse(c, fmt.Errorf("%s", err))
			c.Abort()
		}
	}()
}

func (app *application) rateLimit() gin.HandlerFunc {
	if !app.config.limiter.enabled {
		return func(c *gin.Context) {}
	}
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(c *gin.Context) {
		ip := realip.FromRequest(c.Request)
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst), time.Now()}
		}
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(c)
			c.Abort()
		} else {
			mu.Unlock()
		}
	}
}

func (app *application) authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Authorization")
		authorizationHeader := c.GetHeader("Authorization")
		if authorizationHeader == "" {
			app.contextSetUser(c, data.AnonymousUser)
			return
		}
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationResponse(c)
			c.Abort()
		}
		token := headerParts[1]
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationResponse(c)
			c.Abort()
		}
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationResponse(c)
			default:
				app.serverErrorResponse(c, err)
			}
			c.Abort()
		}
		app.contextSetUser(c, user)
	}
}

func (app *application) requireAuthenticatedUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := app.contextGetUser(c)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(c)
			c.Abort()
		}
	}
}

func (app *application) requireActivatedUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		app.requireAuthenticatedUser()(c)
		if c.Writer.Written() {
			return
		}
		user := app.contextGetUser(c)
		if !user.Activated {
			app.inactiveAccountResponse(c)
			c.Abort()
		}
	}
}

func (app *application) requirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		app.requireActivatedUser()(c)
		if c.Writer.Written() {
			return
		}
		user := app.contextGetUser(c)
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(c, err)
			c.Abort()
			return
		}
		if !permissions.Include(code) {
			app.notPermittedResponse(c)
			c.Abort()
		}
	}
}

func (app *application) enableCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Vary", "Origin")
		c.Writer.Header().Add("Vary", "Access-Control-Request-Method")
		origin := c.GetHeader("Origin")
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			for _, trusted := range app.config.cors.trustedOrigins {
				if origin == trusted {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}
		}
		if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != "" {
			c.Header("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Max-Age", "60")
			c.AbortWithStatus(http.StatusOK)
		}
	}
}

func (app *application) metrics() gin.HandlerFunc {
	totalRequestsReceived := expvar.NewInt("total_requests_recieved")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_microseconds")
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")
	return func(c *gin.Context) {
		totalRequestsReceived.Add(1)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})
		metrics := httpsnoop.CaptureMetrics(handler, c.Writer, c.Request)
		totalResponsesSent.Add(1)
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	}
}
