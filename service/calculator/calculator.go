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

	for _, op := range ops {
		if strings.Contains(input, op) {
			parts := strings.Split(input, op)
			if len(parts) != 2 {
				return "计算错误：输入格式不正确"
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

			// 如果是整数，直接返回整数形式
			if res == float64(int64(res)) {
				return fmt.Sprintf("%.0f", res)
			}
			return fmt.Sprintf("%f", res)
		}
	}
	return "计算错误：无法识别的表达式"
}