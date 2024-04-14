package logger

import (
	"fmt"
	"os"
	"strings"
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

var (
	debugColor   = color.New(color.BgCyan, color.FgBlack, color.Bold)
	infoColor    = color.New(color.BgGreen, color.FgBlack, color.Bold)
	warnColor    = color.New(color.BgYellow, color.FgBlack, color.Bold)
	errorColor   = color.New(color.BgRed, color.FgBlack, color.Bold)
	unknownColor = color.New(color.BgWhite, color.FgBlack, color.Bold)
	timeColor    = color.New(color.FgHiMagenta, color.Bold)
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
	var colorCategory *color.Color
	var subcategoryColor *color.Color
	if category == "Server" {
		colorCategory = debugColor
	} else if category == "Client" {
		colorCategory = infoColor
	} else {
		colorCategory = warnColor
	}

	if strings.HasPrefix(subcategory, "SC") {
		subcategoryColor = infoColor
	} else if strings.HasPrefix(subcategory, "CS") {
		subcategoryColor = warnColor
	} else {
		subcategoryColor = unknownColor
	}

	switch level {
	case LOG_LEVEL_DEBUG:
		colorCategory = debugColor
	case LOG_LEVEL_INFO:
		colorCategory = infoColor
	case LOG_LEVEL_WARN:
		colorCategory = warnColor
	case LOG_LEVEL_ERROR:
		colorCategory = errorColor
	}

	fmt.Fprintf(os.Stderr, "%s |%s|%s| %s\n",
		timeColor.Sprint(getShortDateTime()),
		colorCategory.Sprintf(" %-11s", category),
		subcategoryColor.Sprintf(" %-11s", subcategory),
		description)
}
