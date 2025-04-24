// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
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

func (ctrl *Control) getWorkflows(c *gin.Context) {
	// Get query parameters for ordering
	orderBy := c.DefaultQuery("orderBy", "name")
	orderDir := c.DefaultQuery("orderDir", "asc")

	// Validate order direction
	if orderDir != "asc" && orderDir != "desc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderDir must be 'asc' or 'desc'"})
		return
	}

	// Map frontend field names to database column names
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
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid orderBy field: %s", orderBy)})
		return
	}

	// Build the query
	query := ctrl.db.Model(&table.Workflow{})

	// Apply ordering
	query = query.Order(fmt.Sprintf("%s %s", dbColumn, orderDir))

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

	c.JSON(http.StatusOK, response)
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
