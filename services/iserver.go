package services

import (
	"fmt"
	"net"
	"os/exec"
	"syscall"
	"time"
	"webapi/config"
	"webapi/models"
)

type IServerService struct {
	cfg *config.Config
}

func NewIServerService(cfg *config.Config) *IServerService {
	return &IServerService{cfg: cfg}
}

// RestartAsync перезапускает или останавливает IServer
func (s *IServerService) RestartAsync(stopOnly bool) (*models.CommandResponse, error) {
	// Проверяем, запущен ли IServer
	isRunning := s.isProcessRunning()
	
	if !isRunning && stopOnly {
		return &models.CommandResponse{
			Success: true,
			Message: "IServer.exe is not running",
		}, nil
	}
	
	if isRunning {
		// Останавливаем процесс
		err := s.stopIServer()
		if err != nil {
			return &models.CommandResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to stop: %v", err),
			}, nil
		}
		time.Sleep(2 * time.Second)

		// Проверяем, остановился ли
		if s.isProcessRunning() {
			return &models.CommandResponse{
				Success: false,
				Message: "Process still running after stop attempt",
			}, nil
		}
		
		// Если только остановка - завершаем
		if stopOnly {
			return &models.CommandResponse{
				Success: true,
				Message: "IServer.exe stopped successfully!",
			}, nil
		}
	}
	
	// Если stopOnly = false и процесс не запущен - запускаем
	if !stopOnly {
		err := s.startIServer()
		if err != nil {
			return &models.CommandResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to start: %v", err),
			}, nil
		}
		time.Sleep(3 * time.Second)

		// Проверяем, запустился ли
		if s.isProcessRunning() {
			return &models.CommandResponse{
				Success: true,
				Message: "IServer.exe restarted successfully!",
			}, nil
		} else {
			return &models.CommandResponse{
				Success: false,
				Message: "IServer.exe failed to start",
			}, nil
		}
	}
	
	return &models.CommandResponse{
		Success: true,
		Message: "Operation completed",
	}, nil
}

func (s *IServerService) GetHealthStatus() *models.HealthResponse {
	isRunning := s.isProcessRunning()
	status := "🔴 IServer.exe не запущен"
	if isRunning {
		status = "🟢 IServer.exe запущен"
	}
	return &models.HealthResponse{
		Status:    status,
		IsRunning: isRunning,
	}
}

func (s *IServerService) GetTimeDifferenceAsync() (*models.TimeDiffResponse, error) {
	systemTime := time.Now()

	ntpTimeUtc, err := s.getNtpTime()
	if err != nil {
		return &models.TimeDiffResponse{
			Success:      false,
			TimeDiffNs:   0,
			TimeDiffStr:  "N/A",
			SystemTime:   systemTime,
			NtpTime:      time.Time{},
			ErrorMessage: err.Error(),
		}, nil
	}

	systemTimeUtc := systemTime.UTC()
	timeDiff := ntpTimeUtc.Sub(systemTimeUtc)

	return &models.TimeDiffResponse{
		Success:     true,
		TimeDiffNs:  timeDiff.Nanoseconds(),
		TimeDiffStr: formatDuration(timeDiff),
		SystemTime:  systemTime,
		NtpTime:     ntpTimeUtc,
	}, nil
}

func (s *IServerService) isProcessRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq IServer.exe")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return stringsContains(string(output), "IServer.exe")
}

func (s *IServerService) stopIServer() error {
	cmd := exec.Command("taskkill", "/F", "/IM", "IServer.exe")
	return cmd.Run()
}

func (s *IServerService) startIServer() error {
	cmd := exec.Command(s.cfg.IServer.Path)
	cmd.Dir = "C:\\BeeDotNet\\IServer"
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: false,
	}
	return cmd.Start()
}

func (s *IServerService) getNtpTime() (time.Time, error) {
	// Используем NTP сервер из конфига
	ntpServer := s.cfg.NTP.Server
	ntpData := make([]byte, 48)
	ntpData[0] = 0x1B

	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:123", ntpServer), 5*time.Second)
	if err != nil {
		return time.Time{}, err
	}
	defer conn.Close()

	_, err = conn.Write(ntpData)
	if err != nil {
		return time.Time{}, err
	}

	_, err = conn.Read(ntpData)
	if err != nil {
		return time.Time{}, err
	}

	intPart := uint64(ntpData[40])<<24 | uint64(ntpData[41])<<16 | uint64(ntpData[42])<<8 | uint64(ntpData[43])
	fractPart := uint64(ntpData[44])<<24 | uint64(ntpData[45])<<16 | uint64(ntpData[46])<<8 | uint64(ntpData[47])

	milliseconds := (intPart * 1000) + ((fractPart * 1000) / 0x100000000)
	ntpEpoch := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

	return ntpEpoch.Add(time.Duration(milliseconds) * time.Millisecond), nil
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "-" + formatDuration(-d)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	millis := d.Milliseconds() % 1000
	
	if millis > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%03d0000", hours, minutes, seconds, millis)
	}
	return fmt.Sprintf("%02d:%02d:%02d.0000000", hours, minutes, seconds)
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || stringsContainsHelper(s, substr)))
}

func stringsContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}