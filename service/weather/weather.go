package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// 城市坐标映射表
var cityCoords = map[string][2]float64{
	"北京":   {39.9042, 116.4074},
	"上海":   {31.2304, 121.4737},
	"广州":   {23.1291, 113.2644},
	"深圳":   {22.5431, 114.0579},
	"西安":   {34.3416, 108.9398},
	"成都":   {30.5728, 104.0668},
	"杭州":   {30.2741, 120.1551},
	"武汉":   {30.5928, 114.3055},
	"南京":   {32.0603, 118.7969},
	"重庆":   {29.4316, 106.9123},
	"天津":   {39.3434, 117.3616},
	"苏州":   {31.2989, 120.5853},
	"郑州":   {34.7466, 113.6253},
	"长沙":   {28.2282, 112.9388},
	"青岛":   {36.0671, 120.3826},
	"沈阳":   {41.7968, 123.4315},
	"大连":   {38.9140, 121.6147},
	"厦门":   {24.4798, 118.0894},
	"昆明":   {25.0406, 102.7129},
	"哈尔滨":  {45.8038, 126.5340},
	"长春":   {43.8171, 125.3235},
	"福州":   {26.0745, 119.2965},
	"南昌":   {28.6829, 115.8579},
	"贵阳":   {26.6470, 106.6302},
	"太原":   {37.8706, 112.5489},
	"石家庄":  {38.0428, 114.5149},
	"济南":   {36.6512, 117.1205},
	"兰州":   {36.0611, 103.8343},
	"乌鲁木齐": {43.8256, 87.6168},
	"呼和浩特": {40.8424, 111.7492},
	"南宁":   {22.8170, 108.3665},
	"海口":   {20.0444, 110.1999},
	"银川":   {38.4872, 106.2309},
	"西宁":   {36.6171, 101.7782},
	"拉萨":   {29.6500, 91.1000},
	"东莞":   {23.0209, 113.7518},
	"佛山":   {23.0218, 113.1219},
	"无锡":   {31.4906, 120.3119},
	"宁波":   {29.8683, 121.5440},
	"温州":   {28.0006, 120.6994},
	"嘉兴":   {30.7522, 120.7550},
	"绍兴":   {30.0302, 120.5801},
	"金华":   {29.0787, 119.6479},
	"台州":   {28.6560, 121.4209},
	"湖州":   {30.8929, 120.0930},
	"徐州":   {34.2044, 117.2859},
	"扬州":   {32.3912, 119.4250},
	"镇江":   {32.2044, 119.4551},
	"泰州":   {32.4559, 119.9232},
	"南通":   {31.9802, 120.8942},
	"盐城":   {33.3497, 120.1630},
	"连云港":  {34.5967, 119.2216},
	"淮安":   {33.5517, 119.0153},
	"宿迁":   {33.9631, 118.2752},
	"芜湖":   {31.3350, 118.4330},
	"蚌埠":   {32.9167, 117.3889},
	"淮南":   {32.6264, 116.9997},
	"马鞍山":  {31.6703, 118.5076},
	"淮北":   {33.9560, 116.7980},
	"铜陵":   {30.9294, 117.8129},
	"安庆":   {30.5431, 117.0634},
	"黄山":   {29.7148, 118.3380},
	"滁州":   {32.3017, 118.3275},
	"阜阳":   {32.8897, 115.8143},
	"宿州":   {33.6461, 116.9641},
	"六安":   {31.7348, 116.5080},
	"亳州":   {33.8446, 115.7784},
	"池州":   {30.6644, 117.4917},
	"宣城":   {30.9404, 118.7586},
	"莆田":   {25.4540, 119.0078},
	"三明":   {26.2654, 117.6389},
	"泉州":   {24.8739, 118.6758},
	"漳州":   {24.5134, 117.6474},
	"南平":   {26.6418, 118.1784},
	"龙岩":   {25.0752, 117.0173},
	"宁德":   {26.6656, 119.5475},
	"景德镇":  {29.2688, 117.1786},
	"萍乡":   {27.6229, 113.8545},
	"九江":   {29.7049, 116.0018},
	"新余":   {27.8179, 114.9171},
	"鹰潭":   {28.2601, 117.0693},
	"赣州":   {25.8292, 114.9355},
	"吉安":   {27.1117, 114.9793},
	"宜春":   {27.8136, 114.4163},
	"抚州":   {27.9492, 116.3582},
	"上饶":   {28.4554, 117.9433},
	"烟台":   {37.4639, 121.4476},
	"潍坊":   {36.7068, 119.1619},
	"威海":   {37.5131, 122.1205},
	"淄博":   {36.8131, 118.0548},
	"临沂":   {35.1041, 118.3566},
	"枣庄":   {34.8107, 117.3237},
	"日照":   {35.4164, 119.5269},
	"东营":   {37.4346, 118.6747},
	"济宁":   {35.4147, 116.5871},
	"泰安":   {36.1949, 117.0892},
	"德州":   {37.4360, 116.3575},
	"聊城":   {36.4569, 115.9853},
	"滨州":   {37.3816, 117.9706},
	"菏泽":   {35.2333, 115.4412},
	"洛阳":   {34.6197, 112.4540},
	"开封":   {34.7971, 114.3074},
	"平顶山":  {33.7668, 113.1927},
	"安阳":   {36.0976, 114.3929},
	"鹤壁":   {35.7478, 114.2970},
	"新乡":   {35.3028, 113.9268},
	"焦作":   {35.2157, 113.2419},
	"濮阳":   {35.7616, 115.0293},
	"许昌":   {34.0267, 113.8526},
	"漯河":   {33.5818, 114.0420},
	"三门峡":  {34.7726, 111.1941},
	"南阳":   {32.9908, 112.5292},
	"商丘":   {34.4141, 115.6563},
	"信阳":   {32.1228, 114.0928},
	"周口":   {33.6237, 114.6498},
	"驻马店":  {32.9805, 114.0229},
	"十堰":   {32.6290, 110.7980},
	"宜昌":   {30.6918, 111.2862},
	"襄阳":   {32.0091, 112.1226},
	"鄂州":   {30.3911, 114.8947},
	"荆门":   {31.0354, 112.1991},
	"孝感":   {30.9279, 113.9268},
	"荆州":   {30.3269, 112.2394},
	"黄冈":   {30.4534, 114.8726},
	"咸宁":   {29.8415, 114.3225},
	"随州":   {31.6903, 113.3825},
	"恩施":   {30.2720, 109.4880},
	"韶关":   {24.8107, 113.5975},
	"汕头":   {23.3540, 116.6824},
	"江门":   {22.5789, 113.0816},
	"湛江":   {21.2707, 110.3594},
	"茂名":   {21.6631, 110.9253},
	"肇庆":   {23.0469, 112.4657},
	"惠州":   {23.1115, 114.4158},
	"梅州":   {24.2881, 116.1176},
	"汕尾":   {22.7861, 115.3644},
	"河源":   {23.7463, 114.7006},
	"阳江":   {21.8585, 111.9827},
	"清远":   {23.6820, 113.0510},
	"潮州":   {23.6618, 116.6225},
	"揭阳":   {23.5498, 116.3728},
	"云浮":   {22.9373, 112.0500},
	"柳州":   {24.3263, 109.4286},
	"桂林":   {25.2738, 110.2901},
	"梧州":   {23.4769, 111.2791},
	"北海":   {21.4734, 109.1193},
	"防城港":  {21.6174, 108.3543},
	"钦州":   {21.9674, 108.6548},
	"贵港":   {23.1113, 109.5986},
	"玉林":   {22.6541, 110.1818},
	"百色":   {23.9027, 106.6181},
	"贺州":   {24.4113, 111.5669},
	"河池":   {24.6929, 108.0853},
	"来宾":   {23.7500, 109.2216},
	"崇左":   {22.3765, 107.3650},
	"六盘水":  {26.5941, 104.8301},
	"遵义":   {27.7256, 106.9272},
	"安顺":   {26.2456, 105.9326},
	"毕节":   {27.3017, 105.2830},
	"铜仁":   {27.7183, 109.1912},
	"曲靖":   {25.4900, 103.7962},
	"玉溪":   {24.3518, 102.5457},
	"保山":   {25.1202, 99.1671},
	"昭通":   {27.3380, 103.7172},
	"丽江":   {26.8721, 100.2299},
	"普洱":   {22.7869, 100.9665},
	"临沧":   {23.8865, 100.0866},
	"大理":   {25.6066, 100.2676},
	"咸阳":   {34.3291, 108.7091},
	"渭南":   {34.4994, 109.5101},
	"铜川":   {34.8973, 108.9456},
	"宝鸡":   {34.3619, 107.2372},
	"商洛":   {33.8680, 109.9404},
	"榆林":   {38.2852, 109.7348},
	"延安":   {36.5853, 109.4898},
	"汉中":   {33.0677, 107.0230},
	"安康":   {32.6853, 109.0295},
	"石嘴山":  {38.9842, 106.3839},
	"吴忠":   {37.9976, 106.1991},
	"固原":   {36.0162, 106.2421},
	"中卫":   {37.5146, 105.1893},
	"海东":   {36.5023, 102.1040},
	"克拉玛依": {45.5798, 84.8892},
	"吐鲁番":  {42.9513, 89.1895},
	"哈密":   {42.8180, 93.5150},
	"昌吉":   {44.0145, 87.3087},
	"伊犁":   {43.9219, 81.3240},
	"石河子":  {44.3056, 86.0809},
}

// Weather 天气查询逻辑
type Weather struct{}

func NewWeather() *Weather {
	return &Weather{}
}

func (w *Weather) Name() string {
	return "weather"
}

func (w *Weather) Description() string {
	return "查询天气，输入格式：城市名称（支持：今天、明天、后天、七天），如：北京、上海、或者：北京明天、西安七天"
}

func (w *Weather) Execute(ctx context.Context, input string) string {
	// 解析城市和日期
	city, days := parseWeatherInput(input)
	if city == "" {
		city = input
		days = 1
	}

	// 查找城市坐标
	coords, ok := cityCoords[city]
	if !ok {
		// 回退到 wttr.in
		return getWeatherFromWttr(city, days)
	}

	// 获取天气预报（使用 Open-Meteo）
	weather, err := getWeatherFromOpenMeteo(coords[0], coords[1], days)
	if err != nil {
		// 如果 Open-Meteo 失败，回退到 wttr.in
		return getWeatherFromWttr(city, days)
	}

	return city + "的" + weather
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

	cityPatterns := []string{"今天", "明天", "后天", "七日", "七天", "一周", "7天", "天气", "未来"}
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

// getWeatherFromOpenMeteo 使用 Open-Meteo API 获取天气
func getWeatherFromOpenMeteo(lat, lon float64, days int) (string, error) {
	if days > 16 {
		days = 16
	}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&daily=weather_code,temperature_2m_max,temperature_2m_min&timezone=auto&forecast_days=%d", lat, lon, days)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("请求失败")
	}

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	daily, ok := data["daily"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("数据格式错误")
	}

	dates := daily["time"].([]interface{})
	weatherCodes := daily["weather_code"].([]interface{})
	maxTemps := daily["temperature_2m_max"].([]interface{})
	minTemps := daily["temperature_2m_min"].([]interface{})

	result := "未来" + fmt.Sprintf("%d", len(dates)) + "天天气：\n"

	for i := 0; i < len(dates); i++ {
		date := dates[i].(string)
		code := int(weatherCodes[i].(float64))
		maxTemp := int(maxTemps[i].(float64))
		minTemp := int(minTemps[i].(float64))

		weatherDesc := getWeatherDescByCode(code)

		dayStr := fmt.Sprintf("第%d天", i+1)
		if i == 0 {
			dayStr = "今天"
		} else if i == 1 {
			dayStr = "明天"
		} else if i == 2 {
			dayStr = "后天"
		} else if len(date) >= 10 {
			dayStr = date[5:10]
		}

		result += fmt.Sprintf("%s：%s，最高温度：%d°C，最低温度：%d°C\n", dayStr, weatherDesc, maxTemp, minTemp)
	}

	return strings.TrimSpace(result), nil
}

// getWeatherFromWttr 回退到 wttr.in
func getWeatherFromWttr(city string, days int) string {
	url := "https://wttr.in/" + city + "?format=j1"
	resp, err := http.Get(url)
	if err != nil {
		return "获取天气失败，请稍后重试"
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "获取天气失败，请稍后重试"
	}

	body, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "获取天气失败，请稍后重试"
	}

	weatherData, ok := data["weather"].([]interface{})
	if !ok || len(weatherData) == 0 {
		return "无法获取 " + city + " 的天气信息"
	}

	if days > 1 {
		actualDays := len(weatherData)
		if days > actualDays {
			days = actualDays
		}

		result := city + "未来" + fmt.Sprintf("%d", days) + "天天气：\n"

		for i := 0; i < days; i++ {
			day := weatherData[i].(map[string]interface{})
			maxTemp := getStringValue(day, "maxtempC")
			minTemp := getStringValue(day, "mintempC")
			weatherDesc := getWeatherDesc(day)

			dayStr := fmt.Sprintf("第%d天", i+1)
			if i == 0 {
				dayStr = "今天"
			} else if i == 1 {
				dayStr = "明天"
			} else if i == 2 {
				dayStr = "后天"
			}

			result += fmt.Sprintf("%s：%s，最高温度：%s°C，最低温度：%s°C\n", dayStr, weatherDesc, maxTemp, minTemp)
		}

		return strings.TrimSpace(result)
	}

	// 单天天气
	today := weatherData[0].(map[string]interface{})
	maxTemp := getStringValue(today, "maxtempC")
	minTemp := getStringValue(today, "mintempC")
	weatherDesc := getWeatherDesc(today)

	humidity := ""
	wind := ""
	current, ok := data["current_condition"].([]interface{})
	if ok && len(current) > 0 {
		condition := current[0].(map[string]interface{})
		humidity = getStringValue(condition, "humidity")
		wind = getStringValue(condition, "windspeedKmph")
	}

	result := city + "今日天气：" + weatherDesc + "，最高温度：" + maxTemp + "°C，最低温度：" + minTemp + "°C"
	if humidity != "" {
		result += "，湿度：" + humidity + "%"
	}
	if wind != "" {
		result += "，风速：" + wind + "km/h"
	}

	return result
}

// getWeatherDescByCode 根据天气码获取天气描述
func getWeatherDescByCode(code int) string {
	codes := map[int]string{
		0:  "晴天",
		1:  "晴朗",
		2:  "多云",
		3:  "阴天",
		45: "雾",
		48: "雾",
		51: "小雨",
		53: "中雨",
		55: "大雨",
		61: "小雨",
		63: "中雨",
		65: "大雨",
		71: "小雪",
		73: "中雪",
		75: "大雪",
		77: "雪",
		80: "小雨",
		81: "中雨",
		82: "大雨",
		85: "小雪",
		86: "大雪",
		95: "雷暴",
		96: "雷暴",
		99: "雷暴",
	}
	if desc, ok := codes[code]; ok {
		return desc
	}
	return fmt.Sprintf("天气%d", code)
}

// getWeatherDesc 获取天气描述（wttr.in）
func getWeatherDesc(day map[string]interface{}) string {
	hourly, ok := day["hourly"].([]interface{})
	if ok && len(hourly) > 0 {
		firstHour := hourly[0].(map[string]interface{})
		if desc, ok := firstHour["weatherDesc"].([]interface{}); ok && len(desc) > 0 {
			if v, ok := desc[0].(map[string]interface{})["value"].(string); ok {
				return translateWeather(v)
			}
		}
	}
	return "未知"
}

// getStringValue 安全获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return fmt.Sprintf("%.0f", v)
	}
	return ""
}

// translateWeather 翻译天气描述
func translateWeather(desc string) string {
	desc = strings.TrimSpace(desc)

	translations := map[string]string{
		"Sunny":                "晴天",
		"Clear":                "晴天",
		"Partly cloudy":        "多云",
		"Partly Cloudy":        "多云",
		"Cloudy":               "阴天",
		"Overcast":             "阴天",
		"Light rain":           "小雨",
		"Light rain shower":    "小雨",
		"Moderate rain":        "中雨",
		"Heavy rain":           "大雨",
		"Rain":                 "雨天",
		"Patchy rain possible": "可能有小雨",
		"Snow":                 "雪天",
		"Light snow":           "小雪",
		"Fog":                  "雾",
		"Foggy":                "雾",
		"Mist":                 "薄雾",
		"Thunderstorm":         "雷暴",
		"Thunder":              "雷暴",
	}
	if t, ok := translations[desc]; ok {
		return t
	}
	return desc
}
