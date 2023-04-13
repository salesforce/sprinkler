// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package service

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mce.salesforce.com/sprinkler/database"
	"mce.salesforce.com/sprinkler/database/table"
	"mce.salesforce.com/sprinkler/model"
)

type Control struct {
	db             *gorm.DB
	address        string
	trustedProxies []string
	apiKey         string
}

type postWorkflowReq struct {
	Name        string    `json:"name" binding:"required"`
	Artifact    string    `json:"artifact" binding:"required"`
	Command     string    `json:"command" binding:"required"`
	Every       string    `json:"every" binding:"required"`
	NextRuntime time.Time `json:"nextRuntime" binding:"required"`
	Backfill    bool      `json:"backfill"` // default false if absent
	Owner       *string   `json:"owner"`
	IsActive    bool      `json:"isActive"` // default false if absent
}

type deleteWorkflowReq struct {
	Name     string `json:"name" binding:"required"`
	Artifact string `json:"artifact" binding:"required"`
}

func NewControl(address string, trustedProxies []string, apiKey string) *Control {
	return &Control{
		db:             database.GetInstance(),
		address:        address,
		trustedProxies: trustedProxies,
		apiKey:         apiKey,
	}
}

func (ctrl *Control) postWorkflow(c *gin.Context) {
	var body postWorkflowReq
	if err := c.BindJSON(&body); err != nil {
		// bad request
		fmt.Println(err)
		return
	}

	every, err := model.ParseEvery(body.Every)
	if err != nil {
		fmt.Println(err)
		return
	}

	wf := table.Workflow{
		Name:        body.Name,
		Artifact:    body.Artifact,
		Command:     body.Command,
		Every:       every,
		NextRuntime: body.NextRuntime,
		Backfill:    body.Backfill,
		Owner:       body.Owner,
		IsActive:    body.IsActive,
	}
	// upsert workflow
	ctrl.db.Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "artifact"}},
			DoUpdates: clause.AssignmentColumns([]string{"updated_at", "command", "every", "next_runtime", "backfill", "owner", "is_active"}),
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
	var wf = table.Workflow{}
	dbResult := ctrl.db.Where("name = ? and artifact = ?", body.Name, body.Artifact).First(&wf)
	if dbResult.Error == nil && dbResult.RowsAffected == 1 {
		ctrl.db.Where("workflow_id = ?", &wf.ID).Delete(&table.WorkflowSchedulerLock{})
		ctrl.db.Delete(&wf)
		c.JSON(http.StatusOK, gin.H{"name:": body.Name, "artifact": body.Artifact})
	} else if dbResult.Error != nil && errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"name:": body.Name, "artifact": body.Artifact})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"name:": body.Name, "artifact": body.Artifact, "error": dbResult.Error})
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

	v1 := r.Group("/v1")
	v1.Use(APIKeyAuth(ctrl.apiKey))
	{
		v1.POST("/workflow", ctrl.postWorkflow)
		v1.DELETE("/workflow", ctrl.deleteWorkflow)
	}

	r.GET("__status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"clientIP": c.ClientIP(),
			"status":   "ok",
		})
	})

	if err := r.Run(ctrl.address); err != nil {
		log.Fatal(err)
	}
}
