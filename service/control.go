// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/metrics"
	"mce.salesforce.com/sprinkler/model"
)

type Control struct {
	db              *gorm.DB
	address         string
	trustedProxies  []string
	apiKeyEnabled   bool
	apiKey          string
	xfccEnabled     bool
	xfccHeaderName  string
	xfccMustContain string
}

type putWorkflowReq struct {
	Name                 string    `json:"name" binding:"required"`
	Artifact             string    `json:"artifact" binding:"required"`
	Command              string    `json:"command" binding:"required"`
	Every                string    `json:"every" binding:"required"`
	NextRuntime          time.Time `json:"nextRuntime" binding:"required"`
	Backfill             bool      `json:"backfill"` // default false if absent
	Owner                *string   `json:"owner"`
	IsActive             bool      `json:"isActive"` // default false if absent
	ScheduleDelayMinutes uint      `json:"scheduleDelayMinutes"`
}

type deleteWorkflowReq struct {
	Name string `json:"name" binding:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewControl(db *gorm.DB, address string, trustedProxies []string, apiKeyEnabled bool, apiKey string, xfccEnabled bool, xfccHeaderName string, xfccMustContain string) *Control {
	return &Control{
		db:              db,
		address:         address,
		trustedProxies:  trustedProxies,
		apiKeyEnabled:   apiKeyEnabled,
		apiKey:          apiKey,
		xfccEnabled:     xfccEnabled,
		xfccHeaderName:  xfccHeaderName,
		xfccMustContain: xfccMustContain,
	}
}

func (ctrl *Control) putWorkflow(c *gin.Context) {
	var body putWorkflowReq
	if err := c.BindJSON(&body); err != nil {
		// bad request
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	every, err := model.ParseEvery(body.Every)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	wf := table.Workflow{
		Name:                 body.Name,
		Artifact:             body.Artifact,
		Command:              body.Command,
		Every:                every,
		NextRuntime:          body.NextRuntime,
		Backfill:             body.Backfill,
		Owner:                body.Owner,
		IsActive:             body.IsActive,
		ScheduleDelayMinutes: body.ScheduleDelayMinutes,
	}
	// upsert workflow
	ctrl.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"updated_at", "artifact", "command", "every", "next_runtime", "backfill", "owner", "is_active", "schedule_delay_minutes"}),
		}).Create(&wf)
	ctrl.db.Unscoped().Model(&wf).Update("deleted_at", nil)
	c.JSON(http.StatusOK, "OK")
}

func (ctrl *Control) deleteWorkflow(c *gin.Context) {
	var body deleteWorkflowReq
	if err := c.BindJSON(&body); err != nil {
		// bad request
		c.JSON(http.StatusBadRequest, gin.H{"message": "could not parse body"})
		return
	}

	dbRes := ctrl.db.Where("name = ?", body.Name).Delete(&table.Workflow{})
	if dbRes.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"name:": body.Name})
		return
	}
	if dbRes.Error == nil {
		c.JSON(http.StatusOK, gin.H{"name:": body.Name})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"name:": body.Name, "error": dbRes.Error})
	}
}

func (ctrl *Control) getWorkflow(c *gin.Context) {
	name := c.Param("name")
	var workflow table.Workflow
	dbRes := ctrl.db.Model(&table.Workflow{}).
		Where("name = ?", name).
		Find(&workflow)

	if dbRes.Error != nil || dbRes.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"Workflow not found:": fmt.Sprintf("name=%s", name)})
	} else {
		resp := putWorkflowReq{
			Name:                 workflow.Name,
			Artifact:             workflow.Artifact,
			Command:              workflow.Command,
			Every:                workflow.Every.String(),
			NextRuntime:          workflow.NextRuntime,
			Backfill:             workflow.Backfill,
			Owner:                workflow.Owner,
			IsActive:             workflow.IsActive,
			ScheduleDelayMinutes: workflow.ScheduleDelayMinutes}

		c.IndentedJSON(http.StatusOK, resp)
	}
}

// getWorkflows handles GET /v1/workflows
// Query parameters:
//   - orderBy: field to sort by (default: "name")
//   - orderDir: sort direction ("asc" or "desc", default: "asc")
//   - page: page number (default: 1)
//   - limit: items per page (default: 50)
//   - like: name filter pattern
func (ctrl *Control) getWorkflows(c *gin.Context) {

	start := time.Now()
	// Get query parameters for ordering
	orderBy := c.DefaultQuery("orderBy", "name")
	orderDir := c.DefaultQuery("orderDir", "asc")

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	// Get filtering parameters
	likePattern := c.Query("like")

	// Build the query
	query := ctrl.db.Model(&table.Workflow{})

	// Apply name filtering if like pattern is provided
	if likePattern != "" {
		// Validate like pattern to prevent SQL injection and ensure valid characters
		validCharsPattern := regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
		if !validCharsPattern.MatchString(likePattern) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_like_pattern",
				Code:    "400",
				Message: "like pattern can only contain letters, numbers, underscore (_), and dot (.)",
			})
			return
		}
		query = query.Where("name LIKE ?", "%"+likePattern+"%")
	}

	// Create a new session for the count query
	var total int64
	countQuery := query.Session(&gorm.Session{})
	countQuery.Count(&total)

	// Validate order direction
	if orderDir != "asc" && orderDir != "desc" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_order_direction",
			Code:    "400",
			Message: "orderDir must be 'asc' or 'desc'",
		})
		return
	}

	// Parse pagination parameters
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_page_value",
			Code:    "400",
			Message: "page must be a positive integer",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_limit_value",
			Code:    "400",
			Message: "limit must be a positive integer"})
		return
	}

	// Map field names to database column names
	columnMap := map[string]string{
		"name":                 "name",
		"nextRuntime":          "next_runtime",
		"isActive":             "is_active",
		"owner":                "owner",
		"scheduleDelayMinutes": "schedule_delay_minutes",
		"artifact":             "artifact",
		"command":              "command",
		"backfill":             "backfill",
	}

	// Get the database column name
	dbColumn, ok := columnMap[orderBy]
	if !ok {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_order_by_field",
			Code:    "400",
			Message: fmt.Sprintf("Invalid orderBy field: %s", orderBy)})
		return
	}

	// Calculate pagination
	offset := (page - 1) * limit

	// Apply ordering
	query = query.Order(fmt.Sprintf("%s %s", dbColumn, orderDir))

	// Apply pagination
	query = query.Offset(offset).Limit(limit)

	// handle soft deletes
	query = query.Unscoped().Where("deleted_at IS NULL")

	// Apply context timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	query = query.WithContext(ctx)

	// Execute query
	var workflows []table.Workflow
	result := query.Find(&workflows)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	// Convert to response format
	var response []putWorkflowReq
	for _, workflow := range workflows {
		response = append(response, putWorkflowReq{
			Name:                 workflow.Name,
			Artifact:             workflow.Artifact,
			Command:              workflow.Command,
			Every:                workflow.Every.String(),
			NextRuntime:          workflow.NextRuntime,
			Backfill:             workflow.Backfill,
			Owner:                workflow.Owner,
			IsActive:             workflow.IsActive,
			ScheduleDelayMinutes: workflow.ScheduleDelayMinutes,
		})
	}

	// Return response with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"limit":      limit,
			"totalPages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})

	metrics.UpdateHistogram("http_request_duration_seconds", time.Since(start), map[string]string{"route": "get_workflows"})
	metrics.IncrementCounter("http_requests_total", map[string]string{"route": "get_workflows"})
}

func APIKeyAuth(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		k := c.GetHeader("x-api-key")
		kSha := sha256.Sum256([]byte(k))
		kHex := hex.EncodeToString(kSha[:])
		if kHex != key {
			fmt.Println("API key mismatch")
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func XFCCAuth(headerName, mustContain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		xfcc := c.GetHeader(headerName)
		if xfcc == "" || (mustContain != "" && !strings.Contains(xfcc, mustContain)) {
			fmt.Println("XFCC header mismatch")
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func handleAuth(v *gin.RouterGroup, ctrl *Control) {
	if ctrl.apiKeyEnabled && ctrl.xfccEnabled {
		fmt.Println("API key and XFCC auth enabled, using both")
		v.Use(APIKeyAuth(ctrl.apiKey))
		v.Use(XFCCAuth(ctrl.xfccHeaderName, ctrl.xfccMustContain))
	} else if ctrl.apiKeyEnabled {
		fmt.Println("API key auth enabled")
		v.Use(APIKeyAuth(ctrl.apiKey))
	} else if ctrl.xfccEnabled {
		fmt.Println("XFCC auth enabled")
		v.Use(XFCCAuth(ctrl.xfccHeaderName, ctrl.xfccMustContain))
	} else {
		fmt.Println("No auth enabled")
	}
}

func (ctrl *Control) Run() {
	r := gin.Default()

	if err := r.SetTrustedProxies(ctrl.trustedProxies); err != nil {
		log.Fatal(err)
	}

	r.Use(gin.Recovery())
	r.Use(metrics.GinMiddleware)

	v1 := r.Group("/v1")
	handleAuth(v1, ctrl)
	{
		v1.PUT("/workflow", ctrl.putWorkflow)
		v1.DELETE("/workflow", ctrl.deleteWorkflow)
		v1.GET("/workflow/:name", ctrl.getWorkflow)
		v1.GET("/workflows", ctrl.getWorkflows)
	}

	r.GET("__status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"clientIP": c.ClientIP(),
			"status":   "ok",
		})
	})
	r.GET("__metrics", metrics.GinMetricsHandler)

	if err := r.Run(ctrl.address); err != nil {
		log.Fatal(err)
	}
}
