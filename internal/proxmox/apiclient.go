package proxmox

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	FormContentType = "application/x-www-form-urlencoded"
)

// APIClient is an implementation of http.RoundTripper with convenience methods
type APIClient struct {
	hc                   *http.Client
	delegateRoundTripper http.RoundTripper
	baseURL              *url.URL
	apiToken             string
	tokenMutex           *sync.RWMutex
	nodeName             string
}

func (a *APIClient) token() string {
	if a.tokenMutex == nil {
		a.tokenMutex = new(sync.RWMutex)
	}

	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()

	return a.apiToken
}

func (a *APIClient) setToken(newToken string) {
	if a.tokenMutex == nil {
		a.tokenMutex = new(sync.RWMutex)
	}

	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()

	a.apiToken = newToken
}

func (a *APIClient) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s", a.token()))
	return a.delegateRoundTripper.RoundTrip(req)
}

func NewAPIClient(ctx context.Context, hostname, nodename, username, password, token string, skipTLS bool) (*APIClient, error) {
	hostname = strings.ToLower(hostname)
	if strings.HasPrefix(hostname, "http://") || strings.HasPrefix(hostname, "https://") {
		parts := strings.Split(hostname, "/")
		hostname = parts[2]
	}

	if !strings.Contains(hostname, ":") {
		hostname = hostname + ":8006"
	}

	hc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLS,
			},
		},
	}

	baseURL, err := url.Parse(fmt.Sprintf("https://%s/api2/json", hostname))
	if err != nil {
		return nil, err
	}

	c := &APIClient{
		baseURL:              baseURL,
		nodeName:             nodename,
		tokenMutex:           &sync.RWMutex{},
		delegateRoundTripper: hc.Transport,
		hc:                   hc,
		apiToken:             token,
	}

	c.hc.Transport = c

	if token == "" {
		if username == "" || password == "" {
			return nil, errors.New("either token OR (username AND password) must be set")
		}
		c.createToken(ctx, username, password)
	}

	return c, nil
}
