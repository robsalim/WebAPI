package controllers

import (
	"net/http"
	"os"
	"strconv"
	"webapi/config"
	"webapi/models"
	"webapi/services"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	cfg                *config.Config
	dbHealthService    *services.DatabaseHealthService
	iserverService     *services.IServerService
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{
		cfg:             cfg,
		dbHealthService: services.NewDatabaseHealthService(cfg),
		iserverService:  services.NewIServerService(cfg),
	}
}

// Health endpoint - GET /health
// Возвращает JSON как в документации
func (h *Handlers) Health(c *gin.Context) {
	result := h.iserverService.GetHealthStatus()
	//c.JSON(http.StatusOK, result)
	c.String(http.StatusOK, result.Status)
}



// Restart endpoint - POST /restart
// Поддерживает параметр stopOnly (true - только остановка, false - полный перезапуск)
func (h *Handlers) Restart(c *gin.Context) {
	stopOnly := false
	if stopOnlyParam := c.Query("stopOnly"); stopOnlyParam != "" {
		stopOnly, _ = strconv.ParseBool(stopOnlyParam)
	}
	
	result, err := h.iserverService.RestartAsync(stopOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

// TimeDiff endpoint - GET /time-diff
func (h *Handlers) TimeDiff(c *gin.Context) {
	result, err := h.iserverService.GetTimeDifferenceAsync()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error_message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

// CheckPrivileges endpoint - GET /debug/privileges
func (h *Handlers) CheckPrivileges(c *gin.Context) {
	user := os.Getenv("USERNAME")
	if user == "" {
		user = os.Getenv("USER")
	}
	
	currentDir, _ := os.Getwd()
	
	result := models.PrivilegesResponse{
		User:             user,
		IsAdministrator:  h.cfg.IsAdmin,
		ProcessId:        os.Getpid(),
		CurrentDirectory: currentDir,
	}
	
	c.JSON(http.StatusOK, result)
}

// GetDatabaseHealth endpoint - GET /db-health
func (h *Handlers) GetDatabaseHealth(c *gin.Context) {
	result, err := h.dbHealthService.CheckDatabaseHealthAsync()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"sql_connection": "❌ Error: " + err.Error(),
			"xml_status": "N/A",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetDataDelays endpoint - GET /data-delays
func (h *Handlers) GetDataDelays(c *gin.Context) {
	result, err := h.dbHealthService.CheckDataDelaysAsync()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, result)
}