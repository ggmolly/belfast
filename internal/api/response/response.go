package response

type Payload struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Success(data interface{}) Payload {
	return Payload{
		OK:   true,
		Data: data,
	}
}

func Error(code, message string, details interface{}) Payload {
	return Payload{
		OK: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
