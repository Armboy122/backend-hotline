package repository

import (
	"context"

	"backend-hotlines3/internal/feature/dashboard/entity"
	"backend-hotlines3/internal/models"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Repository interface {
	Summary(ctx context.Context, f entity.SummaryFilter) (entity.SummaryResult, error)
	TopJobs(ctx context.Context, f entity.TopFilter) ([]entity.TopJobResult, error)
	TopFeeders(ctx context.Context, f entity.TopFilter) ([]entity.TopFeederResult, error)
	FeederMatrix(ctx context.Context, f entity.MatrixFilter) (entity.FeederMatrixResult, error)
	Stats(ctx context.Context, f entity.StatsFilter) (entity.StatsResult, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Summary(ctx context.Context, f entity.SummaryFilter) (entity.SummaryResult, error) {
	g, gCtx := errgroup.WithContext(ctx)
	var totalTasks, totalJobTypes, totalFeeders int64
	type tcRow struct{ TeamID, Count int64 }
	var topTeamRow tcRow

	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Scopes(models.TaskNotDeleted, models.ApplyDashboardFilters(f.Year, f.Month, f.TeamID, f.JobTypeID)).
			Count(&totalTasks).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.JobType{}).Count(&totalJobTypes).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.Feeder{}).Count(&totalFeeders).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.TeamID+" as TeamID, count(*) as count").
			Scopes(models.TaskNotDeleted, models.TaskByYear(f.Year), models.TaskByMonth(f.Month)).
			Group(models.TaskCol.TeamID).Order("count DESC").Limit(1).Find(&topTeamRow).Error
	})

	if err := g.Wait(); err != nil {
		return entity.SummaryResult{}, err
	}

	var topTeam *entity.TopTeam
	if topTeamRow.TeamID != 0 {
		var team models.Team
		if err := r.db.WithContext(ctx).First(&team, topTeamRow.TeamID).Error; err == nil {
			topTeam = &entity.TopTeam{ID: team.ID, Name: team.Name, Count: topTeamRow.Count}
		}
	}

	return entity.SummaryResult{
		TotalTasks: totalTasks, TotalJobTypes: totalJobTypes,
		TotalFeeders: totalFeeders, TopTeam: topTeam,
	}, nil
}

func (r *repository) TopJobs(ctx context.Context, f entity.TopFilter) ([]entity.TopJobResult, error) {
	type row struct{ JobDetailID, Count int64 }
	var agg []row
	r.db.WithContext(ctx).Model(&models.TaskDaily{}).
		Select(models.TaskCol.JobDetailID+" as JobDetailID, count(*) as count").
		Scopes(models.TaskNotDeleted, models.ApplyDashboardFilters(f.Year, f.Month, f.TeamID, f.JobTypeID)).
		Group(models.TaskCol.JobDetailID).Order("count DESC").Limit(f.Limit).Find(&agg)

	if len(agg) == 0 {
		return []entity.TopJobResult{}, nil
	}

	ids := make([]int64, len(agg))
	for i, a := range agg {
		ids[i] = a.JobDetailID
	}
	var jds []models.JobDetail
	r.db.WithContext(ctx).Preload("JobType").Where("id IN ?", ids).Find(&jds)
	jdMap := make(map[int64]*models.JobDetail, len(jds))
	for i := range jds {
		jdMap[jds[i].ID] = &jds[i]
	}

	out := make([]entity.TopJobResult, 0, len(agg))
	for _, a := range agg {
		jd := jdMap[a.JobDetailID]
		if jd == nil {
			continue
		}
		jtName := ""
		if jd.JobType != nil {
			jtName = jd.JobType.Name
		}
		out = append(out, entity.TopJobResult{ID: jd.ID, Name: jd.Name, Count: a.Count, JobTypeName: jtName})
	}
	return out, nil
}

func (r *repository) TopFeeders(ctx context.Context, f entity.TopFilter) ([]entity.TopFeederResult, error) {
	type row struct{ FeederID, Count int64 }
	var agg []row
	r.db.WithContext(ctx).Model(&models.TaskDaily{}).
		Select(models.TaskCol.FeederID+" as FeederID, count(*) as count").
		Scopes(models.TaskNotDeleted, models.TaskFeederNotNull, models.ApplyDashboardFilters(f.Year, f.Month, f.TeamID, f.JobTypeID)).
		Group(models.TaskCol.FeederID).Order("count DESC").Limit(f.Limit).Find(&agg)

	if len(agg) == 0 {
		return []entity.TopFeederResult{}, nil
	}

	ids := make([]int64, len(agg))
	for i, a := range agg {
		ids[i] = a.FeederID
	}
	var feeders []models.Feeder
	r.db.WithContext(ctx).Preload("Station").Where("id IN ?", ids).Find(&feeders)
	fMap := make(map[int64]*models.Feeder, len(feeders))
	for i := range feeders {
		fMap[feeders[i].ID] = &feeders[i]
	}

	out := make([]entity.TopFeederResult, 0, len(agg))
	for _, a := range agg {
		fd := fMap[a.FeederID]
		if fd == nil {
			continue
		}
		stName := ""
		if fd.Station != nil {
			stName = fd.Station.Name
		}
		out = append(out, entity.TopFeederResult{ID: fd.ID, Code: fd.Code, StationName: stName, Count: a.Count})
	}
	return out, nil
}

func (r *repository) FeederMatrix(ctx context.Context, f entity.MatrixFilter) (entity.FeederMatrixResult, error) {
	g, gCtx := errgroup.WithContext(ctx)
	var feeder models.Feeder
	g.Go(func() error {
		return r.db.WithContext(gCtx).Preload("Station").First(&feeder, f.FeederID).Error
	})

	type jdRow struct{ JobDetailID, Count int64 }
	var agg []jdRow
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.JobDetailID+" as JobDetailID, count(*) as count").
			Where(models.TaskCol.FeederID+" = ?", f.FeederID).
			Scopes(models.TaskNotDeleted, models.TaskByYear(f.Year), models.TaskByMonth(f.Month)).
			Group(models.TaskCol.JobDetailID).Order("count DESC").Find(&agg).Error
	})

	if err := g.Wait(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return entity.FeederMatrixResult{}, entity.ErrFeederNotFound
		}
		return entity.FeederMatrixResult{}, err
	}

	var jobDetails []entity.JobDetailInMatrix
	var totalCount int64

	if len(agg) > 0 {
		ids := make([]int64, len(agg))
		for i, a := range agg {
			ids[i] = a.JobDetailID
		}
		var jds []models.JobDetail
		r.db.WithContext(ctx).Preload("JobType").Where("id IN ?", ids).Find(&jds)
		jdMap := make(map[int64]*models.JobDetail, len(jds))
		for i := range jds {
			jdMap[jds[i].ID] = &jds[i]
		}
		for _, a := range agg {
			jd := jdMap[a.JobDetailID]
			if jd == nil {
				continue
			}
			jtName := ""
			if jd.JobType != nil {
				jtName = jd.JobType.Name
			}
			jobDetails = append(jobDetails, entity.JobDetailInMatrix{ID: jd.ID, Name: jd.Name, Count: a.Count, JobTypeName: jtName})
			totalCount += a.Count
		}
	}

	stName := ""
	if feeder.Station != nil {
		stName = feeder.Station.Name
	}

	return entity.FeederMatrixResult{
		FeederID: feeder.ID, FeederCode: feeder.Code,
		StationName: stName, TotalCount: totalCount, JobDetails: jobDetails,
	}, nil
}

func (r *repository) Stats(ctx context.Context, f entity.StatsFilter) (entity.StatsResult, error) {
	baseScope := func(db *gorm.DB) *gorm.DB {
		return db.Scopes(models.TaskNotDeleted, models.TaskByDateRange(f.StartDate, f.EndDate), models.TaskByTeam(f.TeamID), models.TaskByFeeder(f.FeederID))
	}

	g, gCtx := errgroup.WithContext(ctx)
	var totalTasks, activeTeams int64
	type jtRow struct{ JobTypeID, Count int64 }
	type fcRow struct{ FeederID, Count int64 }
	type tcRow struct{ TeamID, Count int64 }
	type dtRow struct {
		Date  string
		Count int64
	}
	var topJT jtRow
	var topF fcRow
	var feederAgg []fcRow
	var jtAgg []jtRow
	var teamAgg []tcRow
	var dateAgg []dtRow

	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).Scopes(baseScope).Count(&totalTasks).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).Scopes(baseScope).
			Select("COUNT(DISTINCT " + models.TaskCol.TeamID + ")").Scan(&activeTeams).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.JobTypeID+" as JobTypeID, count(*) as count").
			Scopes(baseScope).Group(models.TaskCol.JobTypeID).Order("count DESC").Limit(1).Find(&topJT).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.FeederID+" as FeederID, count(*) as count").
			Scopes(baseScope, models.TaskFeederNotNull).Group(models.TaskCol.FeederID).Order("count DESC").Limit(1).Find(&topF).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.FeederID+" as FeederID, count(*) as count").
			Scopes(baseScope, models.TaskFeederNotNull).Group(models.TaskCol.FeederID).Order("count DESC").Limit(10).Find(&feederAgg).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.JobTypeID+" as JobTypeID, count(*) as count").
			Scopes(baseScope).Group(models.TaskCol.JobTypeID).Order("count DESC").Find(&jtAgg).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select(models.TaskCol.TeamID+" as TeamID, count(*) as count").
			Scopes(baseScope).Group(models.TaskCol.TeamID).Order("count DESC").Find(&teamAgg).Error
	})
	g.Go(func() error {
		return r.db.WithContext(gCtx).Model(&models.TaskDaily{}).
			Select("TO_CHAR("+models.TaskCol.WorkDate+", 'YYYY-MM-DD') as date, count(*) as count").
			Scopes(models.TaskNotDeleted, models.TaskByDateRange(f.StartDate, f.EndDate)).
			Group("date").Order("date ASC").Find(&dateAgg).Error
	})

	if err := g.Wait(); err != nil {
		return entity.StatsResult{}, err
	}

	topJTName := ""
	if topJT.JobTypeID != 0 {
		var jt models.JobType
		if err := r.db.WithContext(ctx).First(&jt, topJT.JobTypeID).Error; err == nil {
			topJTName = jt.Name
		}
	}
	topFCode := ""
	if topF.FeederID != 0 {
		var fd models.Feeder
		if err := r.db.WithContext(ctx).First(&fd, topF.FeederID).Error; err == nil {
			topFCode = fd.Code
		}
	}

	// Build charts
	tasksByFeeder := []entity.ChartItem{}
	if len(feederAgg) > 0 {
		ids := make([]int64, len(feederAgg))
		for i, a := range feederAgg {
			ids[i] = a.FeederID
		}
		var rows []models.Feeder
		r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows)
		m := make(map[int64]string, len(rows))
		for _, row := range rows {
			m[row.ID] = row.Code
		}
		for _, a := range feederAgg {
			tasksByFeeder = append(tasksByFeeder, entity.ChartItem{Name: m[a.FeederID], Value: a.Count})
		}
	}

	tasksByJobType := []entity.ChartItem{}
	if len(jtAgg) > 0 {
		ids := make([]int64, len(jtAgg))
		for i, a := range jtAgg {
			ids[i] = a.JobTypeID
		}
		var rows []models.JobType
		r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows)
		m := make(map[int64]string, len(rows))
		for _, row := range rows {
			m[row.ID] = row.Name
		}
		for _, a := range jtAgg {
			tasksByJobType = append(tasksByJobType, entity.ChartItem{Name: m[a.JobTypeID], Value: a.Count})
		}
	}

	tasksByTeam := []entity.ChartItem{}
	if len(teamAgg) > 0 {
		ids := make([]int64, len(teamAgg))
		for i, a := range teamAgg {
			ids[i] = a.TeamID
		}
		var rows []models.Team
		r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows)
		m := make(map[int64]string, len(rows))
		for _, row := range rows {
			m[row.ID] = row.Name
		}
		for _, a := range teamAgg {
			tasksByTeam = append(tasksByTeam, entity.ChartItem{Name: m[a.TeamID], Value: a.Count})
		}
	}

	tasksByDate := make([]entity.DateChartItem, 0, len(dateAgg))
	for _, d := range dateAgg {
		tasksByDate = append(tasksByDate, entity.DateChartItem{Date: d.Date, Count: d.Count})
	}

	return entity.StatsResult{
		Summary: entity.StatsSummary{
			TotalTasks: totalTasks, ActiveTeams: activeTeams,
			TopJobType: topJTName, TopFeeder: topFCode,
		},
		Charts: entity.StatsCharts{
			TasksByFeeder:  tasksByFeeder,
			TasksByJobType: tasksByJobType,
			TasksByTeam:    tasksByTeam,
			TasksByDate:    tasksByDate,
		},
	}, nil
}
