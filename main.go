package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(detailedRequestLogger())

	r.GET("/", handleRoot)
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/ping", handlePing)
	r.Any("/echo/*path", handleEcho)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("[STARTUP] Network Receiver запущен на %s", addr)
	log.Printf("[STARTUP] Эндпоинты:")
	log.Printf("[STARTUP]   GET /ping     — простая проверка связи")
	log.Printf("[STARTUP]   ANY /echo/*   — эхо-эндпоинт, возвращает детали запроса")
	log.Printf("[STARTUP]   GET /healthz  — проверка здоровья")
	if err := r.Run(addr); err != nil {
		log.Fatalf("[STARTUP] Ошибка запуска сервера: %v", err)
	}
}

func detailedRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		log.Printf("[INCOMING] ======================================")
		log.Printf("[INCOMING] Входящий запрос:")
		log.Printf("[INCOMING]   Метод: %s", c.Request.Method)
		log.Printf("[INCOMING]   Путь: %s", c.Request.URL.String())
		log.Printf("[INCOMING]   IP клиента: %s", c.ClientIP())
		log.Printf("[INCOMING]   Remote Addr: %s", c.Request.RemoteAddr)
		log.Printf("[INCOMING]   Host: %s", c.Request.Host)
		log.Printf("[INCOMING]   Proto: %s", c.Request.Proto)
		log.Printf("[INCOMING]   Content-Length: %d", c.Request.ContentLength)

		log.Printf("[INCOMING]   Заголовки:")
		for name, values := range c.Request.Header {
			log.Printf("[INCOMING]     %s: %s", name, strings.Join(values, ", "))
		}

		c.Next()

		elapsed := time.Since(start)
		log.Printf("[INCOMING]   Ответ: статус %d, время обработки: %s", c.Writer.Status(), elapsed)
		log.Printf("[INCOMING] ======================================")
	}
}

func handleRoot(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Network Receiver</title></head>
<body>
<h1>Network Receiver</h1>
<p>Принимает входящие запросы и подробно логирует их.</p>
<h3>Эндпоинты:</h3>
<ul>
  <li><b>GET /ping</b> — простая проверка связи, возвращает "pong" и время</li>
  <li><b>ANY /echo/*</b> — эхо-эндпоинт, возвращает все детали входящего запроса</li>
  <li><b>GET /healthz</b> — проверка здоровья</li>
</ul>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func handlePing(c *gin.Context) {
	now := time.Now()
	log.Printf("[PING] Получен ping от %s в %s", c.ClientIP(), now.Format(time.RFC3339Nano))

	c.JSON(http.StatusOK, gin.H{
		"message":    "pong",
		"timestamp":  now.Format(time.RFC3339Nano),
		"your_ip":    c.ClientIP(),
		"server_port": c.Request.Host,
	})

	log.Printf("[PING] Отправлен pong клиенту %s", c.ClientIP())
}

func handleEcho(c *gin.Context) {
	now := time.Now()
	path := c.Param("path")

	headers := make(map[string]string)
	for name, values := range c.Request.Header {
		headers[name] = strings.Join(values, ", ")
	}

	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		queryParams[key] = strings.Join(values, ", ")
	}

	log.Printf("[ECHO] Получен запрос %s /echo%s от %s", c.Request.Method, path, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":      "echo",
		"timestamp":    now.Format(time.RFC3339Nano),
		"your_ip":      c.ClientIP(),
		"method":       c.Request.Method,
		"path":         "/echo" + path,
		"headers":      headers,
		"query_params": queryParams,
		"host":         c.Request.Host,
		"proto":        c.Request.Proto,
	})

	log.Printf("[ECHO] Отправлен echo-ответ клиенту %s", c.ClientIP())
}
