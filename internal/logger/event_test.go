package logger

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestScope(t *testing.T) {
	entry := Scope("TestScope")

	if entry.Scope != "TestScope" {
		t.Fatalf("expected scope 'TestScope', got %q", entry.Scope)
	}
	if len(entry.Fields) != 0 {
		t.Fatalf("expected no fields, got %d", len(entry.Fields))
	}
}

func TestWithFields(t *testing.T) {
	fields := []Field{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
	}

	entry := WithFields("TestScope", fields...)

	if entry.Scope != "TestScope" {
		t.Fatalf("expected scope 'TestScope', got %q", entry.Scope)
	}
	if len(entry.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(entry.Fields))
	}
	if entry.Fields[0].Key != "key1" || entry.Fields[0].Value != "value1" {
		t.Fatalf("expected field 1 to be key1=value1, got %s=%s", entry.Fields[0].Key, entry.Fields[0].Value)
	}
	if entry.Fields[1].Key != "key2" || entry.Fields[1].Value != "value2" {
		t.Fatalf("expected field 2 to be key2=value2, got %s=%s", entry.Fields[1].Key, entry.Fields[1].Value)
	}
}

func TestFieldValue(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value any
		want  string
	}{
		{"string value", "key", "value", "value"},
		{"int value", "count", 42, "42"},
		{"uint value", "id", uint32(123), "123"},
		{"bool value", "enabled", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := FieldValue(tt.key, tt.value)
			if field.Key != tt.key {
				t.Fatalf("expected key %q, got %q", tt.key, field.Key)
			}
			if field.Value != tt.want {
				t.Fatalf("expected value %q, got %q", tt.want, field.Value)
			}
		})
	}
}

func TestPacketFields(t *testing.T) {
	fields := PacketFields(1001, "in")

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Key != "packet" || fields[0].Value != "1001" {
		t.Fatalf("expected packet=1001, got %s=%s", fields[0].Key, fields[0].Value)
	}
	if fields[1].Key != "dir" || fields[1].Value != "in" {
		t.Fatalf("expected dir=in, got %s=%s", fields[1].Key, fields[1].Value)
	}
}

func TestCommanderFields(t *testing.T) {
	fields := CommanderFields(12345, 67890)

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Key != "commander" || fields[0].Value != "12345" {
		t.Fatalf("expected commander=12345, got %s=%s", fields[0].Key, fields[0].Value)
	}
	if fields[1].Key != "account" || fields[1].Value != "67890" {
		t.Fatalf("expected account=67890, got %s=%s", fields[1].Key, fields[1].Value)
	}
}

func TestEntryWith(t *testing.T) {
	baseEntry := Scope("Base")
	baseEntry.Fields = []Field{{Key: "base", Value: "value"}}

	extendedEntry := baseEntry.With(Field{Key: "extra", Value: "data"})

	if len(extendedEntry.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(extendedEntry.Fields))
	}
	if extendedEntry.Fields[0].Key != "base" {
		t.Fatalf("expected first field 'base', got %q", extendedEntry.Fields[0].Key)
	}
	if extendedEntry.Fields[1].Key != "extra" {
		t.Fatalf("expected second field 'extra', got %q", extendedEntry.Fields[1].Key)
	}

	if len(baseEntry.Fields) != 1 {
		t.Fatalf("expected base entry to still have 1 field, got %d", len(baseEntry.Fields))
	}
}

func TestLogLevelFiltering(t *testing.T) {
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", originalLogLevel)

	t.Run("error level logs error", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "error")
		logLevel = LOG_LEVEL_ERROR
		if !shouldLog(LOG_LEVEL_ERROR) {
			t.Fatalf("expected error level to log")
		}
		if shouldLog(LOG_LEVEL_WARN) {
			t.Fatalf("expected warn level to not log when log level is error")
		}
	})

	t.Run("info level logs info", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "info")
		logLevel = LOG_LEVEL_INFO
		if !shouldLog(LOG_LEVEL_INFO) {
			t.Fatalf("expected info level to log")
		}
		if shouldLog(LOG_LEVEL_DEBUG) {
			t.Fatalf("expected debug level to not log when log level is info")
		}
	})

	t.Run("debug level logs debug", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "debug")
		logLevel = LOG_LEVEL_DEBUG
		if !shouldLog(LOG_LEVEL_DEBUG) {
			t.Fatalf("expected debug level to log")
		}
	})
}

func TestLevelName(t *testing.T) {
	tests := []struct {
		level int
		want  string
	}{
		{LOG_LEVEL_DEBUG, "DEBUG"},
		{LOG_LEVEL_INFO, "INFO"},
		{LOG_LEVEL_WARN, "WARN"},
		{LOG_LEVEL_ERROR, "ERROR"},
		{999, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := levelName(tt.level); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFormatFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []Field
		want   string
	}{
		{
			name:   "single field",
			fields: []Field{{Key: "key1", Value: "value1"}},
			want:   "key1=value1",
		},
		{
			name:   "multiple fields",
			fields: []Field{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			want:   "key1=value1 key2=value2",
		},
		{
			name:   "empty value",
			fields: []Field{{Key: "key1", Value: ""}},
			want:   "key1=-",
		},
		{
			name:   "empty key",
			fields: []Field{{Key: "", Value: "value1"}, {Key: "key2", Value: "value2"}},
			want:   "key2=value2",
		},
		{
			name:   "no fields",
			fields: []Field{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFields(tt.fields...); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestLogEvent(t *testing.T) {
	os.Setenv("LOG_LEVEL", "info")
	defer os.Setenv("LOG_LEVEL", "")

	t.Run("with subcategory", func(t *testing.T) {
		LogEvent("ORM", "Init", "Database initialized", LOG_LEVEL_INFO)
	})

	t.Run("without subcategory", func(t *testing.T) {
		LogEvent("ORM", "", "Simple event", LOG_LEVEL_INFO)
	})
}

func TestSpan(t *testing.T) {
	os.Setenv("LOG_LEVEL", "info")
	defer os.Setenv("LOG_LEVEL", "")

	t.Run("success", func(t *testing.T) {
		span := Span("TestScope", Field{Key: "key", Value: "value"})
		span(nil)
	})

	t.Run("error", func(t *testing.T) {
		span := Span("TestScope", Field{Key: "key", Value: "value"})
		span(nil)
	})
}

func TestLogOutputFormat(t *testing.T) {
	os.Setenv("LOG_LEVEL", "info")
	defer os.Setenv("LOG_LEVEL", "")

	t.Run("log without fields", func(t *testing.T) {
		line := formatLine("TestScope", "test message", LOG_LEVEL_INFO)

		expectedParts := []string{"INFO", "TestScope", "test message"}
		for _, part := range expectedParts {
			if !strings.Contains(line, part) {
				t.Fatalf("expected line to contain %q, got %q", part, line)
			}
		}
	})

	t.Run("log with fields", func(t *testing.T) {
		fields := []Field{{Key: "key1", Value: "value1"}}
		line := formatLine("TestScope", "test message", LOG_LEVEL_INFO, fields...)

		expectedParts := []string{"INFO", "TestScope", "test message", "key1=value1"}
		for _, part := range expectedParts {
			if !strings.Contains(line, part) {
				t.Fatalf("expected line to contain %q, got %q", part, line)
			}
		}
	})
}

func TestGetShortDateTime(t *testing.T) {
	datetime := getShortDateTime()

	dateTimeRegex := regexp.MustCompile(`^\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}$`)
	if !dateTimeRegex.MatchString(datetime) {
		t.Fatalf("expected datetime format dd/mm/yyyy hh:mm:ss, got %q", datetime)
	}
}

func TestLogLevels(t *testing.T) {
	os.Setenv("LOG_LEVEL", "info")
	logLevel = LOG_LEVEL_INFO

	t.Run("debug not logged", func(t *testing.T) {
		if shouldLog(LOG_LEVEL_DEBUG) {
			t.Fatalf("expected debug to not be logged when level is info")
		}
	})

	t.Run("info is logged", func(t *testing.T) {
		if !shouldLog(LOG_LEVEL_INFO) {
			t.Fatalf("expected info to be logged when level is info")
		}
	})

	t.Run("warn is logged", func(t *testing.T) {
		if !shouldLog(LOG_LEVEL_WARN) {
			t.Fatalf("expected warn to be logged when level is info")
		}
	})

	t.Run("error is logged", func(t *testing.T) {
		if !shouldLog(LOG_LEVEL_ERROR) {
			t.Fatalf("expected error to be logged when level is info")
		}
	})
}

func TestLogLevelCaseInsensitive(t *testing.T) {
	originalLogLevel := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", originalLogLevel)

	t.Run("env reads debug level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "DEBUG")
		logLevel = LOG_LEVEL_AUTO
		if !shouldLog(LOG_LEVEL_DEBUG) {
			t.Fatalf("expected debug to be logged with env DEBUG")
		}
	})

	t.Run("env reads warning level", func(t *testing.T) {
		os.Setenv("LOG_LEVEL", "WARNING")
		logLevel = LOG_LEVEL_AUTO
		if !shouldLog(LOG_LEVEL_WARN) {
			t.Fatalf("expected warn to be logged with env WARNING")
		}
	})
}
