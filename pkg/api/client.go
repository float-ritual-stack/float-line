package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/evanschultz/float-rw-client/pkg/models"
)

const (
	baseURL         = "https://readwise.io/api/v2"
	defaultPageSize = 100
)

type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:   token,
		baseURL: baseURL,
	}
}

func (c *Client) doRequest(method, path string, params url.Values) ([]byte, error) {
	return c.doRequestWithBody(method, path, params, nil)
}

func (c *Client) doRequestWithBody(method, path string, params url.Values, body interface{}) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, err
	}

	if params != nil {
		u.RawQuery = params.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) GetHighlights(params url.Values) (*models.HighlightList, error) {
	if params == nil {
		params = url.Values{}
	}
	if params.Get("page_size") == "" {
		params.Set("page_size", fmt.Sprintf("%d", defaultPageSize))
	}

	body, err := c.doRequest("GET", "/highlights/", params)
	if err != nil {
		return nil, err
	}

	var result models.HighlightList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetBooks(params url.Values) (*models.BookList, error) {
	if params == nil {
		params = url.Values{}
	}
	if params.Get("page_size") == "" {
		params.Set("page_size", fmt.Sprintf("%d", defaultPageSize))
	}

	body, err := c.doRequest("GET", "/books/", params)
	if err != nil {
		return nil, err
	}

	var result models.BookList
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetHighlight(id int) (*models.Highlight, error) {
	body, err := c.doRequest("GET", fmt.Sprintf("/highlights/%d/", id), nil)
	if err != nil {
		return nil, err
	}

	var result models.Highlight
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) UpdateHighlight(id int, update models.HighlightUpdate) (*models.Highlight, error) {
	body, err := c.doRequestWithBody("PATCH", fmt.Sprintf("/highlights/%d/", id), nil, update)
	if err != nil {
		return nil, err
	}

	var result models.Highlight
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
