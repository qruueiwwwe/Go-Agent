package nbnhhsh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"agent/library/log"
	"agent/models/dao"
)

// CanYouSay 缩写词猜测工具
type CanYouSay struct {
	client     *http.Client
	nbnhhshDAO *dao.NbnhhshDAO
}

// NewCanYouSay 创建工具实例
func NewCanYouSay(nbnhhshDAO *dao.NbnhhshDAO) *CanYouSay {
	return &CanYouSay{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		nbnhhshDAO: nbnhhshDAO,
	}
}

// Name 返回工具名称
func (t *CanYouSay) Name() string {
	return "nbnhhsh"
}

// Description 返回工具描述
func (t *CanYouSay) Description() string {
	return "猜测缩写词含义，输入格式：缩写词（如 ngg, yyds, 233 等），多个词用逗号分隔"
}

// Execute 执行工具
func (t *CanYouSay) Execute(ctx context.Context, input string) string {
	log.Info(ctx, "CanYouSay.Execute: 入参 input=%s", input)

	// 解析输入，提取缩写词
	words := t.parseInput(input)
	if len(words) == 0 {
		return "未识别到有效的缩写词"
	}

	log.Info(ctx, "CanYouSay.Execute: 解析结果 words=%v", words)

	// 收集结果
	var results []NbnhhshResult

	for _, word := range words {
		result := t.queryWord(ctx, word)
		results = append(results, result)
	}

	// 格式化输出
	return t.formatResults(results)
}

// parseInput 解析输入，提取缩写词
func (t *CanYouSay) parseInput(input string) []string {
	input = strings.TrimSpace(input)

	// 尝试解析JSON格式
	var jsonInput struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(input), &jsonInput); err == nil && jsonInput.Text != "" {
		input = jsonInput.Text
	}

	// 提取字母、数字组合（可以包含特殊符号作为分隔）
	// 匹配连续的字母或数字
	re := regexp.MustCompile(`[a-zA-Z0-9]+`)
	matches := re.FindAllString(input, -1)

	// 去重
	seen := make(map[string]bool)
	var words []string
	for _, m := range matches {
		m = strings.ToLower(m)
		if !seen[m] && len(m) > 0 {
			seen[m] = true
			words = append(words, m)
		}
	}

	return words
}

// NbnhhshResult API响应结构
type NbnhhshResult struct {
	Name  string   `json:"name"`
	Trans []string `json:"trans"`
}

// queryWord 查询单个词
func (t *CanYouSay) queryWord(ctx context.Context, word string) NbnhhshResult {
	// 1. 检查数据库是否可用，如果不可用直接调用API
	if t.nbnhhshDAO == nil {
		log.Info(ctx, "CanYouSay.queryWord: 数据库未连接，直接调用API word=%s", word)
		results, err := t.callAPI(ctx, word)
		if err != nil {
			log.Error(ctx, "CanYouSay.queryWord: API调用失败 word=%s, err=%v", word, err)
			return NbnhhshResult{Name: word, Trans: []string{"查询失败"}}
		}
		// 找到匹配的结果
		for _, r := range results {
			if strings.EqualFold(r.Name, word) {
				return NbnhhshResult{Name: word, Trans: r.Trans}
			}
		}
		return NbnhhshResult{Name: word, Trans: []string{"未找到解释"}}
	}

	// 2. 先查数据库缓存
	record, err := t.nbnhhshDAO.GetByName(ctx, word)
	if err != nil {
		log.Error(ctx, "CanYouSay.queryWord: 数据库查询失败 word=%s, err=%v", word, err)
	}

	// 3. 如果缓存有效，直接返回
	if t.nbnhhshDAO.IsCacheValid(record) {
		log.Info(ctx, "CanYouSay.queryWord: 使用缓存 word=%s", word)
		return NbnhhshResult{Name: word, Trans: record.Trans}
	}

	// 4. 调用API
	results, err := t.callAPI(ctx, word)
	if err != nil {
		log.Error(ctx, "CanYouSay.queryWord: API调用失败 word=%s, err=%v", word, err)
		// 如果有旧缓存，使用旧缓存
		if record != nil {
			log.Info(ctx, "CanYouSay.queryWord: API失败，使用旧缓存 word=%s", word)
			return NbnhhshResult{Name: word, Trans: record.Trans}
		}
		return NbnhhshResult{Name: word, Trans: []string{"查询失败"}}
	}

	// 5. 找到匹配的结果
	for _, r := range results {
		if strings.EqualFold(r.Name, word) {
			// 6. 保存到数据库（失败不影响返回结果）
			if err := t.nbnhhshDAO.Save(ctx, word, r.Trans); err != nil {
				log.Error(ctx, "CanYouSay.queryWord: 保存缓存失败 word=%s, err=%v", word, err)
			}
			return NbnhhshResult{Name: word, Trans: r.Trans}
		}
	}

	// 7. 未找到匹配
	return NbnhhshResult{Name: word, Trans: []string{"未找到解释"}}
}

// callAPI 调用能不能好好说话API
func (t *CanYouSay) callAPI(ctx context.Context, text string) ([]NbnhhshResult, error) {
	url := "https://lab.magiconch.com/api/nbnhhsh/guess"

	// 构建请求体
	reqBody := map[string]string{"text": text}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var results []NbnhhshResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	log.Info(ctx, "CanYouSay.callAPI: 成功 text=%s, results=%d", text, len(results))
	return results, nil
}

// formatResults 格式化输出结果
func (t *CanYouSay) formatResults(results []NbnhhshResult) string {
	var sb strings.Builder

	for i, r := range results {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("【%s】可能的含义：\n", r.Name))
		if len(r.Trans) == 0 {
			sb.WriteString("  暂无解释")
		} else {
			for j, trans := range r.Trans {
				if j >= 10 {
					sb.WriteString(fmt.Sprintf("  ... 等共 %d 个解释", len(r.Trans)))
					break
				}
				sb.WriteString(fmt.Sprintf("  %d. %s\n", j+1, trans))
			}
		}
	}

	return strings.TrimSpace(sb.String())
}
