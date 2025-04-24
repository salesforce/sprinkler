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

func TestGetWorkflows(t *testing.T) {
	dbName := fmt.Sprintf("%s_%s", uuid.New().String(), testDBName)
	mockDB := getMockDB(dbName)
	ctrl := &Control{db: mockDB}
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/workflows", ctrl.getWorkflows)

	// Create additional test workflows with different properties
	testWorkflows := []table.Workflow{
		{
			Name:                 "workflow_a",
			Artifact:             "artifact_a.jar",
			Command:              "java -jar a.jar",
			Every:                model.Every{1, model.EveryDay},
			NextRuntime:          time.Now().Add(24 * time.Hour),
			Backfill:             false,
			Owner:                stringPtr("team_a"),
			IsActive:             true,
			ScheduleDelayMinutes: 5,
		},
		{
			Name:                 "workflow_b",
			Artifact:             "artifact_b.jar",
			Command:              "java -jar b.jar",
			Every:                model.Every{2, model.EveryDay},
			NextRuntime:          time.Now().Add(48 * time.Hour),
			Backfill:             true,
			Owner:                stringPtr("team_b"),
			IsActive:             false,
			ScheduleDelayMinutes: 10,
		},
		{
			Name:                 "workflow_c",
			Artifact:             "artifact_c.jar",
			Command:              "java -jar c.jar",
			Every:                model.Every{1, model.EveryWeek},
			NextRuntime:          time.Now().Add(72 * time.Hour),
			Backfill:             false,
			Owner:                stringPtr("team_c"),
			IsActive:             true,
			ScheduleDelayMinutes: 15,
		},
		{
			Name:                 "test4_workflows",
			Artifact:             "test_artifact.jar",
			Command:              "java -jar test.jar",
			Every:                model.Every{1, model.EveryDay},
			NextRuntime:          time.Now().Add(96 * time.Hour),
			Backfill:             false,
			Owner:                stringPtr("test_team"),
			IsActive:             true,
			ScheduleDelayMinutes: 20,
		},
	}

	// Insert test workflows
	for _, wf := range testWorkflows {
		result := mockDB.Create(&wf)
		assert.NoError(t, result.Error, "Failed to create test workflow: %v", wf.Name)
	}

	t.Run("Default ordering by name", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 3)

		// Check that workflows are ordered by name
		workflows := make([]putWorkflowReq, len(data))
		for i, item := range data {
			workflowJSON, _ := json.Marshal(item)
			json.Unmarshal(workflowJSON, &workflows[i])
		}

		for i := 1; i < len(workflows); i++ {
			assert.LessOrEqual(t, workflows[i-1].Name, workflows[i].Name)
		}
	})

	t.Run("Order by nextRuntime ascending", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?orderBy=nextRuntime&orderDir=asc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 3)

		// Check that workflows are ordered by nextRuntime ascending
		workflows := make([]putWorkflowReq, len(data))
		for i, item := range data {
			workflowJSON, _ := json.Marshal(item)
			json.Unmarshal(workflowJSON, &workflows[i])
		}

		for i := 1; i < len(workflows); i++ {
			assert.LessOrEqual(t, workflows[i-1].NextRuntime, workflows[i].NextRuntime)
		}
	})

	t.Run("Order by isActive descending", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?orderBy=isActive&orderDir=desc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 3)

		// Check that workflows are ordered by isActive descending
		workflows := make([]putWorkflowReq, len(data))
		for i, item := range data {
			workflowJSON, _ := json.Marshal(item)
			json.Unmarshal(workflowJSON, &workflows[i])
		}

		// Verify that active workflows come before inactive ones
		foundInactive := false
		for _, wf := range workflows {
			if !wf.IsActive {
				foundInactive = true
			} else if foundInactive {
				t.Error("Found active workflow after inactive workflow")
			}
		}
	})

	t.Run("Order by scheduleDelayMinutes ascending", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?orderBy=scheduleDelayMinutes&orderDir=asc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 3)

		// Check that workflows are ordered by scheduleDelayMinutes ascending
		workflows := make([]putWorkflowReq, len(data))
		for i, item := range data {
			workflowJSON, _ := json.Marshal(item)
			json.Unmarshal(workflowJSON, &workflows[i])
		}

		for i := 1; i < len(workflows); i++ {
			assert.LessOrEqual(t, workflows[i-1].ScheduleDelayMinutes, workflows[i].ScheduleDelayMinutes)
		}
	})

	t.Run("Invalid order direction", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?orderBy=name&orderDir=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid orderBy field", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?orderBy=invalidField&orderDir=asc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Message, "Invalid orderBy field")
	})

	t.Run("Filter by name pattern", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?like=test4", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)

		// Should only find the test_workflow
		assert.Equal(t, 1, len(data), "Expected exactly one workflow with 'test' in the name")

		workflow := make(map[string]interface{})
		workflowJSON, _ := json.Marshal(data[0])
		json.Unmarshal(workflowJSON, &workflow)

		assert.Contains(t, workflow["name"], "test", "Found workflow name does not contain 'test'")
	})

	t.Run("Pagination", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?page=1&limit=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].([]interface{})
		assert.True(t, ok)
		assert.Equal(t, 2, len(data))

		pagination, ok := response["pagination"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(1), pagination["page"])
		assert.Equal(t, float64(2), pagination["limit"])
		assert.GreaterOrEqual(t, pagination["total"], float64(4))
	})

	t.Run("Invalid page parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?page=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Message, "page must be a positive integer")
	})

	t.Run("Invalid limit parameter", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/workflows?limit=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response.Message, "limit must be a positive integer")
	})

	cleanupDB(mockDB, dbName)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
