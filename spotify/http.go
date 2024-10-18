package spotify

import (
	"encoding/json"
	"net/http"
)

func (c *Client) doGetRequest(url string, v any) error {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	return c.sendRequest(request, v)
}

func (c *Client) sendRequest(request *http.Request, v any) error {
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200:
		if err := json.NewDecoder(response.Body).Decode(&v); err != nil {
			return err
		}

		return nil
	case 204:
		c.logger.Info("Playback not available or active")
		return nil
	case 401, 403, 429:
		var errorResponse ErrorResponse
		if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
			return err
		}

		return errorResponse
	default:
		return nil
	}
}
