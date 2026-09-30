package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/pkg/response"
)

// 翻译代理：MyMemory 为主（国内可访问），Google gtx 为备（部署在无墙环境时自动生效）。
// 服务端做内存缓存，避免阅读时同一单词/句子反复请求。

var (
	transMu    sync.Mutex
	transCache = make(map[string]string)
)

func cacheGet(key string) (string, bool) {
	transMu.Lock()
	defer transMu.Unlock()
	v, ok := transCache[key]
	return v, ok
}

func cacheSet(key, val string) {
	transMu.Lock()
	defer transMu.Unlock()
	if len(transCache) > 20000 { // 简单防膨胀
		transCache = make(map[string]string)
	}
	transCache[key] = val
}

var translateHTTPClient = &http.Client{Timeout: 8 * time.Second}

// Translate GET /api/v1/translate?text=...&from=en|de
func Translate(c *gin.Context) {
	text := strings.TrimSpace(c.Query("text"))
	from := c.Query("from")
	if from != model.LangDE {
		from = model.LangEN
	}
	if text == "" {
		response.BadRequest(c, "缺少待翻译文本")
		return
	}
	if len([]rune(text)) > 450 {
		response.BadRequest(c, "文本过长，请分段翻译")
		return
	}

	key := from + ":" + strings.ToLower(text)
	if v, ok := cacheGet(key); ok {
		response.OK(c, gin.H{"text": v, "cached": true})
		return
	}

	result, err := translateText(text, from)
	if err != nil {
		response.Fail(c, http.StatusBadGateway, "翻译服务暂时不可用："+err.Error())
		return
	}
	cacheSet(key, result)
	response.OK(c, gin.H{"text": result, "cached": false})
}

func translateText(text, from string) (string, error) {
	if s, err := viaMyMemory(text, from); err == nil {
		return s, nil
	}
	return viaGoogle(text, from)
}

// viaMyMemory https://mymemory.translated.net（免密钥，国内可访问）
func viaMyMemory(text, from string) (string, error) {
	u := "https://api.mymemory.translated.net/get?q=" + url.QueryEscape(text) +
		"&langpair=" + url.QueryEscape(from+"|zh-CN")

	body, err := httpGet(u)
	if err != nil {
		return "", err
	}

	var payload struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
			ResponseStatus any    `json:"responseStatus"`
		} `json:"responseData"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	t := strings.TrimSpace(payload.ResponseData.TranslatedText)
	if t == "" || strings.Contains(strings.ToLower(t), "mymemory warning") {
		return "", errors.New("翻译结果为空")
	}
	// MyMemory 查不到时会把原文全大写原样返回，视为失败走后备
	if t == strings.ToUpper(text) {
		return "", errors.New("未找到译文")
	}
	return t, nil
}

// viaGoogle 非官方 gtx 接口（需可访问 google）
func viaGoogle(text, from string) (string, error) {
	u := "https://translate.googleapis.com/translate_a/single?client=gtx" +
		"&sl=" + from + "&tl=zh-CN&dt=t&q=" + url.QueryEscape(text)

	body, err := httpGet(u)
	if err != nil {
		return "", err
	}

	var payload []any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if len(payload) == 0 {
		return "", errors.New("响应格式异常")
	}
	segs, ok := payload[0].([]any)
	if !ok || len(segs) == 0 {
		return "", errors.New("响应格式异常")
	}
	var sb strings.Builder
	for _, seg := range segs {
		arr, ok := seg.([]any)
		if !ok || len(arr) == 0 {
			continue
		}
		if s, ok := arr[0].(string); ok {
			sb.WriteString(s)
		}
	}
	t := strings.TrimSpace(sb.String())
	if t == "" {
		return "", errors.New("未找到译文")
	}
	return t, nil
}

func httpGet(u string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp, err := translateHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("上游返回 " + resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}
