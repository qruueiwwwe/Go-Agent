package main

import (
	"fmt"
)

// Calculator 简单计算器
type Calculator struct {
	result int
}

// NewCalculator 创建计算器实例
func NewCalculator() *Calculator {
	return &Calculator{result: 0}
}

// Add 加法操作
func (c *Calculator) Add(a, b int) int {
	c.result = a + b
	return c.result
}

// Multiply 乘法操作
func (c *Calculator) Multiply(a, b int) int {
	c.result = a * b
	return c.result
}

// GetResult 获取结果
func (c *Calculator) GetResult() int {
	return c.result
}

func main() {
	calc := NewCalculator()
	fmt.Println("2 + 3 =", calc.Add(2, 3))
	fmt.Println("5 * 4 =", calc.Multiply(5, 4))
	fmt.Println("Current result:", calc.GetResult())
}
