package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	LOG_LEVEL_AUTO  = iota
	LOG_LEVEL_DEBUG = iota
	LOG_LEVEL_INFO  = iota
	LOG_LEVEL_WARN  = iota
	LOG_LEVEL_ERROR = iota
)

type Field struct {
	Key   string
	Value string
}

type Entry struct {
	Scope  string
	Fields []Field
}

var (
	logLevelOnce sync.Once
	logLevel     = LOG_LEVEL_INFO
)

func Scope(scope string) Entry {
	return Entry{Scope: scope}
}

func WithFields(scope string, fields ...Field) Entry {
	return Entry{Scope: scope, Fields: fields}
}

func FieldValue(key string, value any) Field {
	return Field{Key: key, Value: fmt.Sprint(value)}
}

func PacketFields(packetID uint32, direction string) []Field {
	return []Field{
		{Key: "packet", Value: fmt.Sprint(packetID)},
		{Key: "dir", Value: direction},
	}
}

func CommanderFields(commanderID uint32, accountID uint32) []Field {
	return []Field{
		{Key: "commander", Value: fmt.Sprint(commanderID)},
		{Key: "account", Value: fmt.Sprint(accountID)},
	}
}

func Span(scope string, fields ...Field) func(err error) {
	start := time.Now()
	return func(err error) {
		allFields := append([]Field{}, fields...)
		allFields = append(allFields, Field{Key: "duration", Value: time.Since(start).String()})
		if err != nil {
			WithFields(scope, allFields...).Error(err.Error())
			return
		}
		WithFields(scope, allFields...).Info("completed")
	}
}

func (e Entry) With(fields ...Field) Entry {
	e.Fields = append(append([]Field{}, e.Fields...), fields...)
	return e
}

func (e Entry) Debug(message string) {
	Log(e.Scope, message, LOG_LEVEL_DEBUG, e.Fields...)
}

func (e Entry) Info(message string) {
	Log(e.Scope, message, LOG_LEVEL_INFO, e.Fields...)
}

func (e Entry) Warn(message string) {
	Log(e.Scope, message, LOG_LEVEL_WARN, e.Fields...)
}

func (e Entry) Error(message string) {
	Log(e.Scope, message, LOG_LEVEL_ERROR, e.Fields...)
}

func Log(scope, message string, level int, fields ...Field) {
	if !shouldLog(level) {
		return
	}
	line := formatLine(scope, message, level, fields...)
	fmt.Fprintln(os.Stdout, line)
}

func shouldLog(level int) bool {
	logLevelOnce.Do(func() {
		switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
		case "debug":
			logLevel = LOG_LEVEL_DEBUG
		case "info":
			logLevel = LOG_LEVEL_INFO
		case "warn", "warning":
			logLevel = LOG_LEVEL_WARN
		case "error":
			logLevel = LOG_LEVEL_ERROR
		default:
			logLevel = LOG_LEVEL_INFO
		}
	})
	return level >= logLevel
}

func formatLine(scope, message string, level int, fields ...Field) string {
	levelLabel := levelName(level)
	levelColor := levelColor(level)
	coloredTime := timeColor.Sprint(getShortDateTime())
	coloredLevel := levelColor.Sprint(levelLabel)
	coloredScope := scopeColor.Sprint(scope)
	formattedFields := formatFields(fields...)
	coloredFields := contextColor.Sprint(formattedFields)
	coloredMessage := messageColor.Sprint(message)
	if formattedFields == "" {
		return fmt.Sprintf("%s | %s | %s | %s", coloredTime, coloredLevel, coloredScope, coloredMessage)
	}
	return fmt.Sprintf("%s | %s | %s | %s | %s", coloredTime, coloredLevel, coloredScope, coloredFields, coloredMessage)
}

func levelName(level int) string {
	switch level {
	case LOG_LEVEL_DEBUG:
		return "DEBUG"
	case LOG_LEVEL_INFO:
		return "INFO"
	case LOG_LEVEL_WARN:
		return "WARN"
	case LOG_LEVEL_ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func levelColor(level int) *color.Color {
	switch level {
	case LOG_LEVEL_DEBUG:
		return debugColor
	case LOG_LEVEL_INFO:
		return infoColor
	case LOG_LEVEL_WARN:
		return warnColor
	case LOG_LEVEL_ERROR:
		return errorColor
	default:
		return unknownColor
	}
}

func formatFields(fields ...Field) string {
	if len(fields) == 0 {
		return ""
	}
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.Key == "" {
			continue
		}
		value := field.Value
		if value == "" {
			value = "-"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", field.Key, value))
	}
	return strings.Join(parts, " ")
}

var (
	debugColor   = color.New(color.FgHiBlue, color.Bold)
	infoColor    = color.New(color.FgHiGreen, color.Bold)
	warnColor    = color.New(color.FgHiYellow, color.Bold)
	errorColor   = color.New(color.FgHiRed, color.Bold)
	unknownColor = color.New(color.FgHiWhite, color.Bold)
	timeColor    = color.New(color.FgHiBlack)
	scopeColor   = color.New(color.FgHiCyan, color.Bold)
	contextColor = color.New(color.FgHiBlack)
	messageColor = color.New(color.FgWhite, color.Bold)
)

// Get current time in dd/mm/yyyy hh:mm:ss format
func getShortDateTime() string {
	return fmt.Sprintf("%02d/%02d/%04d %02d:%02d:%02d",
		time.Now().Day(),
		time.Now().Month(),
		time.Now().Year(),
		time.Now().Hour(),
		time.Now().Minute(),
		time.Now().Second())
}

func LogEvent(category, subcategory, description string, level int) {
	fields := []Field{}
	if subcategory != "" {
		fields = append(fields, Field{Key: "event", Value: subcategory})
	}
	scope := category
	if subcategory != "" {
		scope = fmt.Sprintf("%s/%s", category, subcategory)
	}
	Log(scope, description, level, fields...)
}
