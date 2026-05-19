package calculator

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"

	"agent/library/log"
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
	return "用于数学计算，支持复杂表达式：1+2*3, 2^16-1, 2**16-1"
}

func (c *Calculator) Execute(ctx context.Context, input string) string {
	log.Info(ctx, "Calculator.Execute: 入参 input=%s", input)

	defer func() {
		if r := recover(); r != nil {
			log.Error(ctx, "Calculator.Execute: panic恢复: %v", r)
		}
	}()

	// 移除空格
	input = strings.ReplaceAll(input, " ", "")

	// 将 Python 风格的幂运算符 ** 替换为 ^
	input = strings.ReplaceAll(input, "**", "^")

	// 添加 pow 函数
	functions := map[string]govaluate.ExpressionFunction{
		"pow": func(args ...interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("pow 需要2个参数")
			}
			// 转换参数为 float64
			a := toFloat64(args[0])
			b := toFloat64(args[1])
			result := math.Pow(a, b)
			return result, nil
		},
		"abs": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("abs 需要1个参数")
			}
			a := toFloat64(args[0])
			if a < 0 {
				return -a, nil
			}
			return a, nil
		},
		"sqrt": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sqrt 需要1个参数")
			}
			a := toFloat64(args[0])
			if a < 0 {
				return nil, fmt.Errorf("sqrt 参数不能为负数")
			}
			return math.Sqrt(a), nil
		},
	}

	// 预处理表达式，将 ^ 替换为 pow 函数
	preprocessed := preprocessPowerOp(input)

	// 创建带函数的 govaluate 表达式
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(preprocessed, functions)
	if err != nil {
		log.Error(ctx, "Calculator.Execute: 解析表达式失败 input=%s, preprocessed=%s, err=%v", input, preprocessed, err)
		// 如果解析失败，尝试用原始表达式（可能只是简单计算）
		expr, err = govaluate.NewEvaluableExpressionWithFunctions(input, functions)
		if err != nil {
			return fmt.Sprintf("计算错误：无法解析表达式 %s", input)
		}
	}

	// 计算结果
	result, err := expr.Evaluate(nil)
	if err != nil {
		log.Error(ctx, "Calculator.Execute: 计算失败 input=%s, err=%v", input, err)
		return fmt.Sprintf("计算错误：计算失败 %s", input)
	}

	// 格式化结果
	var resultStr string
	switch v := result.(type) {
	case float64:
		// 如果是整数，不显示小数
		if v == float64(int64(v)) {
			resultStr = fmt.Sprintf("%.0f", v)
		} else {
			// 移除尾随的零
			resultStr = fmt.Sprintf("%f", v)
			resultStr = strings.TrimRight(resultStr, "0")
			resultStr = strings.TrimRight(resultStr, ".")
			if resultStr == "" {
				resultStr = "0"
			}
		}
	case int:
		resultStr = fmt.Sprintf("%d", v)
	case int64:
		resultStr = fmt.Sprintf("%d", v)
	default:
		resultStr = fmt.Sprintf("%v", result)
	}

	log.Info(ctx, "Calculator.Execute: 计算成功 %s=%s", input, resultStr)
	return resultStr
}

// toFloat64 将任意类型转换为 float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f
		}
	}
	return 0
}

// preprocessPowerOp 预处理表达式，将 ^ 替换为 pow 函数
// 这是一个简化版本，处理常见情况
func preprocessPowerOp(input string) string {
	// 查找所有的 ^ 运算符
	// 这是一个简化的处理，从右到左处理，从最低优先级开始

	result := input
	changed := true

	for changed {
		changed = false
		// 查找最右边的 ^ 运算符（最低优先级）
		// 正则：找到 (数字)^ (数字)，其中两边可以是括号
		re := regexp.MustCompile(`([0-9.]+)\^([0-9.]+)`)
		matches := re.FindAllStringSubmatchIndex(result, -1)

		for _, match := range matches {
			// match: [full_match, start_full, end_full, start_group1, end_group1, start_group2, end_group2]
			if len(match) >= 6 {
				fullStart := match[0]
				fullEnd := match[1]
				group1Start := match[2]
				group1End := match[3]
				group2Start := match[4]
				group2End := match[5]

				base := result[group1Start:group1End]
				exponent := result[group2Start:group2End]

				// 替换为 pow(base,exponent)
				newExpr := fmt.Sprintf("pow(%s,%s)", base, exponent)
				result = result[:fullStart] + newExpr + result[fullEnd:]
				changed = true
				break // 重新开始
			}
		}
	}

	return result
}
