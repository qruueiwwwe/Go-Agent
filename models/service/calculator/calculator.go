package calculator

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Calculator 计算器逻辑
type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Name() string {
	return "calculator"
}

func (c *Calculator) Description() string {
	return "用于数学计算，输入格式：1+2 或 3*4"
}

func (c *Calculator) Execute(ctx context.Context, input string) string {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	input = strings.ReplaceAll(input, " ", "")
	ops := []string{"+", "-", "*", "/"}
	hasDecimal := strings.Contains(input, ".")

	for _, op := range ops {
		// Find the operator position, skipping the first character if it's a negative sign
		var opIndex int
		found := false
		startPos := 0
		if len(input) > 0 && input[0] == '-' {
			startPos = 1
		}
		for i := startPos; i < len(input); i++ {
			if string(input[i]) == op {
				opIndex = i
				found = true
				break
			}
		}

		if found {
			parts := []string{input[:opIndex], input[opIndex+1:]}
			if len(parts[0]) == 0 || len(parts[1]) == 0 {
				return "计算错误：无法识别的表达式"
			}

			a, err := strconv.ParseFloat(parts[0], 64)
			if err != nil {
				return "计算错误：无法解析数字 " + parts[0]
			}

			b, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				return "计算错误：无法解析数字 " + parts[1]
			}

			var res float64
			switch op {
			case "+":
				res = a + b
			case "-":
				res = a - b
			case "*":
				res = a * b
			case "/":
				if b == 0 {
					return "计算错误：除数不能为0"
				}
				res = a / b
			}

			// 如果输入包含小数点，使用浮点数格式
			if hasDecimal {
				return fmt.Sprintf("%f", res)
			}
			// 如果是整数，直接返回整数形式
			if res == float64(int64(res)) {
				return fmt.Sprintf("%.0f", res)
			}
			return fmt.Sprintf("%f", res)
		}
	}
	return "计算错误：无法识别的表达式"
}
