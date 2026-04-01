package calculator

import (
	"context"
	"testing"
)

func TestCalculator_Name(t *testing.T) {
	c := NewCalculator()
	if c.Name() != "calculator" {
		t.Errorf("Expected Name() to return 'calculator', got %s", c.Name())
	}
}

func TestCalculator_Description(t *testing.T) {
	c := NewCalculator()
	desc := c.Description()
	if desc == "" {
		t.Error("Description() should not return empty string")
	}
}

func TestCalculator_Execute_Addition(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple addition", "1+2", "3"},
		{"Large numbers", "100+200", "300"},
		{"Negative numbers", "-5+3", "-2"},
		{"Decimals", "1.5+2.5", "4.000000"},
		{"With spaces", "1 + 2", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(ctx, tt.input)
			if result != tt.expected {
				t.Errorf("Execute(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Execute_Subtraction(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple subtraction", "5-3", "2"},
		{"Negative result", "3-5", "-2"},
		{"Large numbers", "1000-500", "500"},
		{"Decimals", "5.5-2.5", "3.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(ctx, tt.input)
			if result != tt.expected {
				t.Errorf("Execute(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Execute_Multiplication(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple multiplication", "3*4", "12"},
		{"Zero", "5*0", "0"},
		{"Negative", "-3*4", "-12"},
		{"Decimals", "2.5*4", "10.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(ctx, tt.input)
			if result != tt.expected {
				t.Errorf("Execute(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Execute_Division(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple division", "6/3", "2"},
		{"Decimals", "5/2", "2.500000"},
		{"Division by integer", "10/5", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(ctx, tt.input)
			if result != tt.expected {
				t.Errorf("Execute(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Execute_DivisionByZero(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()
	result := c.Execute(ctx, "5/0")
	expected := "计算错误：除数不能为0"
	if result != expected {
		t.Errorf("Execute(5/0) = %s, expected %s", result, expected)
	}
}

func TestCalculator_Execute_InvalidInput(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No operator", "123", "计算错误：无法识别的表达式"},
		{"Invalid operator", "1%2", "计算错误：无法识别的表达式"},
		{"Empty input", "", "计算错误：无法识别的表达式"},
		{"Missing operand", "+3", "计算错误：无法识别的表达式"},
		{"Invalid number", "a+b", "计算错误：无法解析数字 a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Execute(ctx, tt.input)
			if result != tt.expected {
				t.Errorf("Execute(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculator_Execute_PanicRecovery(t *testing.T) {
	c := NewCalculator()
	ctx := context.Background()
	// This should not panic
	result := c.Execute(ctx, "1+2")
	if result != "3" {
		t.Errorf("Execute(1+2) = %s, expected 3", result)
	}
}
