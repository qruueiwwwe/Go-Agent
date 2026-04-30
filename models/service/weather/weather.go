package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"agent/global"
	"agent/library/log"
)

// Weather 天气查询逻辑
type Weather struct {
	config global.WeatherAPIConfig
	client *http.Client
}

func NewWeather(cfg global.WeatherAPIConfig) *Weather {
	return &Weather{
		config: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

func (w *Weather) Name() string {
	return "weather"
}

func (w *Weather) Description() string {
	return "查询天气，输入格式：城市名称（支持：今天、明天、后天、七天），如：北京、上海、或者：北京明天、西安七天"
}

func (w *Weather) Execute(ctx context.Context, input string) string {
	log.Info(ctx, "Weather.Execute: 入参 input=%s", input)

	// 解析城市和天数
	city, days := parseWeatherInput(input)
	if city == "" {
		city = input
		days = 1
	}

	log.Info(ctx, "Weather.Execute: 解析结果 city=%s, days=%d", city, days)

	// 优先调用接口盒子API（中国气象局数据）
	result, err := w.getWeatherFromAPIHZ(ctx, city, days)
	if err == nil {
		log.Info(ctx, "Weather.Execute: 接口盒子查询成功 city=%s", city)
		return result
	}

	log.Error(ctx, "Weather.Execute: 接口盒子查询失败 city=%s, err=%v", city, err)

	// 接口盒子失败，尝试高德天气API
	result, err = w.getWeatherFromAmap(ctx, city, days)
	if err == nil {
		log.Info(ctx, "Weather.Execute: 高德API查询成功 city=%s", city)
		return result
	}

	log.Error(ctx, "Weather.Execute: 高德API查询失败 city=%s, err=%v", city, err)

	// 两个都失败
	return fmt.Sprintf("查询「%s」天气失败：%s", city, err.Error())
}

// parseWeatherInput 解析输入，提取城市和天数
func parseWeatherInput(input string) (string, int) {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, " ", "")

	days := 1
	if strings.Contains(input, "七天") || strings.Contains(input, "7天") || strings.Contains(input, "一周") || strings.Contains(input, "未来七天") {
		days = 7
	} else if strings.Contains(input, "三天") || strings.Contains(input, "3天") || strings.Contains(input, "未来三天") {
		days = 3
	} else if strings.Contains(input, "明天") || strings.Contains(input, "明日") {
		days = 2
	} else if strings.Contains(input, "后天") || strings.Contains(input, "后日") {
		days = 3
	}

	cityPatterns := []string{"今天", "明天", "后天", "三日", "三天", "3天", "七日", "七天", "一周", "7天", "天气", "未来"}
	city := input
	for _, p := range cityPatterns {
		city = strings.ReplaceAll(city, p, "")
	}
	city = strings.TrimSpace(city)

	if city == "" {
		city = input
	}

	return city, days
}

// ==================== 接口盒子 API ====================

// APIHZResponse 接口盒子API响应结构
type APIHZResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`

	// 地区信息
	Guo   string `json:"guo"`   // 国家
	Sheng string `json:"sheng"` // 省份
	Shi   string `json:"shi"`   // 城市
	Name  string `json:"name"`  // 地点

	// 今日天气
	Weather1 string `json:"weather1"` // 白天天气
	Weather2 string `json:"weather2"` // 夜间天气
	WD1      string `json:"wd1"`      // 白天温度
	WD2      string `json:"wd2"`      // 夜间温度

	WindDirection1 string `json:"winddirection1"` // 白天风向
	WindDirection2 string `json:"winddirection2"` // 夜间风向
	WindLevel1     string `json:"windleve1"`      // 白天风力
	WindLevel2     string `json:"windleve2"`      // 夜间风力

	Lon    string `json:"lon"`    // 经度
	Lat    string `json:"lat"`    // 纬度
	Uptime string `json:"uptime"` // 更新时间

	// 实时天气
	NowInfo struct {
		Temperature interface{} `json:"temperature"` // 实时温度
		Humidity    interface{} `json:"humidity"`    // 湿度
		Pressure    interface{} `json:"pressure"`    // 气压
		Feelst      interface{} `json:"feelst"`      // 体感温度
	} `json:"nowinfo"`

	// 多天预报
	WeatherDay2 *APIHZDayData `json:"weatherday2"`
	WeatherDay3 *APIHZDayData `json:"weatherday3"`
	WeatherDay4 *APIHZDayData `json:"weatherday4"`
	WeatherDay5 *APIHZDayData `json:"weatherday5"`
	WeatherDay6 *APIHZDayData `json:"weatherday6"`
	WeatherDay7 *APIHZDayData `json:"weatherday7"`
}

// APIHZDayData 多天天气数据
type APIHZDayData struct {
	Date     string      `json:"date"`
	Weather1 string      `json:"weather1"`
	WD1      interface{} `json:"wd1"`
	WD2      interface{} `json:"wd2"`
}

// getWeatherFromAPIHZ 调用接口盒子API
func (w *Weather) getWeatherFromAPIHZ(ctx context.Context, city string, days int) (string, error) {
	params := url.Values{}
	params.Set("id", w.config.ID)
	params.Set("key", w.config.Key)
	params.Set("place", city)
	params.Set("day", fmt.Sprintf("%d", days))

	requestURL := w.config.BaseURL + "?" + params.Encode()

	resp, err := w.client.Get(requestURL)
	if err != nil {
		log.Error(ctx, "Weather.getWeatherFromAPIHZ: 网络请求失败 city=%s, err=%v", city, err)
		return "", fmt.Errorf("网络请求失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(ctx, "Weather.getWeatherFromAPIHZ: 读取响应失败 city=%s, err=%v", city, err)
		return "", fmt.Errorf("读取响应失败")
	}

	var data APIHZResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Error(ctx, "Weather.getWeatherFromAPIHZ: 解析响应失败 city=%s, body=%s", city, string(body)[:minInt(100, len(body))])
		return "", fmt.Errorf("解析响应失败: %s", string(body)[:minInt(100, len(body))])
	}

	if data.Code != 200 {
		log.Error(ctx, "Weather.getWeatherFromAPIHZ: API返回错误 city=%s, code=%d, msg=%s", city, data.Code, data.Msg)
		return "", fmt.Errorf("%s", data.Msg)
	}

	return w.formatAPIHZResult(&data, days), nil
}

// formatAPIHZResult 格式化接口盒子结果
func (w *Weather) formatAPIHZResult(data *APIHZResponse, days int) string {
	var result strings.Builder

	location := data.Name
	if location == "" {
		location = data.Shi
	}
	result.WriteString(fmt.Sprintf("%s未来%d天天气：\n", location, days))
	result.WriteString(fmt.Sprintf("更新时间：%s\n", data.Uptime))

	// 今日天气
	result.WriteString(fmt.Sprintf("\n今天：%s，%s~%s°C", data.Weather1, data.WD1, data.WD2))
	if data.WindDirection1 != "" && data.WindLevel1 != "" {
		result.WriteString(fmt.Sprintf("，%s%s", data.WindDirection1, data.WindLevel1))
	}
	result.WriteString("\n")

	// 实时天气
	if data.NowInfo.Temperature != nil {
		temp := toFloat64(data.NowInfo.Temperature)
		humidity := toFloat64(data.NowInfo.Humidity)
		if temp > 0 {
			result.WriteString(fmt.Sprintf("实时：%.0f°C，湿度%.0f%%", temp, humidity))
			if feelst := toFloat64(data.NowInfo.Feelst); feelst > 0 {
				result.WriteString(fmt.Sprintf("，体感%.0f°C", feelst))
			}
			result.WriteString("\n")
		}
	}

	// 多天预报
	dayDataList := []*APIHZDayData{
		data.WeatherDay2, data.WeatherDay3, data.WeatherDay4,
		data.WeatherDay5, data.WeatherDay6, data.WeatherDay7,
	}
	dayNames := []string{"明天", "后天", "第4天", "第5天", "第6天", "第7天"}

	for i, d := range dayDataList {
		if i >= days-1 || d == nil {
			break
		}
		if d.Weather1 != "" {
			result.WriteString(fmt.Sprintf("%s：%s，%.0f~%.0f°C\n", dayNames[i], d.Weather1, toFloat64(d.WD1), toFloat64(d.WD2)))
		}
	}

	return strings.TrimSpace(result.String())
}

// ==================== 高德天气 API ====================

// AmapWeatherResponse 高德天气API响应
type AmapWeatherResponse struct {
	Status   string `json:"status"`
	Count    string `json:"count"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Lives    []struct {
		Province      string `json:"province"`
		City          string `json:"city"`
		Adcode        string `json:"adcode"`
		Weather       string `json:"weather"`
		Temperature   string `json:"temperature"`
		WindDirection string `json:"winddirection"`
		WindPower     string `json:"windpower"`
		Humidity      string `json:"humidity"`
		ReportTime    string `json:"reporttime"`
	} `json:"lives"`
	Forecasts []struct {
		City       string `json:"city"`
		Adcode     string `json:"adcode"`
		Province   string `json:"province"`
		ReportTime string `json:"reporttime"`
		Casts      []struct {
			Date         string `json:"date"`
			Week         string `json:"week"`
			DayWeather   string `json:"dayweather"`
			NightWeather string `json:"nightweather"`
			DayTemp      string `json:"daytemp"`
			NightTemp    string `json:"nighttemp"`
			DayWind      string `json:"daywind"`
			NightWind    string `json:"nightwind"`
			DayPower     string `json:"daypower"`
			NightPower   string `json:"nightpower"`
		} `json:"casts"`
	} `json:"forecasts"`
}

// getWeatherFromAmap 调用高德天气API
func (w *Weather) getWeatherFromAmap(ctx context.Context, city string, days int) (string, error) {
	// 高德API需要adcode，尝试用城市名查找
	adcode := getAdcodeByCity(city)
	if adcode == "" {
		log.Error(ctx, "Weather.getWeatherFromAmap: 未找到城市编码 city=%s", city)
		return "", fmt.Errorf("未找到城市编码")
	}

	log.Info(ctx, "Weather.getWeatherFromAmap: 城市编码 city=%s, adcode=%s", city, adcode)

	// 构建请求
	params := url.Values{}
	params.Set("key", "ed105e515b93def32b5db5b0b420e3e4")
	params.Set("city", adcode)
	if days > 1 {
		params.Set("extensions", "all")
	}

	requestURL := "https://restapi.amap.com/v3/weather/weatherInfo?" + params.Encode()

	resp, err := w.client.Get(requestURL)
	if err != nil {
		log.Error(ctx, "Weather.getWeatherFromAmap: 网络请求失败 city=%s, err=%v", city, err)
		return "", fmt.Errorf("高德API网络请求失败")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(ctx, "Weather.getWeatherFromAmap: 读取响应失败 city=%s, err=%v", city, err)
		return "", fmt.Errorf("读取响应失败")
	}

	var data AmapWeatherResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Error(ctx, "Weather.getWeatherFromAmap: 解析响应失败 city=%s, err=%v", city, err)
		return "", fmt.Errorf("解析响应失败")
	}

	if data.Status != "1" {
		log.Error(ctx, "Weather.getWeatherFromAmap: API返回错误 city=%s, status=%s, info=%s", city, data.Status, data.Info)
		return "", fmt.Errorf("高德API返回错误: %s", data.Info)
	}

	return w.formatAmapResult(&data, city, days), nil
}

// formatAmapResult 格式化高德天气结果
func (w *Weather) formatAmapResult(data *AmapWeatherResponse, city string, days int) string {
	var result strings.Builder

	// 实况天气
	if len(data.Lives) > 0 {
		live := data.Lives[0]
		result.WriteString(fmt.Sprintf("%s今日天气：\n", live.City))
		result.WriteString(fmt.Sprintf("天气：%s，温度：%s°C\n", live.Weather, live.Temperature))
		result.WriteString(fmt.Sprintf("风向：%s，风力：%s级\n", live.WindDirection, live.WindPower))
		result.WriteString(fmt.Sprintf("湿度：%s%%\n", live.Humidity))
		result.WriteString(fmt.Sprintf("更新时间：%s", live.ReportTime))
		return result.String()
	}

	// 预报天气
	if len(data.Forecasts) > 0 && len(data.Forecasts[0].Casts) > 0 {
		forecast := data.Forecasts[0]
		result.WriteString(fmt.Sprintf("%s未来%d天天气：\n", forecast.City, minInt(days, len(forecast.Casts))))
		result.WriteString(fmt.Sprintf("更新时间：%s\n", forecast.ReportTime))

		dayNames := []string{"今天", "明天", "后天", "第4天", "第5天", "第6天"}
		for i, cast := range forecast.Casts {
			if i >= days {
				break
			}
			dayName := cast.Date
			if i < len(dayNames) {
				dayName = dayNames[i]
			}
			result.WriteString(fmt.Sprintf("\n%s：%s，%s~%s°C", dayName, cast.DayWeather, cast.DayTemp, cast.NightTemp))
		}
		return result.String()
	}

	return fmt.Sprintf("%s暂无天气数据", city)
}

// ==================== 辅助函数 ====================

// toFloat64 将interface{}转换为float64
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
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getAdcodeByCity 城市名转adcode（常用城市）
func getAdcodeByCity(city string) string {
	// 省会及热门城市adcode映射
	adcodeMap := map[string]string{
		"北京": "110000", "上海": "310000", "广州": "440100", "深圳": "440300",
		"西安": "610100", "成都": "510100", "杭州": "330100", "武汉": "420100",
		"南京": "320100", "重庆": "500100", "天津": "120000", "苏州": "320500",
		"郑州": "410100", "长沙": "430100", "青岛": "370200", "沈阳": "210100",
		"大连": "210200", "厦门": "350200", "昆明": "530100", "哈尔滨": "230100",
		"长春": "220100", "福州": "350100", "南昌": "360100", "贵阳": "520100",
		"太原": "140100", "石家庄": "130100", "济南": "370100", "兰州": "620100",
		"乌鲁木齐": "650100", "呼和浩特": "150100", "南宁": "450100", "海口": "460100",
		"银川": "640100", "西宁": "630100", "拉萨": "540100",
		// 简称
		"京": "110000", "沪": "310000", "穗": "440100", "鹏": "440300",
	}

	if code, ok := adcodeMap[city]; ok {
		return code
	}

	// 尝试模糊匹配
	for name, code := range adcodeMap {
		if strings.Contains(name, city) || strings.Contains(city, name) {
			return code
		}
	}

	return ""
}
