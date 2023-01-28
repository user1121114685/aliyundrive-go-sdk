package aliyundrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Array []any
type Object map[string]any

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf(`{"code":"%v","message":"%v"}`, r.Code, r.Message)
}

func IsPreHashMatchedError(err error) bool {
	errResponse, ok := err.(*ErrorResponse)
	if !ok {
		return false
	}
	return errResponse.Code == "PreHashMatched"
}

func (c *Drive) toRequest(ctx context.Context, url string, params any) (*http.Request, error) {
	bodyData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(bodyData)
	request, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func (c *Drive) requestWithCredit(ctx context.Context, url string, params any) ([]byte, error) {
	accessToken, err := c.tokenManager.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	request, err := c.toRequest(ctx, url, params)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)
	return c.doRequest(request)
}

func (c *Drive) requestWithoutCredit(ctx context.Context, url string, params any) ([]byte, error) {
	request, err := c.toRequest(ctx, url, params)
	if err != nil {
		return nil, err
	}
	return c.doRequest(request)
}

func (c *Drive) doRequest(request *http.Request) ([]byte, error) {
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	result := new(ErrorResponse)
	err = json.Unmarshal(respData, result)
	if err != nil {
		return nil, err
	}

	if result.Code != "" || result.Message != "" {
		return nil, result
	}

	return respData, nil
}
