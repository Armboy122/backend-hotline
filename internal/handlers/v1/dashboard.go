package v1

import (
	"backend-hotlines3/internal/dto"
	"backend-hotlines3/internal/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
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
	ctx := c.Request.Context()
	year := c.Query("year")
	month := c.Query("month")
	teamID := c.Query("teamId")
	jobTypeID := c.Query("jobTypeId")

	// Build base query
	query := h.db.WithContext(ctx).Model(&models.TaskDaily{}).
		Scopes(models.TaskNotDeleted).
		Scopes(models.ApplyDashboardFilters(year, month, teamID, jobTypeID))

	// Concurrent queries using errgroup
	g, ctx := errgroup.WithContext(ctx)

	var totalTasks, totalJobTypes, totalFeeders int64

	g.Go(func() error {
		return query.Count(&totalTasks).Error
	})

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.JobType{}).Count(&totalJobTypes).Error
	})

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.Feeder{}).Count(&totalFeeders).Error
	})

	// Top team query (need result for team lookup)
	type TeamCount struct {
		TeamID int64
		Count  int64
	}
	var topTeamResult TeamCount

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.TeamID+" as TeamID, count(*) as count").
			Scopes(models.TaskNotDeleted, models.TaskByYear(year), models.TaskByMonth(month)).
			Group(models.TaskCol.TeamID).
			Order("count DESC").
			Limit(1).
			Find(&topTeamResult).Error
	})

	if err := g.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to load dashboard data",
			},
		})
		return
	}

	// Load team details (sequential after getting topTeamResult)
	var topTeam *dto.TopTeam
	if topTeamResult.TeamID != 0 {
		var team models.Team
		if err := h.db.WithContext(ctx).First(&team, topTeamResult.TeamID).Error; err == nil {
			topTeam = &dto.TopTeam{
				ID:    team.ID,
				Name:  team.Name,
				Count: topTeamResult.Count,
			}
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
	ctx := c.Request.Context()
	year := c.Query("year")
	month := c.Query("month")
	teamID := c.Query("teamId")
	jobTypeID := c.Query("jobTypeId")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Aggregate query
	type JobCount struct {
		JobDetailID int64
		Count       int64
	}
	var results []JobCount

	h.db.WithContext(ctx).Model(&models.TaskDaily{}).
		Select(models.TaskCol.JobDetailID + " as JobDetailID, count(*) as count").
		Scopes(models.TaskNotDeleted).
		Scopes(models.ApplyDashboardFilters(year, month, teamID, jobTypeID)).
		Group(models.TaskCol.JobDetailID).
		Order("count DESC").
		Limit(limit).
		Find(&results)

	if len(results) == 0 {
		c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: []dto.TopJobResponse{}})
		return
	}

	// Batch load job details (eliminates N+1)
	ids := make([]int64, len(results))
	for i, r := range results {
		ids[i] = r.JobDetailID
	}
	var jobDetails []models.JobDetail
	h.db.WithContext(ctx).Preload("JobType").Where("id IN ?", ids).Find(&jobDetails)

	// Build lookup map
	detailMap := make(map[int64]*models.JobDetail, len(jobDetails))
	for i := range jobDetails {
		detailMap[jobDetails[i].ID] = &jobDetails[i]
	}

	// Build response preserving order from aggregate
	response := make([]dto.TopJobResponse, 0, len(results))
	for _, r := range results {
		jd := detailMap[r.JobDetailID]
		if jd == nil {
			continue
		}
		jobTypeName := ""
		if jd.JobType != nil {
			jobTypeName = jd.JobType.Name
		}
		response = append(response, dto.TopJobResponse{
			ID:          jd.ID,
			Name:        jd.Name,
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
	ctx := c.Request.Context()
	year := c.Query("year")
	month := c.Query("month")
	teamID := c.Query("teamId")
	jobTypeID := c.Query("jobTypeId")
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Aggregate query
	type FeederCount struct {
		FeederID int64
		Count    int64
	}
	var results []FeederCount

	h.db.WithContext(ctx).Model(&models.TaskDaily{}).
		Select(models.TaskCol.FeederID+" as FeederID, count(*) as count").
		Scopes(models.TaskNotDeleted, models.TaskFeederNotNull).
		Scopes(models.ApplyDashboardFilters(year, month, teamID, jobTypeID)).
		Group(models.TaskCol.FeederID).
		Order("count DESC").
		Limit(limit).
		Find(&results)

	if len(results) == 0 {
		c.JSON(http.StatusOK, dto.StandardResponse{Success: true, Data: []dto.TopFeederResponse{}})
		return
	}

	// Batch load feeders (eliminates N+1)
	ids := make([]int64, len(results))
	for i, r := range results {
		ids[i] = r.FeederID
	}
	var feeders []models.Feeder
	h.db.WithContext(ctx).Preload("Station").Where("id IN ?", ids).Find(&feeders)

	feederMap := make(map[int64]*models.Feeder, len(feeders))
	for i := range feeders {
		feederMap[feeders[i].ID] = &feeders[i]
	}

	response := make([]dto.TopFeederResponse, 0, len(results))
	for _, r := range results {
		f := feederMap[r.FeederID]
		if f == nil {
			continue
		}
		stationName := ""
		if f.Station != nil {
			stationName = f.Station.Name
		}
		response = append(response, dto.TopFeederResponse{
			ID:          f.ID,
			Code:        f.Code,
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
	ctx := c.Request.Context()
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

	// Concurrent queries using errgroup
	g, ctx := errgroup.WithContext(ctx)

	// Get feeder info
	var feeder models.Feeder
	g.Go(func() error {
		return h.db.WithContext(ctx).Preload("Station").First(&feeder, feederID).Error
	})

	// Aggregate query for job details breakdown
	type JobDetailCount struct {
		JobDetailID int64
		Count       int64
	}
	var results []JobDetailCount

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.JobDetailID+" as JobDetailID, count(*) as count").
			Where(models.TaskCol.FeederID+" = ?", feederID).
			Scopes(models.TaskNotDeleted, models.TaskByYear(year), models.TaskByMonth(month)).
			Group(models.TaskCol.JobDetailID).
			Order("count DESC").
			Find(&results).Error
	})

	if err := g.Wait(); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.StandardResponse{
				Success: false,
				Error: &dto.ErrorInfo{
					Code:    "NOT_FOUND",
					Message: "Feeder not found",
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to load feeder matrix",
			},
		})
		return
	}

	// Batch load job details (eliminates N+1)
	jobDetails := []dto.JobDetailInMatrix{}
	var totalCount int64

	if len(results) > 0 {
		ids := make([]int64, len(results))
		for i, r := range results {
			ids[i] = r.JobDetailID
		}
		var jds []models.JobDetail
		h.db.WithContext(ctx).Preload("JobType").Where("id IN ?", ids).Find(&jds)

		jdMap := make(map[int64]*models.JobDetail, len(jds))
		for i := range jds {
			jdMap[jds[i].ID] = &jds[i]
		}

		for _, r := range results {
			jd := jdMap[r.JobDetailID]
			if jd == nil {
				continue
			}
			jobTypeName := ""
			if jd.JobType != nil {
				jobTypeName = jd.JobType.Name
			}
			jobDetails = append(jobDetails, dto.JobDetailInMatrix{
				ID:          jd.ID,
				Name:        jd.Name,
				Count:       r.Count,
				JobTypeName: jobTypeName,
			})
			totalCount += r.Count
		}
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
	ctx := c.Request.Context()
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	teamID := c.Query("teamId")
	feederID := c.Query("feederId")

	// Build base query with soft delete filter
	baseScope := func(db *gorm.DB) *gorm.DB {
		return db.Scopes(models.TaskNotDeleted).
			Scopes(models.TaskByDateRange(startDate, endDate)).
			Scopes(models.TaskByTeam(teamID)).
			Scopes(models.TaskByFeeder(feederID))
	}

	// Concurrent queries using errgroup for independent aggregates
	g, ctx := errgroup.WithContext(ctx)

	var totalTasks, activeTeams int64

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).Scopes(baseScope).Count(&totalTasks).Error
	})

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Scopes(baseScope).
			Select("COUNT(DISTINCT " + models.TaskCol.TeamID + ")").
			Scan(&activeTeams).Error
	})

	// Top job type
	type JobTypeCount struct {
		JobTypeID int64
		Count     int64
	}
	var topJobTypeResult JobTypeCount

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.JobTypeID + " as JobTypeID, count(*) as count").
			Scopes(baseScope).
			Group(models.TaskCol.JobTypeID).
			Order("count DESC").
			Limit(1).
			Find(&topJobTypeResult).Error
	})

	// Top feeder
	type FeederCount struct {
		FeederID int64
		Count    int64
	}
	var topFeederResult FeederCount

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.FeederID+" as FeederID, count(*) as count").
			Scopes(baseScope, models.TaskFeederNotNull).
			Group(models.TaskCol.FeederID).
			Order("count DESC").
			Limit(1).
			Find(&topFeederResult).Error
	})

	// Tasks by feeder aggregate
	var feederAgg []struct {
		FeederID int64
		Count    int64
	}

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.FeederID+" as FeederID, count(*) as count").
			Scopes(baseScope, models.TaskFeederNotNull).
			Group(models.TaskCol.FeederID).
			Order("count DESC").
			Limit(10).
			Find(&feederAgg).Error
	})

	// Tasks by job type aggregate
	var jtAgg []struct {
		JobTypeID int64
		Count     int64
	}

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.JobTypeID + " as JobTypeID, count(*) as count").
			Scopes(baseScope).
			Group(models.TaskCol.JobTypeID).
			Order("count DESC").
			Find(&jtAgg).Error
	})

	// Tasks by team aggregate
	var teamAgg []struct {
		TeamID int64
		Count  int64
	}

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.TeamID + " as TeamID, count(*) as count").
			Scopes(baseScope).
			Group(models.TaskCol.TeamID).
			Order("count DESC").
			Find(&teamAgg).Error
	})

	// Tasks by date aggregate
	var dateResults []struct {
		Date  string
		Count int64
	}

	g.Go(func() error {
		return h.db.WithContext(ctx).Model(&models.TaskDaily{}).
			Select("TO_CHAR("+models.TaskCol.WorkDate+", 'YYYY-MM-DD') as date, count(*) as count").
			Scopes(models.TaskNotDeleted, models.TaskByDateRange(startDate, endDate)).
			Group("date").
			Order("date ASC").
			Find(&dateResults).Error
	})

	if err := g.Wait(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.StandardResponse{
			Success: false,
			Error: &dto.ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to load dashboard stats",
			},
		})
		return
	}

	// Load job type name (sequential after getting topJobTypeResult)
	topJobType := ""
	if topJobTypeResult.JobTypeID != 0 {
		var jobType models.JobType
		if err := h.db.WithContext(ctx).First(&jobType, topJobTypeResult.JobTypeID).Error; err == nil {
			topJobType = jobType.Name
		}
	}

	// Load feeder code (sequential after getting topFeederResult)
	topFeeder := ""
	if topFeederResult.FeederID != 0 {
		var feeder models.Feeder
		if err := h.db.WithContext(ctx).First(&feeder, topFeederResult.FeederID).Error; err == nil {
			topFeeder = feeder.Code
		}
	}

	// Build tasks by feeder response
	tasksByFeeder := []dto.ChartItem{}
	if len(feederAgg) > 0 {
		feederIDs := make([]int64, len(feederAgg))
		for i, r := range feederAgg {
			feederIDs[i] = r.FeederID
		}
		var feeders []models.Feeder
		h.db.WithContext(ctx).Where("id IN ?", feederIDs).Find(&feeders)

		fMap := make(map[int64]string, len(feeders))
		for _, f := range feeders {
			fMap[f.ID] = f.Code
		}
		for _, r := range feederAgg {
			tasksByFeeder = append(tasksByFeeder, dto.ChartItem{
				Name:  fMap[r.FeederID],
				Value: r.Count,
			})
		}
	}

	// Build tasks by job type response
	tasksByJobType := []dto.ChartItem{}
	if len(jtAgg) > 0 {
		jtIDs := make([]int64, len(jtAgg))
		for i, r := range jtAgg {
			jtIDs[i] = r.JobTypeID
		}
		var jobTypes []models.JobType
		h.db.WithContext(ctx).Where("id IN ?", jtIDs).Find(&jobTypes)

		jtMap := make(map[int64]string, len(jobTypes))
		for _, jt := range jobTypes {
			jtMap[jt.ID] = jt.Name
		}
		for _, r := range jtAgg {
			tasksByJobType = append(tasksByJobType, dto.ChartItem{
				Name:  jtMap[r.JobTypeID],
				Value: r.Count,
			})
		}
	}

	// Build tasks by team response
	tasksByTeam := []dto.ChartItem{}
	if len(teamAgg) > 0 {
		teamIDs := make([]int64, len(teamAgg))
		for i, r := range teamAgg {
			teamIDs[i] = r.TeamID
		}
		var teams []models.Team
		h.db.WithContext(ctx).Where("id IN ?", teamIDs).Find(&teams)

		tMap := make(map[int64]string, len(teams))
		for _, t := range teams {
			tMap[t.ID] = t.Name
		}
		for _, r := range teamAgg {
			tasksByTeam = append(tasksByTeam, dto.ChartItem{
				Name:  tMap[r.TeamID],
				Value: r.Count,
			})
		}
	}

	// Tasks by date response
	tasksByDate := make([]dto.DateChartItem, 0, len(dateResults))
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

	log.Printf("Dashboard stats loaded: %d tasks", totalTasks)
}
