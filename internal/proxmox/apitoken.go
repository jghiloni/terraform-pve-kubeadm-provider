package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
)

func (a *APIClient) createToken(ctx context.Context, username string, password string) error {
	ticket, err := a.getTicket(username, password)
	if err != nil {
		return err
	}

	tokenName, _ := uuid.GenerateUUID()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/access/users/%s/token/%s", a.baseURL, username, tokenName), nil)
	if err != nil {
		return err
	}
	req.Header.Set("CSRFPreventionToken", ticket.CSRFPreventionToken)
	req.AddCookie(&http.Cookie{
		Name:  "PVEAuthCookie",
		Value: ticket.Ticket,
	})

	hc := &http.Client{
		Transport: a.delegateRoundTripper,
	}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var response APIReponse[TokenResponse]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	tokenResponse := response.Data
	a.setToken(tokenResponse.String())
	a.renewToken(ctx, tokenResponse.Info)

	return nil
}

func (a *APIClient) renewToken(ctx context.Context, info TokenInfo) {
	if info.Expire > 0 {
		go func() {
			expireTime := time.Unix(info.Expire, 0)
			waitDuration := time.Until(expireTime) - (10 * time.Second)

			select {
			case <-time.After(waitDuration):
				a.updateToken(ctx)
				return
			case <-ctx.Done():
				return
			}
		}()
	}
}

func (a *APIClient) updateToken(ctx context.Context) {
	token := a.token()
	parts := strings.Split(token, "!")
	username := parts[0]
	valparts := strings.Split(parts[1], "=")
	tokenid := valparts[0]

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/access/users/%s/token/%s", a.baseURL, username, tokenid), nil)
	if err != nil {
		return
	}

	resp, err := a.hc.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var response APIReponse[TokenInfo]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return
	}

	a.renewToken(ctx, response.Data)
}

func (a *APIClient) getTicket(username string, password string) (APITicket, error) {
	postBody := url.Values{}
	postBody.Set("username", username)
	postBody.Set("password", password)

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/access/ticket", a.baseURL), strings.NewReader(postBody.Encode()))
	if err != nil {
		return APITicket{}, err
	}

	req.Header.Set("Content-Type", FormContentType)
	hc := &http.Client{
		Transport: a.delegateRoundTripper,
	}
	resp, err := hc.Do(req)
	if err != nil {
		return APITicket{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return APITicket{}, fmt.Errorf("http status error: %s", resp.Status)
	}
	defer resp.Body.Close()

	var response APIReponse[APITicket]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return APITicket{}, err
	}

	return response.Data, nil
}

func (a *APIClient) absoluteURL(relativeURL string) string {
	return a.baseURL.JoinPath(strings.Split(relativeURL, "/")...).String()
}
