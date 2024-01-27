package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Sukrati192/greenlight/internal/validator"
	"github.com/gin-gonic/gin"
)

type envelope map[string]interface{}

func (app *application) readIDParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		app.logger.PrintError(err, nil)
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

func (app *application) writeJSON(c *gin.Context, status int, data envelope, headers http.Header) error {
	for key := range headers {
		c.Header(key, headers.Get(key))
	}
	c.Header("Content-Type", "application/json")
	c.IndentedJSON(status, data)
	return nil
}

func (app *application) writeJSONWithoutIndent(c *gin.Context, status int, data envelope, headers http.Header) error {
	for key := range headers {
		c.Header(key, headers.Get(key))
	}
	c.Header("Content-Type", "application/json")
	c.JSON(status, data)
	return nil
}

func (app *application) readJSON(c *gin.Context, target interface{}) error {
	maxBytes := 1_048_576
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxBytes))
	gin.EnableJsonDecoderDisallowUnknownFields()
	if err := c.ShouldBindJSON(&target); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly formatted JSON (at character %d)", syntaxErr.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formatted JSON")
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for the field %q", unmarshalTypeErr.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeErr.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalErr):
			panic(err)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key: %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		default:
			return err
		}
	}

	//TODO: Handle the error for multiple JSON
	// rawData, _ := c.GetRawData()
	// if err := json.NewDecoder(strings.NewReader(string(rawData))).Decode(&struct{}{}); err != io.EOF {
	// 	app.logError(c, err)
	// 	return errors.New("body must only contain a single JSON value")
	// }
	return nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}
	return i
}

func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}
