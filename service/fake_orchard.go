package service

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FakeOrchard struct {
	mu        sync.Mutex
	Workflows map[string]WorkflowStatus
}

type WorkflowStatus struct {
	name   string
	status string
}

type orchardWorkflow struct {
	name string `json:"name"`
}

func NewFakeOrchard() *FakeOrchard {
	return &FakeOrchard{
		Workflows: make(map[string]WorkflowStatus),
	}
}

func (o *FakeOrchard) postWorkflow(c *gin.Context) {
	var workflow orchardWorkflow
	if err := c.BindJSON(&workflow); err != nil {
		return
	}
	o.mu.Lock()
	defer o.mu.Unlock()

	orchardId := fmt.Sprintf("wf-%s", uuid.New().String())
	o.Workflows[orchardId] = WorkflowStatus{
		name:   workflow.name,
		status: "pending",
	}
	c.JSON(http.StatusOK, orchardId)
}

func (o *FakeOrchard) activateWorkflow(c *gin.Context) {
	orchardId := c.Param("id")
	o.mu.Lock()
	defer o.mu.Unlock()
	if nameSts, ok := o.Workflows[orchardId]; ok {
		o.Workflows[orchardId] = WorkflowStatus{
			name:   nameSts.name,
			status: "activated",
		}
		c.JSON(http.StatusOK, orchardId)
	} else {
		c.JSON(http.StatusNotFound, "not exist")
	}
}

func (o *FakeOrchard) Run() {
	r := gin.Default()
	r.POST("v1/workflow", o.postWorkflow)
	r.PUT("v1/workflow/:id/activate", o.activateWorkflow)
	r.Run()
}
