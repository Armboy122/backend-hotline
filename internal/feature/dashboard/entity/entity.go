package entity

import "errors"

var ErrFeederNotFound = errors.New("feeder not found")

type SummaryFilter struct {
	Year      string
	Month     string
	TeamID    string
	JobTypeID string
}

type TopFilter struct {
	Year      string
	Month     string
	TeamID    string
	JobTypeID string
	Limit     int
}

type MatrixFilter struct {
	FeederID int64
	Year     string
	Month    string
}

type StatsFilter struct {
	StartDate string
	EndDate   string
	TeamID    string
	FeederID  string
}

type TopTeam struct {
	ID    int64
	Name  string
	Count int64
}

type SummaryResult struct {
	TotalTasks    int64
	TotalJobTypes int64
	TotalFeeders  int64
	TopTeam       *TopTeam
}

type TopJobResult struct {
	ID          int64
	Name        string
	Count       int64
	JobTypeName string
}

type TopFeederResult struct {
	ID          int64
	Code        string
	StationName string
	Count       int64
}

type JobDetailInMatrix struct {
	ID          int64
	Name        string
	Count       int64
	JobTypeName string
}

type FeederMatrixResult struct {
	FeederID    int64
	FeederCode  string
	StationName string
	TotalCount  int64
	JobDetails  []JobDetailInMatrix
}

type ChartItem struct {
	Name  string
	Value int64
}

type DateChartItem struct {
	Date  string
	Count int64
}

type StatsSummary struct {
	TotalTasks  int64
	ActiveTeams int64
	TopJobType  string
	TopFeeder   string
}

type StatsCharts struct {
	TasksByFeeder  []ChartItem
	TasksByJobType []ChartItem
	TasksByTeam    []ChartItem
	TasksByDate    []DateChartItem
}

type StatsResult struct {
	Summary StatsSummary
	Charts  StatsCharts
}
