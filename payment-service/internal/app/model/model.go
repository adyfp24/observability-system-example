package model

import "time"

type Response struct {
	Code    int         `json:"code"`
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type LogType string
type LogLevel string

const (
	LogTypeHTTP     LogType  = "http"
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

type LogEntry struct {
	Time                                                      time.Time
	Level                                                     LogLevel
	Type                                                      LogType
	Service, Action, Message, RequestID, IPAddress, UserAgent string
	UserID                                                    *int
	Metadata                                                  map[string]interface{}
}
type APIRequestLogEntry struct {
	Time                                                       time.Time
	RequestID, Service, Endpoint, HTTPMethod, IPAddress        string
	UserID                                                     *int
	RequestHeaders, RequestBody, ResponseHeaders, ResponseBody map[string]interface{}
	StatusCode                                                 int
	Duration                                                   int64
}
