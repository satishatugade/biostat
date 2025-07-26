package service

import (
	"biostat/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type HealthMonitorService struct {
	redisClient *redis.Client
	serviceURL  string
	cacheKey    string
	interval    time.Duration
	timeout     time.Duration
}

func NewHealthMonitorService(redisClient *redis.Client, serviceURL string, interval, timeout time.Duration) *HealthMonitorService {
	return &HealthMonitorService{
		redisClient: redisClient,
		serviceURL:  serviceURL,
		cacheKey:    fmt.Sprintf("ai-service_status:%s", serviceURL),
		interval:    interval,
		timeout:     timeout,
	}
}

func (h *HealthMonitorService) Start() {
	go func() {
		log.Println("[HealthMonitor] Starting background health check...")
		for {
			status := h.checkHealth()
			ctx := context.Background()

			if err := h.redisClient.Set(ctx, h.cacheKey, status, time.Duration(config.PropConfig.HealthCheck.Expiration)*time.Second).Err(); err != nil {
				log.Printf("[HealthMonitor] Failed to set Redis key: %v", err)
			} else {
				log.Printf("[HealthMonitor] Service status '%s' cached for %ds", status, config.PropConfig.HealthCheck.Expiration)
			}
			time.Sleep(h.interval)
		}
	}()
}

func (h *HealthMonitorService) checkHealth() string {
	log.Println("health check URL : ", h.serviceURL)
	client := http.Client{Timeout: h.timeout}
	resp, err := client.Get(h.serviceURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err != nil {
			log.Printf("[HealthMonitor] Health check failed: %v", err)
		} else {
			log.Printf("[HealthMonitor] Health check returned non-200 status: %d", resp.StatusCode)
		}
		return "down"
	}
	defer resp.Body.Close()
	return "up"
}

func (h *HealthMonitorService) IsServiceUp() bool {
	val, err := h.redisClient.Get(context.Background(), h.cacheKey).Result()
	if err != nil {
		log.Printf("[HealthMonitor] Redis read failed: %v. Assuming service is DOWN.", err)
		return false
	}
	return val == "up"
}
