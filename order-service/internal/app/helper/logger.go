package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/adyfp24/okejek-go-service/order-service/internal/app/model"
	"github.com/gofiber/fiber/v2"
)

type Logger struct {
	service     string
	logDir      string
	environment string
	mu          sync.Mutex
	bufPool     sync.Pool
	logChan     chan logMessage
}

type LoggerConfig struct {
	Service     string
	LogDir      string
	Environment string
	BufferSize  int
}

type logMessage struct {
	filename string
	data     interface{}
}

var (
	defaultLogger *Logger
)

func NewLogger(config LoggerConfig) *Logger {
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	absLogDir, err := filepath.Abs(config.LogDir)
	if err != nil {
		log.Printf("Warning: Could not get absolute path for %s: %v", config.LogDir, err)
		absLogDir = config.LogDir
	}

	if err := os.MkdirAll(absLogDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
	}

	logger := &Logger{
		service:     config.Service,
		logDir:      absLogDir,
		environment: config.Environment,
		bufPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		logChan: make(chan logMessage, config.BufferSize),
	}

	go logger.processLogs()

	log.Printf("Created new logger instance: service=%s, logDir=%s, environment=%s",
		logger.service, logger.logDir, logger.environment)

	return logger
}

func (l *Logger) processLogs() {
	for msg := range l.logChan {
		if err := l.writeLogSync(msg.filename, msg.data); err != nil {
			log.Printf("Failed to write log: %v", err)
		}
	}
}

func (l *Logger) getLogFileName(logType model.LogType) string {
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.log", string(logType), date)

	if l.environment == "production" {
		return filepath.Join(l.logDir, l.environment, filename)
	}
	return filepath.Join(l.logDir, filename)
}

func (l *Logger) writeLog(filename string, entry interface{}) error {
	select {
	case l.logChan <- logMessage{filename: filename, data: entry}:
		return nil
	default:
		return l.writeLogSync(filename, entry)
	}
}

func (l *Logger) writeLogSync(filename string, entry interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	buf := l.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufPool.Put(buf)

	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to write log: %v", err)
	}

	return nil
}

func (l *Logger) getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	return "unknown:0"
}

func (l *Logger) Log(ctx context.Context, level model.LogLevel, logType model.LogType, action, message string, metadata map[string]interface{}) {
	entry := model.LogEntry{
		Time:      time.Now(),
		Level:     level,
		Type:      logType,
		Service:   l.service,
		Action:    action,
		Message:   message,
		RequestID: l.getRequestID(ctx),
		UserID:    l.getUserID(ctx),
		IPAddress: l.getIPAddress(ctx),
		UserAgent: l.getUserAgent(ctx),
		Metadata:  metadata,
	}

	if level == model.LogLevelDebug || level == model.LogLevelError {
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]interface{})
		}
		entry.Metadata["caller"] = l.getCallerInfo()
	}

	filename := l.getLogFileName(logType)
	if err := l.writeLog(filename, entry); err != nil {
		log.Printf("Failed to write log: %v", err)
	}

	if l.environment == "development" {
		log.Printf("[%s] %s: %s", level, action, message)
	}
}

func (l *Logger) Debug(ctx context.Context, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.Log(ctx, model.LogLevelDebug, logType, action, message, metadata)
}

func (l *Logger) Info(ctx context.Context, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.Log(ctx, model.LogLevelInfo, logType, action, message, metadata)
}

func (l *Logger) Warning(ctx context.Context, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.Log(ctx, model.LogLevelWarning, logType, action, message, metadata)
}

func (l *Logger) Error(ctx context.Context, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.Log(ctx, model.LogLevelError, logType, action, message, metadata)
}

func (l *Logger) Fatal(ctx context.Context, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.Log(ctx, model.LogLevelFatal, logType, action, message, metadata)
	os.Exit(1)
}

func (l *Logger) LogAPIRequest(ctx context.Context, endpoint, method string, statusCode int, duration int64, requestHeaders, requestBody, responseHeaders, responseBody map[string]interface{}) {
	entry := model.APIRequestLogEntry{
		Time:            time.Now(),
		RequestID:       l.getRequestID(ctx),
		Service:         l.service,
		Endpoint:        endpoint,
		HTTPMethod:      method,
		IPAddress:       l.getIPAddress(ctx),
		UserID:          l.getUserID(ctx),
		RequestHeaders:  requestHeaders,
		RequestBody:     requestBody,
		ResponseHeaders: responseHeaders,
		ResponseBody:    responseBody,
		StatusCode:      statusCode,
		Duration:        duration,
	}

	filename := l.getLogFileName(model.LogTypeHTTP)
	if err := l.writeLog(filename, entry); err != nil {
		log.Printf("Failed to write API log: %v", err)
	}
}

func (l *Logger) getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func (l *Logger) getUserID(ctx context.Context) *int {
	if userID, ok := ctx.Value(userIDKey).(int); ok {
		return &userID
	}
	return nil
}

func (l *Logger) getIPAddress(ctx context.Context) string {
	if ip, ok := ctx.Value(ipAddressKey).(string); ok {
		return ip
	}
	return ""
}

func (l *Logger) getUserAgent(ctx context.Context) string {
	if ua, ok := ctx.Value(userAgentKey).(string); ok {
		return ua
	}
	return ""
}

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
	ipAddressKey contextKey = "ip_address"
	userAgentKey contextKey = "user_agent"
)

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func WithIPAddress(ctx context.Context, ipAddress string) context.Context {
	return context.WithValue(ctx, ipAddressKey, ipAddress)
}

func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, userAgentKey, userAgent)
}

func FiberContextToContext(c *fiber.Ctx) context.Context {
	ctx := context.Background()

	if requestID, ok := c.Locals("request_id").(string); ok {
		ctx = WithRequestID(ctx, requestID)
	}

	if userID, ok := c.Locals("user_id").(int); ok {
		ctx = WithUserID(ctx, userID)
	}

	ctx = WithIPAddress(ctx, c.IP())
	ctx = WithUserAgent(ctx, c.Get("User-Agent"))

	return ctx
}

func (l *Logger) LogFiber(c *fiber.Ctx, level model.LogLevel, logType model.LogType, action, message string, metadata map[string]interface{}) {
	ctx := FiberContextToContext(c)
	l.Log(ctx, level, logType, action, message, metadata)
}

func (l *Logger) DebugFiber(c *fiber.Ctx, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.LogFiber(c, model.LogLevelDebug, logType, action, message, metadata)
}

func (l *Logger) InfoFiber(c *fiber.Ctx, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.LogFiber(c, model.LogLevelInfo, logType, action, message, metadata)
}

func (l *Logger) WarningFiber(c *fiber.Ctx, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.LogFiber(c, model.LogLevelWarning, logType, action, message, metadata)
}

func (l *Logger) ErrorFiber(c *fiber.Ctx, logType model.LogType, action, message string, metadata map[string]interface{}) {
	l.LogFiber(c, model.LogLevelError, logType, action, message, metadata)
}
