package weather

import (
	"context"
	"testing"
)

func TestWeather_Name(t *testing.T) {
	w := NewWeather()
	if w.Name() != "weather" {
		t.Errorf("Expected Name() to return 'weather', got %s", w.Name())
	}
}

func TestWeather_Description(t *testing.T) {
	w := NewWeather()
	desc := w.Description()
	if desc == "" {
		t.Error("Description() should not return empty string")
	}
	if !contains(desc, "天气") {
		t.Errorf("Description should contain '天气', got %s", desc)
	}
}

func TestParseWeatherInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected struct {
			city string
			days int
		}
	}{
		{"Simple city", "北京", struct {
			city string
			days int
		}{city: "北京", days: 1}},
		{"Tomorrow", "北京明天", struct {
			city string
			days int
		}{city: "北京", days: 2}},
		{"Day after tomorrow", "上海后天", struct {
			city string
			days int
		}{city: "上海", days: 3}},
		{"Three days", "广州三天", struct {
			city string
			days int
		}{city: "广州", days: 3}},
		{"Seven days", "深圳七天", struct {
			city string
			days int
		}{city: "深圳", days: 7}},
		{"Seven days with number", "杭州7天", struct {
			city string
			days int
		}{city: "杭州", days: 7}},
		{"One week", "南京一周", struct {
			city string
			days int
		}{city: "南京", days: 7}},
		{"Future seven days", "苏州未来七天", struct {
			city string
			days int
		}{city: "苏州", days: 7}},
		{"With spaces", " 北京 明天 ", struct {
			city string
			days int
		}{city: "北京", days: 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			city, days := parseWeatherInput(tt.input)
			if city != tt.expected.city {
				t.Errorf("parseWeatherInput(%s) city = %s, expected %s", tt.input, city, tt.expected.city)
			}
			if days != tt.expected.days {
				t.Errorf("parseWeatherInput(%s) days = %d, expected %d", tt.input, days, tt.expected.days)
			}
		})
	}
}

func TestGetWeatherDescByCode(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{0, "晴天"},
		{1, "晴朗"},
		{2, "多云"},
		{3, "阴天"},
		{45, "雾"},
		{48, "雾"},
		{51, "小雨"},
		{61, "小雨"},
		{71, "小雪"},
		{95, "雷暴"},
		{99, "雷暴"},
		{9999, "天气9999"}, // Unknown code
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := getWeatherDescByCode(tt.code)
			if result != tt.expected {
				t.Errorf("getWeatherDescByCode(%d) = %s, expected %s", tt.code, result, tt.expected)
			}
		})
	}
}

func TestTranslateWeather(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Sunny", "晴天"},
		{"Clear", "晴天"},
		{"Partly cloudy", "多云"},
		{"Cloudy", "阴天"},
		{"Overcast", "阴天"},
		{"Light rain", "小雨"},
		{"Moderate rain", "中雨"},
		{"Heavy rain", "大雨"},
		{"Snow", "雪天"},
		{"Fog", "雾"},
		{"Thunderstorm", "雷暴"},
		{"Unknown condition", "Unknown condition"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := translateWeather(tt.input)
			if result != tt.expected {
				t.Errorf("translateWeather(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetStringValue(t *testing.T) {
	m := map[string]interface{}{
		"stringKey":  "stringValue",
		"intKey":     42.0,
		"missingKey": nil,
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"String value", "stringKey", "stringValue"},
		{"Int value", "intKey", "42"},
		{"Missing key", "missingKey", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringValue(m, tt.key)
			if result != tt.expected {
				t.Errorf("getStringValue(%s) = %s, expected %s", tt.key, result, tt.expected)
			}
		})
	}
}

func TestWeather_Execute_UnknownCity(t *testing.T) {
	w := NewWeather()
	ctx := context.Background()
	// This will try to fetch from wttr.in which may fail or return a different format
	// So we just check it doesn't panic
	result := w.Execute(ctx, "NonexistentCity")
	if result == "" {
		t.Error("Execute() should return some result even for unknown city")
	}
}

func TestWeather_Execute_KnownCity(t *testing.T) {
	w := NewWeather()
	ctx := context.Background()

	// Test with a known city that has coordinates
	result := w.Execute(ctx, "北京")
	if result == "" {
		t.Error("Execute() should return a result for Beijing")
	}
	// Result should contain the city name
	if !contains(result, "北京") {
		t.Errorf("Result should contain '北京', got: %s", result)
	}
}

func TestWeather_Execute_WithDays(t *testing.T) {
	w := NewWeather()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"Today", "北京"},
		{"Tomorrow", "北京明天"},
		{"Seven days", "北京七天"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := w.Execute(ctx, tt.input)
			if result == "" {
				t.Errorf("Execute(%s) should return a result", tt.input)
			}
			// Result should contain the city name
			if !contains(result, "北京") {
				t.Errorf("Result should contain '北京', got: %s", result)
			}
		})
	}
}

func TestWeather_Execute_EmptyInput(t *testing.T) {
	w := NewWeather()
	ctx := context.Background()
	result := w.Execute(ctx, "")
	// Should not panic
	if result == "" {
		t.Error("Execute() should handle empty input")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
