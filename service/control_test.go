// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/model"
)

const (
	apiKey         = "changeme"
	apiKeyHashed   = "057ba03d6c44104863dc7361fe4578965d1887360f90a0895882e58a6248fc86"
	authGetPath    = "/v1/workflow/test"
	authTestName   = "auth_test"
	getTestName    = "get_test"
	deleteTestName = "delete_test8"
	testDBName     = "sprinkler_unit_test.db"
)

var now = time.Now()
var mockNextRuntime = staticNextRuntime()

func getMockDB(dbPath string) *gorm.DB {
	os.Remove(dbPath)
	mockDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	mockDB.AutoMigrate(database.Tables...)

	var deleteTest table.Workflow
	mockDB.Where(table.Workflow{Name: deleteTestName}).Assign(table.Workflow{
		Name:        deleteTestName,
		Artifact:    "test.jar",
		Command:     "java -jar test.jar",
		Every:       model.Every{1, model.EveryDay},
		NextRuntime: mockNextRuntime,
		Backfill:    false,
		IsActive:    true,
	}).FirstOrCreate(&deleteTest)

	var getTest table.Workflow
	mockDB.Where(table.Workflow{Name: getTestName}).Assign(table.Workflow{
		Name:        getTestName,
		Artifact:    "test.jar",
		Command:     "java -jar test.jar",
		Every:       model.Every{1, model.EveryDay},
		NextRuntime: mockNextRuntime,
		Backfill:    false,
		IsActive:    true,
	}).FirstOrCreate(&getTest)

	var authTest table.Workflow
	mockDB.Where(table.Workflow{Name: authTestName}).Assign(table.Workflow{
		Name:        authTestName,
		Artifact:    "test.jar",
		Command:     "java -jar test.jar",
		Every:       model.Every{1, model.EveryDay},
		NextRuntime: mockNextRuntime,
		Backfill:    false,
		IsActive:    true,
	}).FirstOrCreate(&authTest)

	return mockDB
}

func cleanupDB(db *gorm.DB, dbName string) {
	db.Exec("DELETE FROM workflow_activator_lock")
	db.Exec("DELETE FROM workflow_scheduler_lock")
	db.Exec("DELETE FROM scheduled_workflow")
	db.Exec("DELETE FROM workflow")
	os.Remove(dbName)
}

func staticNextRuntime() time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func TestDeleteWorkflow(t *testing.T) {
	dbName := fmt.Sprintf("%s_%s", uuid.New().String(), testDBName)
	mockDB := getMockDB(dbName)
	ctrl := &Control{db: mockDB}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/v1/workflow", ctrl.deleteWorkflow)

	t.Run("Valid request", func(t *testing.T) {

		body := deleteWorkflowReq{Name: deleteTestName}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("DELETE", "/v1/workflow", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Workflow not found", func(t *testing.T) {

		body := deleteWorkflowReq{Name: "nonexistent"}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("DELETE", "/v1/workflow", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid request - bad JSON", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/v1/workflow", bytes.NewBufferString("{\"foo\":\"bad json\"}"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	cleanupDB(mockDB, dbName)
}

func TestPutWorkflow(t *testing.T) {
	dbName := fmt.Sprintf("%s_%s", uuid.New().String(), testDBName)
	mockDB := getMockDB(dbName)
	ctrl := &Control{db: mockDB}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/v1/workflow", ctrl.putWorkflow)

	t.Run("Valid request", func(t *testing.T) {

		body := putWorkflowReq{
			Name:        "put_test",
			Artifact:    "test.jar",
			Command:     "java -jar test.jar",
			Every:       "1.hour",
			NextRuntime: staticNextRuntime(),
			Backfill:    false,
			IsActive:    true,
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", "/v1/workflow", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "\"OK\"", w.Body.String())
	})

	t.Run("Invalid request - bad JSON", func(t *testing.T) {
		req, _ := http.NewRequest("PUT", "/v1/workflow", bytes.NewBufferString("{\"foo\":\"bad json\"}"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid request - invalid 'every' field", func(t *testing.T) {
		body := putWorkflowReq{
			Name:        "invalid_put_test",
			Artifact:    "test.jar",
			Command:     "java -jar test.jar",
			Every:       "invalid",
			NextRuntime: time.Now(),
		}
		jsonBody, _ := json.Marshal(body)

		req, _ := http.NewRequest("PUT", "/v1/workflow", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	cleanupDB(mockDB, dbName)
}

func TestGetWorkflow(t *testing.T) {
	dbName := fmt.Sprintf("%s_%s", uuid.New().String(), testDBName)
	mockDB := getMockDB(dbName)
	ctrl := &Control{db: mockDB}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/workflow/:name", ctrl.getWorkflow)

	t.Run("Existing workflow", func(t *testing.T) {
		workflow := table.Workflow{
			Name:        getTestName,
			Artifact:    "test.jar",
			Command:     "java -jar test.jar",
			Every:       model.Every{1, model.EveryDay},
			NextRuntime: staticNextRuntime(),
			Backfill:    false,
			IsActive:    true,
		}

		testPath := fmt.Sprintf("/v1/workflow/%s", getTestName)
		req, _ := http.NewRequest("GET", testPath, nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response putWorkflowReq
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, workflow.Name, response.Name)
		assert.Equal(t, workflow.Artifact, response.Artifact)
		assert.Equal(t, workflow.Command, response.Command)
		assert.Equal(t, workflow.Every.String(), response.Every)
		assert.Equal(t, workflow.NextRuntime, response.NextRuntime)
		assert.Equal(t, workflow.Backfill, response.Backfill)
		assert.Equal(t, workflow.IsActive, response.IsActive)
	})

	t.Run("Non-existent workflow", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1//workflow/nonexistent", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	cleanupDB(mockDB, dbName)
}

func TestAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(APIKeyAuth(apiKeyHashed))

	router.GET(authGetPath, func(c *gin.Context) {
		c.String(http.StatusOK, "Authorized")
	})

	t.Run("Valid API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", authGetPath, nil)
		req.Header.Set("x-api-key", apiKey)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Authorized", w.Body.String())
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", authGetPath, nil)
		req.Header.Set("x-api-key", "wrongKey")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", authGetPath, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestXFCCAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	headerName := "X-Forwarded-Client-Cert"
	mustContain := "test"
	router.Use(XFCCAuth(headerName, mustContain))

	router.GET(authGetPath, func(c *gin.Context) {
		c.String(http.StatusOK, "Authorized")
	})

	t.Run("Valid XFCC", func(t *testing.T) {
		req, _ := http.NewRequest("GET", authGetPath, nil)
		req.Header.Set(headerName, "test-cert")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Authorized", w.Body.String())
	})

	t.Run("Invalid XFCC", func(t *testing.T) {
		req, _ := http.NewRequest("GET", authGetPath, nil)
		req.Header.Set(headerName, "invalid-cert")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing XFCC", func(t *testing.T) {
		req, _ := http.NewRequest("GET", authGetPath, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestHandleAuth(t *testing.T) {
	testCases := []struct {
		name            string
		apiKeyEnabled   bool
		xfccEnabled     bool
		apiKey          string
		xfccHeaderName  string
		xfccMustContain string
		setupRequest    func(*http.Request)
		expectedStatus  int
	}{
		{
			name:            "Both API Key and XFCC Enabled",
			apiKeyEnabled:   true,
			xfccEnabled:     true,
			apiKey:          apiKeyHashed,
			xfccHeaderName:  "X-Forwarded-Client-Cert",
			xfccMustContain: "changeme",
			setupRequest: func(req *http.Request) {
				req.Header.Set("x-api-key", apiKey)
				req.Header.Set("X-Forwarded-Client-Cert", "changeme")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Only API Key Enabled",
			apiKeyEnabled: true,
			xfccEnabled:   false,
			apiKey:        apiKeyHashed,
			setupRequest: func(req *http.Request) {
				req.Header.Set("x-api-key", apiKey)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "Only XFCC Enabled",
			apiKeyEnabled:   false,
			xfccEnabled:     true,
			xfccHeaderName:  "X-Forwarded-Client-Cert",
			xfccMustContain: "changeme",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-Forwarded-Client-Cert", "changeme")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Both Disabled",
			apiKeyEnabled:  false,
			xfccEnabled:    false,
			setupRequest:   func(req *http.Request) {},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API Key Enabled but Missing",
			apiKeyEnabled:  true,
			xfccEnabled:    false,
			apiKey:         apiKeyHashed,
			setupRequest:   func(req *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:            "XFCC Enabled but Missing",
			apiKeyEnabled:   false,
			xfccEnabled:     true,
			xfccHeaderName:  "X-Forwarded-Client-Cert",
			xfccMustContain: "changeme",
			setupRequest:    func(req *http.Request) {},
			expectedStatus:  http.StatusUnauthorized,
		},
		{
			name:            "Both Enabled but API Key Missing",
			apiKeyEnabled:   true,
			xfccEnabled:     true,
			apiKey:          apiKeyHashed,
			xfccHeaderName:  "X-Forwarded-Client-Cert",
			xfccMustContain: "changeme",
			setupRequest: func(req *http.Request) {
				req.Header.Set("X-Forwarded-Client-Cert", "changeme")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:            "Both Enabled but XFCC Missing",
			apiKeyEnabled:   true,
			xfccEnabled:     true,
			apiKey:          apiKeyHashed,
			xfccHeaderName:  "X-Forwarded-Client-Cert",
			xfccMustContain: "changeme",
			setupRequest: func(req *http.Request) {
				req.Header.Set("x-api-key", apiKey)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	dbName := fmt.Sprintf("%s_%s", uuid.New().String(), testDBName)
	mockDB := getMockDB(dbName)
	gin.SetMode(gin.TestMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			ctrl := &Control{
				db:              mockDB,
				apiKeyEnabled:   tc.apiKeyEnabled,
				apiKey:          tc.apiKey,
				xfccEnabled:     tc.xfccEnabled,
				xfccHeaderName:  tc.xfccHeaderName,
				xfccMustContain: tc.xfccMustContain,
			}
			getPath := fmt.Sprintf("/v1/workflow/%s", authTestName)
			v1 := router.Group("/v1")
			handleAuth(v1, ctrl)
			v1.GET("workflow/:name", ctrl.getWorkflow)

			req, _ := http.NewRequest("GET", getPath, nil)
			tc.setupRequest(req)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, true, strings.Contains(w.Body.String(), authTestName))
			}
		})
	}
	cleanupDB(mockDB, dbName)
}
