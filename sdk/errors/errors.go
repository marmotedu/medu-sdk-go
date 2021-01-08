package errors

import (
	"fmt"
)

type MEDUSDKError struct {
	Code      int
	Message   string
	RequestID string
}

func (err *MEDUSDKError) Error() string {
	return fmt.Sprintf("[MEDUSDKError] code=%d, message=%s, requestID=%s", err.Code, err.Message, err.RequestID)
}

func NewMEDUSDKError(code int, message, requestID string) error {
	return &MEDUSDKError{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}
}

func (err *MEDUSDKError) GetCode() int {
	return err.Code
}

func (err *MEDUSDKError) GetMessage() string {
	return err.Message
}

func (err *MEDUSDKError) GetRequestID() string {
	return err.RequestID
}
