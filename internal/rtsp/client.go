package rtsp

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/local-video-server/pkg/utils"
)

// Client представляет RTSP клиент
type Client struct {
	conn       net.Conn
	reader     *bufio.Reader
	url        *url.URL
	username   string
	password   string
	timeout    time.Duration
	sessionID  string
	cseq       int
	authMethod string
	realm      string
	nonce      string
}

// NewClient создает новый RTSP клиент
func NewClient(rtspURL string, username, password string, timeout time.Duration) (*Client, error) {
	parsedURL, err := url.Parse(rtspURL)
	if err != nil {
		return nil, fmt.Errorf("invalid RTSP URL: %w", err)
	}

	// Определяем порт
	port := parsedURL.Port()
	if port == "" {
		port = "554"
	}

	// Подключаемся к серверу
	address := net.JoinHostPort(parsedURL.Hostname(), port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := &Client{
		conn:     conn,
		reader:   bufio.NewReader(conn),
		url:      parsedURL,
		username: username,
		password: password,
		timeout:  timeout,
		cseq:     1,
	}

	// Устанавливаем таймаут на соединение
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	return client, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sendRequest отправляет RTSP запрос
func (c *Client) sendRequest(method, path string, headers map[string]string) (*Response, error) {
	// Увеличиваем CSeq
	c.cseq++

	// Формируем запрос
	request := fmt.Sprintf("%s %s RTSP/1.0\r\n", method, path)
	request += fmt.Sprintf("CSeq: %d\r\n", c.cseq)

	// Добавляем заголовки
	for key, value := range headers {
		request += fmt.Sprintf("%s: %s\r\n", key, value)
	}

	// Добавляем User-Agent
	request += "User-Agent: Local-video-server/1.0\r\n"

	// Добавляем аутентификацию, если требуется
	if c.authMethod != "" && c.username != "" && c.password != "" {
		authHeader := c.buildAuthHeader(method, path)
		if authHeader != "" {
			request += fmt.Sprintf("Authorization: %s\r\n", authHeader)
		}
	}

	request += "\r\n"

	// Отправляем запрос
	if _, err := c.conn.Write([]byte(request)); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Обновляем таймаут
	if err := c.conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	// Читаем ответ
	return c.readResponse()
}

// readResponse читает RTSP ответ
func (c *Client) readResponse() (*Response, error) {
	// Читаем первую строку (статус)
	statusLine, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read status line: %w", err)
	}
	statusLine = strings.TrimSpace(statusLine)

	// Парсим статус
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid status line: %s", statusLine)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %w", err)
	}

	response := &Response{
		StatusCode: statusCode,
		StatusText: parts[2],
		Headers:   make(map[string]string),
	}

	// Читаем заголовки
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// Парсим заголовок
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			response.Headers[key] = value
		}
	}

	// Читаем тело, если есть Content-Length
	if contentLengthStr, ok := response.Headers["Content-Length"]; ok {
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err == nil && contentLength > 0 {
			body := make([]byte, contentLength)
			if _, err := io.ReadFull(c.reader, body); err != nil {
				return nil, fmt.Errorf("failed to read body: %w", err)
			}
			response.Body = string(body)
		}
	}

	return response, nil
}

// Response представляет RTSP ответ
type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       string
}

// Options отправляет OPTIONS запрос
func (c *Client) Options() (*Response, error) {
	return c.sendRequest("OPTIONS", "*", nil)
}

// Describe отправляет DESCRIBE запрос
func (c *Client) Describe() (*Response, error) {
	path := c.url.Path
	if path == "" {
		path = "/"
	}

	headers := map[string]string{
		"Accept": "application/sdp",
	}

	response, err := c.sendRequest("DESCRIBE", path, headers)
	if err != nil {
		return nil, err
	}

	// Обрабатываем аутентификацию
	if response.StatusCode == 401 {
		// Парсим WWW-Authenticate заголовок
		authHeader := response.Headers["WWW-Authenticate"]
		if authHeader != "" {
			c.parseAuthHeader(authHeader)
			// Повторяем запрос с аутентификацией
			return c.sendRequest("DESCRIBE", path, headers)
		}
	}

	return response, nil
}

// Setup отправляет SETUP запрос
func (c *Client) Setup(transport string) (*Response, error) {
	path := c.url.Path
	if path == "" {
		path = "/"
	}

	headers := map[string]string{
		"Transport": transport,
	}

	response, err := c.sendRequest("SETUP", path, headers)
	if err != nil {
		return nil, err
	}

	// Извлекаем Session ID
	if sessionID, ok := response.Headers["Session"]; ok {
		// Session может быть в формате "Session: 123456; timeout=60"
		parts := strings.Split(sessionID, ";")
		c.sessionID = strings.TrimSpace(parts[0])
	}

	return response, nil
}

// Play отправляет PLAY запрос
func (c *Client) Play() (*Response, error) {
	path := c.url.Path
	if path == "" {
		path = "/"
	}

	headers := make(map[string]string)
	if c.sessionID != "" {
		headers["Session"] = c.sessionID
	}

	return c.sendRequest("PLAY", path, headers)
}

// parseAuthHeader парсит заголовок WWW-Authenticate
func (c *Client) parseAuthHeader(header string) {
	// Пример: Digest realm="IP Camera(12345)", nonce="abc123", qop="auth"
	// Или: Basic realm="IP Camera"

	if strings.HasPrefix(header, "Basic") {
		c.authMethod = "Basic"
		// Извлекаем realm
		if idx := strings.Index(header, "realm="); idx != -1 {
			realm := header[idx+6:]
			if endIdx := strings.Index(realm, ","); endIdx != -1 {
				realm = realm[:endIdx]
			}
			realm = strings.Trim(realm, "\"")
			c.realm = realm
		}
	} else if strings.HasPrefix(header, "Digest") {
		c.authMethod = "Digest"
		// Извлекаем realm и nonce
		parts := strings.Split(header, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "realm=") {
				c.realm = strings.Trim(part[6:], "\"")
			} else if strings.HasPrefix(part, "nonce=") {
				c.nonce = strings.Trim(part[6:], "\"")
			}
		}
	}
}

// buildAuthHeader создает заголовок Authorization
func (c *Client) buildAuthHeader(method, path string) string {
	if c.authMethod == "Basic" {
		// Basic аутентификация
		auth := c.username + ":" + c.password
		encoded := base64.StdEncoding.EncodeToString([]byte(auth))
		return fmt.Sprintf("Basic %s", encoded)
	} else if c.authMethod == "Digest" {
		// Digest аутентификация
		ha1 := md5Hash(fmt.Sprintf("%s:%s:%s", c.username, c.realm, c.password))
		ha2 := md5Hash(fmt.Sprintf("%s:%s", method, path))
		response := md5Hash(fmt.Sprintf("%s:%s:%s", ha1, c.nonce, ha2))

		return fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
			c.username, c.realm, c.nonce, path, response)
	}

	return ""
}

// md5Hash вычисляет MD5 хеш
func md5Hash(data string) string {
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// CheckStream проверяет RTSP поток и возвращает информацию о нем
func CheckStream(rtspURL string, username, password string, timeout time.Duration) (*StreamInfo, error) {
	logger := utils.GetLogger()

	client, err := NewClient(rtspURL, username, password, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create RTSP client: %w", err)
	}
	defer client.Close()

	// Отправляем OPTIONS
	logger.Debugf("Sending OPTIONS to %s", rtspURL)
	optionsResp, err := client.Options()
	if err != nil {
		return nil, fmt.Errorf("OPTIONS failed: %w", err)
	}

	if optionsResp.StatusCode != 200 {
		return nil, fmt.Errorf("OPTIONS returned status %d: %s", optionsResp.StatusCode, optionsResp.StatusText)
	}

	// Отправляем DESCRIBE
	logger.Debugf("Sending DESCRIBE to %s", rtspURL)
	describeResp, err := client.Describe()
	if err != nil {
		return nil, fmt.Errorf("DESCRIBE failed: %w", err)
	}

	if describeResp.StatusCode != 200 {
		return nil, fmt.Errorf("DESCRIBE returned status %d: %s", describeResp.StatusCode, describeResp.StatusText)
	}

	// Парсим SDP
	if describeResp.Body == "" {
		return nil, fmt.Errorf("empty SDP response")
	}

	streamInfo, err := ParseSDP(describeResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDP: %w", err)
	}

	streamInfo.URL = rtspURL
	streamInfo.Available = true

	return streamInfo, nil
}
