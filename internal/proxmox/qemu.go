package proxmox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gideaworx.io/go-encoding/urlvalues"
)

func (a *APIClient) nextID() (int, error) {
	var response APIReponse[IntOrString]

	resp, err := a.hc.Get(a.absoluteURL("/cluster/next"))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("http status error: expected: 200 OK, actual: %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	return int(response.Data), nil
}

func (a *APIClient) CloneVM(vmTemplateID int, linked bool, sync bool) (int, error) {
	nextID, err := a.nextID()
	if err != nil {
		return 0, err
	}

	request := url.Values{}
	request.Set("newid", fmt.Sprintf("%d", nextID))

	full := "1"
	if linked {
		full = "0"
	}

	request.Set("full", full)
	req, err := http.NewRequest(http.MethodPost,
		a.absoluteURL(fmt.Sprintf("nodes/%s/qemu/%d/clone", a.nodeName, vmTemplateID)),
		strings.NewReader(request.Encode()))
	if err != nil {
		return 0, err
	}

	resp, err := a.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return 0, fmt.Errorf("http status error: expected 2xx, actual %s", resp.Status)
	}

	var response APIReponse[string]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, err
	}

	taskID := response.Data

	if !sync {
		return nextID, nil
	}

	pollTimeout := 2 * time.Minute
	for {
		resp, err := a.hc.Get(a.absoluteURL(fmt.Sprintf("nodes/%s/tasks/%s/status", a.nodeName, taskID)))
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("http status error: expected 200 OK, actual %s", resp.Status)
		}

		var statusResponse APIReponse[TaskSummary]
		if err = json.NewDecoder(resp.Body).Decode(&statusResponse); err != nil {
			return 0, err
		}

		summary := statusResponse.Data
		if summary.Status == "stopped" {
			if summary.ExitStatus != "OK" {
				return 0, fmt.Errorf("expected exit status OK, got %q", summary.ExitStatus)
			}

			return nextID, nil
		}

		select {
		case <-time.After(pollTimeout):
			return 0, fmt.Errorf("timed out after %s", pollTimeout.String())
		case <-time.After(5 * time.Second):
			continue
		}
	}
}

func (a *APIClient) UpdateConfig(vmid int, config MachineConfig) error {
	url := a.absoluteURL(fmt.Sprintf("nodes/%s/qemu/%d/config", a.nodeName, vmid))
	body, err := urlvalues.MarshalURLValues(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}

	resp, err := a.hc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got %s", resp.Status)
	}

	return nil
}

func (a *APIClient) StartVM(vmid int) error {
	url := a.absoluteURL(fmt.Sprintf("nodes/%s/qemu/%d/status/start", a.nodeName, vmid))
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := a.hc.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status error: expected 200 OK, got %s", resp.Status)
	}

	return nil
}
