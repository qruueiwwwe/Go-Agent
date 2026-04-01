package utils

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"sync"
	"time"

	"agent/library/log"
)

// ========== 校验工具 ==========

// Validator 校验器
type Validator struct{}

// NewValidator 创建校验器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateStruct 校验结构体
func (v *Validator) ValidateStruct(ctx context.Context, s interface{}) error {
	if s == nil {
		return errors.New("参数不能为空")
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("参数必须是结构体")
	}

	// 简单校验：如果有 string 字段不能为空
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		fieldVal := val.Field(i)

		// 检查必须字段 (tag: required)
		required := field.Tag.Get("required")
		if required == "true" {
			if fieldVal.Kind() == reflect.String && fieldVal.String() == "" {
				return errors.New(field.Name + " 不能为空")
			}
		}
	}

	return nil
}

// IsEmail 校验邮箱
func (v *Validator) IsEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsPhone 校验手机号（中国大陆）
func (v *Validator) IsPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsURL 校验URL
func (v *Validator) IsURL(url string) bool {
	pattern := `^https?://[\w\-.]+(:\d+)?(/[\w\-./?%&=]*)?$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// ========== 协程池 ==========

// Task 任务函数
type Task func()

// Pool 协程池
type Pool struct {
	workers   int
	tasks     chan Task
	wg        sync.WaitGroup
	semaphore chan struct{}
}

// NewPool 创建协程池
func NewPool(workers int, queueSize int) *Pool {
	if workers <= 0 {
		workers = 10
	}
	if queueSize <= 0 {
		queueSize = 1000
	}

	p := &Pool{
		workers:   workers,
		tasks:     make(chan Task, queueSize),
		semaphore: make(chan struct{}, workers),
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		go p.worker()
	}

	return p
}

// worker 工作协程
func (p *Pool) worker() {
	for task := range p.tasks {
		p.semaphore <- struct{}{} // 获取信号量
		task()
		<-p.semaphore // 释放信号量
	}
}

// Submit 提交任务
func (p *Pool) Submit(task Task) bool {
	select {
	case p.tasks <- task:
		return true
	default:
		return false // 队列满
	}
}

// Close 关闭协程池
func (p *Pool) Close() {
	close(p.tasks)
}

// Wait 等待所有任务完成
func (p *Pool) Wait() {
	p.wg.Wait()
}

// ========== 时间工具 ==========

// FormatTime 格式化时间
func FormatTime(t time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return t.Format(format)
}

// ParseTime 解析时间
func ParseTime(str string, format string) (time.Time, error) {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Parse(format, str)
}

// DurationString 解析Duration字符串
func DurationString(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// ========== 通用工具 ==========

// SafeGo 安全启动协程
func SafeGo(ctx context.Context, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error(ctx, "goroutine panic: %v", r)
			}
		}()
		fn()
	}()
}

// ChunkSlice 分片
func ChunkSlice(slice []interface{}, chunkSize int) [][]interface{} {
	var chunks [][]interface{}
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// UniqueStringSlice 去重字符串切片
func UniqueStringSlice(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// ContainsString 检查切片是否包含字符串
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
