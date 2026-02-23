package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Project-DSView/backend/go/internal/infrastructure/config"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type SystemHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewSystemHandler(db *gorm.DB, cfg *config.Config) *SystemHandler {
	return &SystemHandler{db: db, cfg: cfg}
}

// HealthCheck godoc
// @Summary Health probe
// @Description Check if the service process is running and dependencies are reachable
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health [get]
func (h *SystemHandler) HealthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	deps := fiber.Map{}
	hasError := false

	// 1. Check Postgres
	dbCheck := "ok"
	sqlDB, err := h.db.DB()
	if err != nil {
		dbCheck = "down"
		hasError = true
	} else if err := sqlDB.PingContext(ctx); err != nil {
		dbCheck = "down"
		hasError = true
	}
	deps["postgres"] = dbCheck

	// 2. Check MinIO
	minioCheck := "ok"
	minioUrl := "http://" + h.cfg.MinIO.Endpoint + "/minio/health/live"
	if h.cfg.MinIO.UseSSL {
		minioUrl = "https://" + h.cfg.MinIO.Endpoint + "/minio/health/live"
	}
	reqMinio, _ := http.NewRequestWithContext(ctx, "GET", minioUrl, nil)
	client := &http.Client{}
	if resp, err := client.Do(reqMinio); err != nil || resp.StatusCode != 200 {
		minioCheck = "down"
		hasError = true
	} else {
		resp.Body.Close()
	}
	deps["minio"] = minioCheck

	// 3. Check RabbitMQ (Management API ping)
	rabbitCheck := "ok"
	// Config usually has AMQP URL, we can check the management port instead if available, or just a generic TCP dial.
	// We'll use the management API endpoint usually at port 15672
	rabbitUrl := fmt.Sprintf("http://%s:%s/api/aliveness-test/%%2F", h.cfg.RabbitMQ.Host, "15672")
	reqRabbit, _ := http.NewRequestWithContext(ctx, "GET", rabbitUrl, nil)
	// We need to pass basic auth for the aliveness test
	reqRabbit.SetBasicAuth(h.cfg.RabbitMQ.Username, h.cfg.RabbitMQ.Password)
	if resp, err := client.Do(reqRabbit); err != nil || resp.StatusCode != 200 {
		rabbitCheck = "down"
		hasError = true
	} else {
		resp.Body.Close()
	}
	deps["rabbitmq"] = rabbitCheck

	// 4. Check Docker-DinD (usually at tcp://docker-dind:2375/_ping)
	dockerCheck := "ok"
	// The frontend/backend usually accesses it via a host named 'docker-dind'. We will assume it's on port 2375.
	reqDocker, _ := http.NewRequestWithContext(ctx, "GET", "http://docker-dind:2375/_ping", nil)
	if resp, err := client.Do(reqDocker); err != nil || resp.StatusCode != 200 {
		dockerCheck = "down"
		hasError = true
	} else {
		resp.Body.Close()
	}
	deps["docker-dind"] = dockerCheck

	status := "ok"
	statusCode := fiber.StatusOK
	if hasError {
		status = "error"
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(fiber.Map{
		"status":       status,
		"service":      "dsview-go",
		"version":      "1.0.0",
		"env":          h.cfg.Server.Environment,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"dependencies": deps,
	})
}
