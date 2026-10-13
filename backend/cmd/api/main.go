package main

import (
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-trackclientservice/internal/clients"
	"go-trackclientservice/internal/config"
	"go-trackclientservice/internal/db"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DSN())
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Printf("mysql ยังต่อไม่ได้: %v", err)
		log.Printf("ตรวจว่า MySQL เปิดอยู่ และค่าใน .env ถูก (host=%s port=%s db=%s user=%s)",
			cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser)
	} else {
		log.Println("mysql: ต่อได้")
	}

	r := gin.Default()

	clientHandler := clients.NewHandler(clients.NewRepository(database))
	clientHandler.RegisterRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "service": "api"})
	})

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from my first Go API!"})
	})

	r.GET("/add/:number", func(c *gin.Context) {
		number, err := strconv.Atoi(c.Param("number"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "number ต้องเป็นจำนวนเต็ม เช่น /add/5"})
			return
		}

		total := number + 10
		message := "ผลรวมน้อยกว่าหรือเท่ากับ 20"
		if total > 20 {
			message = "ผลรวมมากกว่า 20"
		}

		c.JSON(http.StatusOK, gin.H{
			"number":  number,
			"added":   10,
			"total":   total,
			"message": message,
		})
	})

	r.GET("/health/db", func(c *gin.Context) {
		if err := database.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "database": "down", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "database": "up"})
	})

	addr := net.JoinHostPort(cfg.HTTPHost, cfg.HTTPPort)
	log.Printf("listening on http://%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
