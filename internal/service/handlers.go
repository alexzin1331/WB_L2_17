package service

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"WB_L2_17/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server and contains business logic handlers
type Server struct {
	calendarService *storage.Manager // storage manager for calendar events
}

var (
	logFile *os.File // file handle for request logging
)

// initRequestLogger initializes the log file for HTTP request logging
func initRequestLogger() {
	var err error
	logFile, err = os.OpenFile("http_requests.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to initialize request logger: %v", err)
	}
}

// NewRouter creates and configures a new Gin router with all endpoints
func NewRouter(calendarService *storage.Manager) *gin.Engine {
	server := &Server{calendarService: calendarService}
	router := gin.Default()

	initRequestLogger()

	// register middleware
	router.Use(requestLoggerMiddleware())
	router.Use(RequestLoggerToFileMiddleware())

	// API documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// event management endpoints
	router.POST("/create_event", server.handleCreateEvent)
	router.POST("/update_event", server.handleUpdateEvent)
	router.POST("/delete_event", server.handleDeleteEvent)

	// event query endpoints
	router.GET("/events_for_day", server.handleGetDayEvents)
	router.GET("/events_for_week", server.handleGetWeekEvents)
	router.GET("/events_for_month", server.handleGetMonthEvents)

	// metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}

// EventRequest defines the JSON structure for event creation/update
type EventRequest struct {
	UserID int    `json:"user_id" binding:"required"` // unique user identifier
	Date   string `json:"date" binding:"required"`    // event date in YYYY-MM-DD format
	Title  string `json:"title" binding:"required"`   // brief event title
	Text   string `json:"event" binding:"required"`   // detailed event description
}

// DeleteEventRequest defines the JSON structure for event deletion
type DeleteEventRequest struct {
	UserID int    `json:"user_id" binding:"required"` // unique user identifier
	Date   string `json:"date" binding:"required"`    // event date in YYYY-MM-DD format
}

// @Summary Create new event
// @Description Creates a new calendar event for specified user
// @Tags events
// @Accept json
// @Produce json
// @Param event body EventRequest true "Event details"
// @Success 200 {object} map[string]string{"result": "event created"}
// @Failure 400 {object} map[string]string{"error": "error description"}
// @Router /create_event [post]
func (s *Server) handleCreateEvent(c *gin.Context) {
	var request EventRequest

	// bind and validate JSON input
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// parse and validate date format
	eventDate, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format: " + err.Error()})
		return
	}

	// create event in storage
	event := storage.Event{
		UserID: request.UserID,
		Date:   eventDate,
		Title:  request.Title,
		Text:   request.Text,
	}

	if err := s.calendarService.CreateEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "event created"})
}

// @Summary Update existing event
// @Description Updates an existing calendar event
// @Tags events
// @Accept json
// @Produce json
// @Param event body EventRequest true "Updated event details"
// @Success 200 {object} map[string]string{"result": "event updated"}
// @Failure 400 {object} map[string]string{"error": "error description"}
// @Failure 503 {object} map[string]string{"error": "error description"}
// @Router /update_event [post]
func (s *Server) handleUpdateEvent(c *gin.Context) {
	var request EventRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	eventDate, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format: " + err.Error()})
		return
	}

	err = s.calendarService.UpdateEvent(request.UserID, eventDate, request.Title, request.Text)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "event updated"})
}

// @Summary Delete event
// @Description Deletes an existing calendar event
// @Tags events
// @Accept json
// @Produce json
// @Param event body DeleteEventRequest true "Event deletion parameters"
// @Success 200 {object} map[string]string{"result": "event deleted"}
// @Failure 400 {object} map[string]string{"error": "error description"}
// @Failure 503 {object} map[string]string{"error": "error description"}
// @Router /delete_event [post]
func (s *Server) handleDeleteEvent(c *gin.Context) {
	var request DeleteEventRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	eventDate, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format: " + err.Error()})
		return
	}

	err = s.calendarService.DeleteEvent(request.UserID, eventDate)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "event deleted"})
}

// @Summary Get daily events
// @Description Returns all events for specified user on given day
// @Tags events
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Param date query string true "Date in YYYY-MM-DD format"
// @Success 200 {object} map[string][]storage.Event{"result": "list of events"}
// @Failure 400 {object} map[string]string{"error": "error description"}
// @Router /events_for_day [get]
func (s *Server) handleGetDayEvents(c *gin.Context) {
	s.handleGetEvents(c, 0) // 0 days means single day
}

// @Summary Get weekly events
// @Description Returns all events for specified user in 7-day period starting from given date
// @Tags events
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Param date query string true "Start date in YYYY-MM-DD format"
// @Success 200 {object} map[string][]storage.Event{"result": "list of events"}
// @Failure 400 {object} map[string]string{"error": "error description"}
// @Router /events_for_week [get]
func (s *Server) handleGetWeekEvents(c *gin.Context) {
	s.handleGetEvents(c, 7) // 7 days = 1 week
}

// @Summary Get monthly events
// @Description Returns all events for specified user in 30-day period starting from given date
// @Tags events
// @Accept json
// @Produce json
// @Param user_id query int true "User ID"
// @Param date query string true "Start date in YYYY-MM-DD format"
// @Success 200 {object} map[string][]storage.Event{"result": "list of events"}
// @Failure 400 {object} map[string]string{"error": "error description"}
// @Router /events_for_month [get]
func (s *Server) handleGetMonthEvents(c *gin.Context) {
	s.handleGetEvents(c, 30)
}

// handleGetEvents is a helper method for event query handlers
func (s *Server) handleGetEvents(c *gin.Context, days int) {
	// extract and validate user ID
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID: " + err.Error()})
		return
	}

	// extract and validate date
	startDate, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format: " + err.Error()})
		return
	}

	// calculate end date based on days parameter
	endDate := startDate
	if days > 0 {
		endDate = startDate.AddDate(0, 0, days)
	}

	// retrieve events from storage
	events := s.calendarService.GetEventsForRange(userID, startDate, endDate)
	c.JSON(http.StatusOK, gin.H{"result": events})
}

// requestLoggerMiddleware logs HTTP requests to stdout
func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// log request start
		log.Printf("started %s %s", method, path)

		// process request
		c.Next()

		// log request completion
		duration := time.Since(startTime)
		status := c.Writer.Status()
		log.Printf("completed %s %s - status %d - duration %v",
			method, path, status, duration)
	}
}
