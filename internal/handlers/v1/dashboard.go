package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// Summary - GET /v1/dashboard/summary
func (h *DashboardHandler) Summary(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")
	teamID := c.Query("teamId")
	jobTypeID := c.Query("jobTypeId")

	// Build base query
	query := h.db.Model(&models.TaskDaily{}).Where("deleted_at IS NULL")

	if year != "" {
		query = query.Where("EXTRACT(YEAR FROM work_date) = ?", year)
	}
	if month != "" {
		query = query.Where("EXTRACT(MONTH FROM work_date) = ?", month)
	}
	if teamID != "" && teamID != "all" {
		query = query.Where("team_id = ?", teamID)
	}
	if jobTypeID != "" && jobTypeID != "all" {
		query = query.Where("job_type_id = ?", jobTypeID)
	}

	// Total tasks
	var totalTasks int64
	query.Count(&totalTasks)

	// Total job types used
	var totalJobTypes int64
	h.db.Model(&models.JobType{}).Count(&totalJobTypes)

	// Total feeders used
	var totalFeeders int64
	h.db.Model(&models.Feeder{}).Count(&totalFeeders)

	// Top team
	type TeamCount struct {
		TeamID int64
		Count  int64
	}
	var topTeamResult TeamCount
	topTeamQuery := h.db.Model(&models.TaskDaily{}).
		Select("team_id, count(*) as count").
		Where("deleted_at IS NULL")

	if year != "" {
		topTeamQuery = topTeamQuery.Where("EXTRACT(YEAR FROM work_date) = ?", year)
	}
	if month != "" {
		topTeamQuery = topTeamQuery.Where("EXTRACT(MONTH FROM work_date) = ?", month)
	}

	topTeamQuery.Group("team_id").
		Order("count DESC").
		Limit(1).
		Find(&topTeamResult)

	var topTeam *dto.TopTeam
	if topTeamResult.TeamID != 0 {
		var team models.Team
		h.db.First(&team, topTeamResult.TeamID)
		topTeam = &dto.TopTeam{
			ID:    team.ID,
			Name:  team.Name,
			Count: topTeamResult.Count,
		}
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.DashboardSummaryResponse{
			TotalTasks:    totalTasks,
			TotalJobTypes: totalJobTypes,
			TotalFeeders:  totalFeeders,
			TopTeam:       topTeam,
		},
	})
}

// TopJobs - GET /v1/dashboard/top-jobs
func (h *DashboardHandler) TopJobs(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")
	teamID := c.Query("teamId")
	jobTypeID := c.Query("jobTypeId")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build query
	type JobCount struct {
		JobDetailID int64
		Count       int64
	}
	var results []JobCount

	query := h.db.Model(&models.TaskDaily{}).
		Select("job_detail_id, count(*) as count").
		Where("deleted_at IS NULL")

	if year != "" {
		query = query.Where("EXTRACT(YEAR FROM work_date) = ?", year)
	}
	if month != "" {
		query = query.Where("EXTRACT(MONTH FROM work_date) = ?", month)
	}
	if teamID != "" && teamID != "all" {
		query = query.Where("team_id = ?", teamID)
	}
	if jobTypeID != "" && jobTypeID != "all" {
		query = query.Where("job_type_id = ?", jobTypeID)
	}

	query.Group("job_detail_id").
		Order("count DESC").
		Limit(limit).
		Find(&results)

	// Get job details with job types
	var response []dto.TopJobResponse
	for _, r := range results {
		var jobDetail models.JobDetail
		h.db.Preload("JobType").First(&jobDetail, r.JobDetailID)

		jobTypeName := ""
		if jobDetail.JobType != nil {
			jobTypeName = jobDetail.JobType.Name
		}

		response = append(response, dto.TopJobResponse{
			ID:          jobDetail.ID,
			Name:        jobDetail.Name,
			Count:       r.Count,
			JobTypeName: jobTypeName,
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// TopFeeders - GET /v1/dashboard/top-feeders
func (h *DashboardHandler) TopFeeders(c *gin.Context) {
	year := c.Query("year")
	month := c.Query("month")
	teamID := c.Query("teamId")
	jobTypeID := c.Query("jobTypeId")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build query
	type FeederCount struct {
		FeederID int64
		Count    int64
	}
	var results []FeederCount

	query := h.db.Model(&models.TaskDaily{}).
		Select("feeder_id, count(*) as count").
		Where("deleted_at IS NULL").
		Where("feeder_id IS NOT NULL")

	if year != "" {
		query = query.Where("EXTRACT(YEAR FROM work_date) = ?", year)
	}
	if month != "" {
		query = query.Where("EXTRACT(MONTH FROM work_date) = ?", month)
	}
	if teamID != "" && teamID != "all" {
		query = query.Where("team_id = ?", teamID)
	}
	if jobTypeID != "" && jobTypeID != "all" {
		query = query.Where("job_type_id = ?", jobTypeID)
	}

	query.Group("feeder_id").
		Order("count DESC").
		Limit(limit).
		Find(&results)

	// Get feeder details with stations
	var response []dto.TopFeederResponse
	for _, r := range results {
		var feeder models.Feeder
		h.db.Preload("Station").First(&feeder, r.FeederID)

		stationName := ""
		if feeder.Station != nil {
			stationName = feeder.Station.Name
		}

		response = append(response, dto.TopFeederResponse{
			ID:          feeder.ID,
			Code:        feeder.Code,
			StationName: stationName,
			Count:       r.Count,
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// FeederMatrix - GET /v1/dashboard/feeder-matrix
func (h *DashboardHandler) FeederMatrix(c *gin.Context) {
	feederIDStr := c.Query("feederId")
	if feederIDStr == "" {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Message: "feederId is required",
			},
		})
		return
	}

	feederID, err := strconv.ParseInt(feederIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INVALID_ID",
				Message: "Invalid feeder ID",
			},
		})
		return
	}

	year := c.Query("year")
	month := c.Query("month")

	// Get feeder info
	var feeder models.Feeder
	if err := h.db.Preload("Station").First(&feeder, feederID).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "NOT_FOUND",
				Message: "Feeder not found",
			},
		})
		return
	}

	// Build query for job details breakdown
	type JobDetailCount struct {
		JobDetailID int64
		Count       int64
	}
	var results []JobDetailCount

	query := h.db.Model(&models.TaskDaily{}).
		Select("job_detail_id, count(*) as count").
		Where("deleted_at IS NULL").
		Where("feeder_id = ?", feederID)

	if year != "" {
		query = query.Where("EXTRACT(YEAR FROM work_date) = ?", year)
	}
	if month != "" {
		query = query.Where("EXTRACT(MONTH FROM work_date) = ?", month)
	}

	query.Group("job_detail_id").
		Order("count DESC").
		Find(&results)

	// Build job details response
	var jobDetails []dto.JobDetailInMatrix
	var totalCount int64
	for _, r := range results {
		var jobDetail models.JobDetail
		h.db.Preload("JobType").First(&jobDetail, r.JobDetailID)

		jobTypeName := ""
		if jobDetail.JobType != nil {
			jobTypeName = jobDetail.JobType.Name
		}

		jobDetails = append(jobDetails, dto.JobDetailInMatrix{
			ID:          jobDetail.ID,
			Name:        jobDetail.Name,
			Count:       r.Count,
			JobTypeName: jobTypeName,
		})
		totalCount += r.Count
	}

	stationName := ""
	if feeder.Station != nil {
		stationName = feeder.Station.Name
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.FeederMatrixResponse{
			FeederID:    feeder.ID,
			FeederCode:  feeder.Code,
			StationName: stationName,
			TotalCount:  totalCount,
			JobDetails:  jobDetails,
		},
	})
}

// Stats - GET /v1/dashboard/stats
func (h *DashboardHandler) Stats(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	teamID := c.Query("teamId")
	feederID := c.Query("feederId")

	// Build base query
	query := h.db.Model(&models.TaskDaily{}).Where("deleted_at IS NULL")

	if startDate != "" {
		query = query.Where("work_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("work_date <= ?", endDate)
	}
	if teamID != "" && teamID != "all" {
		query = query.Where("team_id = ?", teamID)
	}
	if feederID != "" && feederID != "all" {
		query = query.Where("feeder_id = ?", feederID)
	}

	// Total tasks
	var totalTasks int64
	query.Count(&totalTasks)

	// Active teams
	var activeTeams int64
	h.db.Model(&models.TaskDaily{}).
		Select("DISTINCT team_id").
		Where("deleted_at IS NULL").
		Count(&activeTeams)

	// Top job type
	type JobTypeCount struct {
		JobTypeID int64
		Count     int64
	}
	var topJobTypeResult JobTypeCount
	h.db.Model(&models.TaskDaily{}).
		Select("job_type_id, count(*) as count").
		Where("deleted_at IS NULL").
		Group("job_type_id").
		Order("count DESC").
		Limit(1).
		Find(&topJobTypeResult)

	topJobType := ""
	if topJobTypeResult.JobTypeID != 0 {
		var jobType models.JobType
		h.db.First(&jobType, topJobTypeResult.JobTypeID)
		topJobType = jobType.Name
	}

	// Top feeder
	type FeederCount struct {
		FeederID int64
		Count    int64
	}
	var topFeederResult FeederCount
	h.db.Model(&models.TaskDaily{}).
		Select("feeder_id, count(*) as count").
		Where("deleted_at IS NULL").
		Where("feeder_id IS NOT NULL").
		Group("feeder_id").
		Order("count DESC").
		Limit(1).
		Find(&topFeederResult)

	topFeeder := ""
	if topFeederResult.FeederID != 0 {
		var feeder models.Feeder
		h.db.First(&feeder, topFeederResult.FeederID)
		topFeeder = feeder.Code
	}

	// Charts data

	// Tasks by feeder
	type ChartData struct {
		Name  string
		Value int64
	}
	var tasksByFeeder []dto.ChartItem
	var feederResults []struct {
		FeederID int64
		Count    int64
	}
	h.db.Model(&models.TaskDaily{}).
		Select("feeder_id, count(*) as count").
		Where("deleted_at IS NULL").
		Where("feeder_id IS NOT NULL").
		Group("feeder_id").
		Order("count DESC").
		Limit(10).
		Find(&feederResults)

	for _, r := range feederResults {
		var feeder models.Feeder
		h.db.First(&feeder, r.FeederID)
		tasksByFeeder = append(tasksByFeeder, dto.ChartItem{
			Name:  feeder.Code,
			Value: r.Count,
		})
	}

	// Tasks by job type
	var tasksByJobType []dto.ChartItem
	var jobTypeResults []struct {
		JobTypeID int64
		Count     int64
	}
	h.db.Model(&models.TaskDaily{}).
		Select("job_type_id, count(*) as count").
		Where("deleted_at IS NULL").
		Group("job_type_id").
		Order("count DESC").
		Find(&jobTypeResults)

	for _, r := range jobTypeResults {
		var jobType models.JobType
		h.db.First(&jobType, r.JobTypeID)
		tasksByJobType = append(tasksByJobType, dto.ChartItem{
			Name:  jobType.Name,
			Value: r.Count,
		})
	}

	// Tasks by team
	var tasksByTeam []dto.ChartItem
	var teamResults []struct {
		TeamID int64
		Count  int64
	}
	h.db.Model(&models.TaskDaily{}).
		Select("team_id, count(*) as count").
		Where("deleted_at IS NULL").
		Group("team_id").
		Order("count DESC").
		Find(&teamResults)

	for _, r := range teamResults {
		var team models.Team
		h.db.First(&team, r.TeamID)
		tasksByTeam = append(tasksByTeam, dto.ChartItem{
			Name:  team.Name,
			Value: r.Count,
		})
	}

	// Tasks by date
	var tasksByDate []dto.DateChartItem
	var dateResults []struct {
		Date  string
		Count int64
	}
	dateQuery := h.db.Model(&models.TaskDaily{}).
		Select("TO_CHAR(work_date, 'YYYY-MM-DD') as date, count(*) as count").
		Where("deleted_at IS NULL")

	if startDate != "" {
		dateQuery = dateQuery.Where("work_date >= ?", startDate)
	}
	if endDate != "" {
		dateQuery = dateQuery.Where("work_date <= ?", endDate)
	}

	dateQuery.Group("work_date").
		Order("work_date ASC").
		Find(&dateResults)

	for _, r := range dateResults {
		tasksByDate = append(tasksByDate, dto.DateChartItem{
			Date:  r.Date,
			Count: r.Count,
		})
	}

	c.JSON(http.StatusOK, dto.StandardResponse{
		Success: true,
		Data: dto.DashboardStatsResponse{
			Summary: dto.DashboardStatsSummary{
				TotalTasks:  totalTasks,
				ActiveTeams: activeTeams,
				TopJobType:  topJobType,
				TopFeeder:   topFeeder,
			},
			Charts: dto.DashboardCharts{
				TasksByFeeder:  tasksByFeeder,
				TasksByJobType: tasksByJobType,
				TasksByTeam:    tasksByTeam,
				TasksByDate:    tasksByDate,
			},
		},
	})
}
