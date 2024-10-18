package spotify

import "fmt"

type ErrorResponse struct {
	Err struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"error"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("[%d] %s", e.Err.Status, e.Err.Message)
}
