package orchard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Host string
}

func (c OrchardRestClient) Create(wf table.Workflow) (string, error) {
	runner := OrchardStdoutRunner{}
	result, err := runner.Generate(wf.Artifact, wf.Command)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/v1/workflow", c.Host)
	rsp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(result)))
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Invalid http code %d", rsp.StatusCode)
	}
	rawJson, err := ioutil.ReadAll(rsp.Body)
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

func (c OrchardRestClient) Activate(orchardID string) error {
	url := fmt.Sprintf("%s/v1/workflow/%s/activate", c.Host, orchardID)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid http code %d", rsp.StatusCode)
	}
	return nil
}

type FakeOrchardClient struct {
}

func (c FakeOrchardClient) Create(wf table.Workflow) (string, error) {
	runner := OrchardStdoutRunner{}
	result, err := runner.Generate(wf.Artifact, wf.Command)
	if err != nil {
		return "", err
	}
	log.Println("generating workflow", result)
	time.Sleep(1 * time.Second)
	return fmt.Sprintf("wf-%s", uuid.New().String()), nil
}

func (c FakeOrchardClient) Activate(wfID string) error {
	time.Sleep(1 * time.Second)
	return nil
}
