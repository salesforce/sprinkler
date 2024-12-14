// Copyright (c) 2022, Salesforce, Inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package orchard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"mce.salesforce.com/sprinkler/database/table"
)

type OrchardClient interface {
	Create(*table.Workflow) (string, error)
	Activate(string) error
}

type OrchardRestClient struct {
	Host       string
	APIKeyName string
	APIKey     string
}

func (c OrchardRestClient) request(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKeyName != "" && c.APIKey != "" {
		req.Header.Set(c.APIKeyName, c.APIKey)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid http code %d", rsp.StatusCode)
	}
	return rsp, nil
}

func (c OrchardRestClient) Create(wf table.Workflow) ([]string, error) {
	runner := OrchardStdoutRunner{}
	results, err := runner.Generate(wf.Artifact, wf.Command)
	if err != nil {
		log.Printf("OrchardRestClient Create > Generate error: %v\n", err)
		return []string{}, err
	}
	createdIDs := []string{}
	for _, result := range results {
		orchardID, err := c.create(result)
		if err != nil {
			return createdIDs, err
		}
		createdIDs = append(createdIDs, orchardID)
	}
	return createdIDs, nil
}

func (c OrchardRestClient) create(result string) (string, error) {
	url := fmt.Sprintf("%s/v1/workflow", c.Host)
	body := bytes.NewBuffer([]byte(result))
	rsp, err := c.request(http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	rawJson, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	var orchardID string
	err = json.Unmarshal(rawJson, &orchardID)
	if err != nil {
		return "", err
	}
	return string(orchardID), nil
}

func (c OrchardRestClient) IsActivated(orchardID string) (bool, error) {
	resp, err := c.Details(orchardID)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("workflow %s is not found", orchardID)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	details, err := ParseDetails(body)
	if err != nil {
		return false, err
	}
	if details.Status == "pending" {
		return false, fmt.Errorf("workflow %s is in `pending` status", orchardID)
	}
	return true, nil
}

func (c OrchardRestClient) Details(orchardID string) (*http.Response, error) {
	url := fmt.Sprintf("%s/v1/workflow/%s/details", c.Host, orchardID)
	return c.request(http.MethodGet, url, nil)
}

func (c OrchardRestClient) Activate(orchardID string) error {
	url := fmt.Sprintf("%s/v1/workflow/%s/activate", c.Host, orchardID)
	_, err := c.request(http.MethodPut, url, nil)
	return err
}

func (c OrchardRestClient) Cancel(orchardID string) error {
	url := fmt.Sprintf("%s/v1/workflow/%s/cancel", c.Host, orchardID)
	_, err := c.request(http.MethodPut, url, nil)
	return err
}

func (c OrchardRestClient) Delete(orchardID string) error {
	url := fmt.Sprintf("%s/v1/workflow/%s", c.Host, orchardID)
	_, err := c.request(http.MethodDelete, url, nil)
	return err
}

type FakeOrchardClient struct {
}

func (c FakeOrchardClient) Create(wf table.Workflow) ([]string, error) {
	runner := OrchardStdoutRunner{}
	results, err := runner.Generate(wf.Artifact, wf.Command)
	if err != nil {
		return []string{""}, err
	}
	log.Println("generating workflow", results)
	time.Sleep(1 * time.Second)
	return []string{fmt.Sprintf("wf-%s", uuid.New().String())}, nil
}

func (c FakeOrchardClient) Activate(wfID string) error {
	time.Sleep(1 * time.Second)
	return nil
}
