package service

import (
	"biostat/config"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
		for {
			status := h.checkHealth()
			ctx := context.Background()
			if err := h.redisClient.Set(ctx, h.cacheKey, status, time.Duration(config.PropConfig.HealthCheck.Expiration)*time.Second).Err(); err != nil {
				config.Log.Warn("[HealthMonitor] Failed to set Redis key", zap.Error(err))
			} else {
				config.Log.Debug("[HealthMonitor] Service status cached", zap.String("status", status), zap.Int("expires_in_sec", config.PropConfig.HealthCheck.Expiration))
			}
			time.Sleep(h.interval)
		}
	}()
}

func (h *HealthMonitorService) checkHealth() string {
	client := http.Client{Timeout: h.timeout}
	resp, err := client.Get(h.serviceURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err != nil {
			log.Printf("[HealthMonitor] Health check failed: %v", err)
			config.Log.Warn("[HealthMonitor] Health check failed", zap.Error(err))
		} else {
			config.Log.Info("[HealthMonitor] Health check returned non-200 status", zap.Int("status_code", resp.StatusCode))
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
