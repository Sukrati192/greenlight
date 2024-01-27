package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sukrati192/greenlight/internal/logger"
	"github.com/gin-gonic/gin"
)

func GetTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
	}
	return ctx
}

func Test_application_readIDParam(t *testing.T) {
	type fields struct {
		config config
		logger *logger.Logger
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			got, err := app.readIDParam(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("application.readIDParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("application.readIDParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_application_writeJSON(t *testing.T) {
	type fields struct {
		config config
		logger *logger.Logger
	}
	type args struct {
		c       *gin.Context
		status  int
		data    map[string]interface{}
		headers http.Header
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{
				config: tt.fields.config,
				logger: tt.fields.logger,
			}
			if err := app.writeJSON(tt.args.c, tt.args.status, tt.args.data, tt.args.headers); (err != nil) != tt.wantErr {
				t.Errorf("application.writeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	app := &application{config: config{}, logger: nil}
	r := GetTestContext()
	for n := 0; n < b.N; n++ {
		app.writeJSON(r, http.StatusOK, map[string]interface{}{"test": healthcheckResponse{Status: "okay"}}, nil)
	}
}

func BenchmarkWriteJSONWithoutIndent(b *testing.B) {
	app := &application{config: config{}, logger: nil}
	r := GetTestContext()
	for n := 0; n < b.N; n++ {
		app.writeJSONWithoutIndent(r, http.StatusOK, map[string]interface{}{"test": healthcheckResponse{Status: "okay"}}, nil)
	}
}
