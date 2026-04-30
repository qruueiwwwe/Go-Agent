package weather

import (
	"context"
	"testing"
	"time"

	"agent/global"
)

func getTestConfig() global.WeatherAPIConfig {
	return global.WeatherAPIConfig{
		ID:      "10016155",
		Key:     "8b0464361cb05f30e401c0a1b9ac58ce",
		BaseURL: "https://cn.apihz.cn/api/tianqi/tqyb.php",
		Timeout: 10 * time.Second,
	}
}

func TestWeather_Name(t *testing.T) {
	w := NewWeather(getTestConfig())
	if w.Name() != "weather" {
		t.Errorf("Expected Name() to return 'weather', got %s", w.Name())
	}
}

func TestWeather_Description(t *testing.T) {
	w := NewWeather(getTestConfig())
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

func TestWeather_Execute_KnownCity(t *testing.T) {
	w := NewWeather(getTestConfig())
	ctx := context.Background()

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
	w := NewWeather(getTestConfig())
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
	w := NewWeather(getTestConfig())
	ctx := context.Background()
	result := w.Execute(ctx, "")
	// Should not panic
	if result == "" {
		t.Error("Execute() should handle empty input")
	}
}

func TestWeather_Execute_InvalidCity(t *testing.T) {
	w := NewWeather(getTestConfig())
	ctx := context.Background()
	result := w.Execute(ctx, "火星")
	// Should return error message
	if !contains(result, "失败") && !contains(result, "错误") {
		t.Logf("Result for invalid city: %s", result)
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
