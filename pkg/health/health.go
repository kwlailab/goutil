package health

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type HealthCheck struct {
	once             sync.Once
	healthStatus     grpc_health_v1.HealthCheckResponse_ServingStatus
	mapWatchChannels map[string]chan grpc_health_v1.HealthCheckResponse_ServingStatus
	chHealthStatus   chan grpc_health_v1.HealthCheckResponse_ServingStatus
}

func (hc *HealthCheck) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	healthStatus := hc.healthStatus
	if healthStatus == grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN {
		healthStatus = grpc_health_v1.HealthCheckResponse_UNKNOWN
	}
	return &grpc_health_v1.HealthCheckResponse{
		Status: healthStatus,
	}, nil
}

func (hc *HealthCheck) Watch(req *grpc_health_v1.HealthCheckRequest, rsp grpc_health_v1.Health_WatchServer) error {
	rsp.Send(&grpc_health_v1.HealthCheckResponse{Status: hc.healthStatus})
	key, err := uuid.NewV7()
	if err != nil {
		return err
	}
	chWatch := make(chan grpc_health_v1.HealthCheckResponse_ServingStatus)
	hc.mapWatchChannels[key.String()] = chWatch
	for {
		healthStatus, ok := <-chWatch
		if !ok {
			return nil // 服務關閉
		} else {
			rsp.Send(&grpc_health_v1.HealthCheckResponse{Status: healthStatus})
		}
	}
}

func (hc *HealthCheck) List(ctx context.Context, req *grpc_health_v1.HealthListRequest) (*grpc_health_v1.HealthListResponse, error) {
	// 返回當前服務的健康狀態
	statuses := make(map[string]*grpc_health_v1.HealthCheckResponse)
	healthStatus := hc.healthStatus
	if healthStatus == grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN {
		healthStatus = grpc_health_v1.HealthCheckResponse_UNKNOWN
	}
	statuses[""] = &grpc_health_v1.HealthCheckResponse{
		Status: healthStatus,
	}
	return &grpc_health_v1.HealthListResponse{
		Statuses: statuses,
	}, nil
}

func (hc *HealthCheck) SetHealthStatus(healthStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
	hc.once.Do(func() {
		go func() {
			for {
				healthStatus, ok := <-hc.chHealthStatus
				if !ok {
					for _, ch := range hc.mapWatchChannels {
						close(ch)
					}
					hc.mapWatchChannels = make(map[string]chan grpc_health_v1.HealthCheckResponse_ServingStatus, 0)

					return // 服務關閉
				} else {
					hc.healthStatus = healthStatus
					for _, ch := range hc.mapWatchChannels {
						ch <- healthStatus
					}
				}
			}
		}()
	})
	if hc.chHealthStatus != nil {
		hc.chHealthStatus <- healthStatus
	}
}

func (hc *HealthCheck) End() {
	close(hc.chHealthStatus)
}

func NewHealthCheck(initStatus grpc_health_v1.HealthCheckResponse_ServingStatus) *HealthCheck {
	return &HealthCheck{
		once:             sync.Once{},
		healthStatus:     initStatus,
		chHealthStatus:   make(chan grpc_health_v1.HealthCheckResponse_ServingStatus),
		mapWatchChannels: make(map[string]chan grpc_health_v1.HealthCheckResponse_ServingStatus, 0),
	}
}

var _defaultHealthCheck = NewHealthCheck(grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN)

func DefaultHealthCheck() *HealthCheck {
	return _defaultHealthCheck
}

func Check() bool {
	return _defaultHealthCheck.healthStatus == grpc_health_v1.HealthCheckResponse_SERVING
}

func SetHealthStatus(healthStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
	_defaultHealthCheck.SetHealthStatus(healthStatus)
}

func SetHealthStatusServing() {
	_defaultHealthCheck.SetHealthStatus(grpc_health_v1.HealthCheckResponse_SERVING)
}

func HealthStatusEnd() {
	_defaultHealthCheck.End()
}
