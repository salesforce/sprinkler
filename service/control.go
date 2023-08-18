// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/metrics"
	"mce.salesforce.com/sprinkler/model"
)

type Control struct {
	db             *gorm.DB
	address        string
	trustedProxies []string
	apiKey         string
}

type postWorkflowReq struct {
	Name                string    `json:"name" binding:"required"`
	Artifact            string    `json:"artifact" binding:"required"`
	Command             string    `json:"command" binding:"required"`
	Every               string    `json:"every" binding:"required"`
	NextRuntime         time.Time `json:"nextRuntime" binding:"required"`
	Backfill            bool      `json:"backfill"` // default false if absent
	Owner               *string   `json:"owner"`
	IsActive            bool      `json:"isActive"` // default false if absent
	StaggerStartMinutes uint      `json:"staggerStartMinutes"`
}

type deleteWorkflowReq struct {
	Name string `json:"name" binding:"required"`
}

func NewControl(address string, trustedProxies []string, apiKey string) *Control {
	return &Control{
		db:             database.GetInstance(),
		address:        address,
		trustedProxies: trustedProxies,
		apiKey:         apiKey,
	}
}

func (ctrl *Control) putWorkflow(c *gin.Context) {
	var body postWorkflowReq
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
		Name:                body.Name,
		Artifact:            body.Artifact,
		Command:             body.Command,
		Every:               every,
		NextRuntime:         body.NextRuntime,
		Backfill:            body.Backfill,
		Owner:               body.Owner,
		IsActive:            body.IsActive,
		StaggerStartMinutes: body.StaggerStartMinutes,
	}
	// upsert workflow
	ctrl.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"updated_at", "artifact", "command", "every", "next_runtime", "backfill", "owner", "is_active", "stagger_start_minutes"}),
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

func APIKeyAuth(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		k := c.GetHeader("x-api-key")
		if k != key {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
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
	v1.Use(APIKeyAuth(ctrl.apiKey))
	{
		v1.PUT("/workflow", ctrl.putWorkflow)
		v1.DELETE("/workflow", ctrl.deleteWorkflow)
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
