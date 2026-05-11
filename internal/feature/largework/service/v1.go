package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend-hotlines3/internal/feature/largework/entity"
	repo "backend-hotlines3/internal/feature/largework/repository"
)

var (
	ErrForbidden               = errors.New("forbidden")
	ErrNotFound                = errors.New("large work item not found")
	ErrInvalidID               = errors.New("invalid large work item ID")
	ErrInvalidParticipantTeams = errors.New("participant team list must contain at least two distinct teams including owner")
	ErrInvalidStartDate        = errors.New("startDate is required")
	ErrInvalidTitle            = errors.New("title is required")
	ErrInvalidLocation         = errors.New("locationText is required")
	ErrInvalidOwnerTeam        = errors.New("ownerTeamId is required")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidDateRange        = errors.New("endDate must be on or after startDate")
	ErrInvalidStateTransition  = errors.New("large work item state does not allow this operation")
	ErrInvalidTask             = errors.New("invalid large work task")
	ErrPhotoRequired           = errors.New("before and after photo data is required")
)

type CreateInput struct {
	OwnerTeamID        int64
	ParticipantTeamIDs []int64
	Title              string
	WorkType           *string
	StartDate          time.Time
	EndDate            *time.Time
	WorkTime           *string
	LocationText       string
	PEAID              *int64
	OperationCenterID  *int64
	FeederID           *int64
	StationID          *int64
	Notes              *string
	Status             string
}

type UpdateInput struct {
	ID                 int64
	OwnerTeamID        *int64
	ParticipantTeamIDs []int64
	Title              *string
	WorkType           *string
	StartDate          *time.Time
	EndDate            *time.Time
	WorkTime           *string
	LocationText       *string
	PEAID              *int64
	OperationCenterID  *int64
	FeederID           *int64
	StationID          *int64
	Notes              *string
	Status             *string
}

type ListInput struct {
	Page     int
	Limit    int
	From     *time.Time
	To       *time.Time
	TeamID   *int64
	Statuses []string
}

type ListOutput struct {
	Items []entity.LargeWorkItem
	Page  int
	Limit int
	Total int64
}

type TaskPointInput struct {
	AssignedTeamID int64
	Sequence       int
	PointLabel     string
	Latitude       *float64
	Longitude      *float64
	WorkType       string
	WorkDetail     *string
	PointCount     *int
	TreeCount      *int
	ItemCount      *int
	Notes          *string
	Metadata       map[string]any
}

type ReplaceTasksInput struct{ Tasks []TaskPointInput }

type MyTodosInput struct {
	Page     int
	Limit    int
	Statuses []string
}
type MyTodosOutput struct {
	Items []entity.LargeWorkTask
	Page  int
	Limit int
	Total int64
}

type ProgressSummary struct {
	Total      int
	Todo       int
	InProgress int
	Done       int
	Blocked    int
	Cancelled  int
}
type TeamProgressSummary struct {
	AssignedTeamID int64
	Total          int
	Todo           int
	InProgress     int
	Done           int
	Blocked        int
	Cancelled      int
}
type OverviewOutput struct {
	Plan         entity.LargeWorkItem
	Progress     ProgressSummary
	TeamProgress []TeamProgressSummary
}

const (
	PhotoKindBefore = "before"
	PhotoKindAfter  = "after"
)

type CompleteTaskInput struct {
	AfterPhotoURLs []string
	CompletionNote string
}

type Service interface {
	Create(context.Context, entity.Actor, CreateInput) (*entity.LargeWorkItem, error)
	GetByID(context.Context, entity.Actor, int64) (*entity.LargeWorkItem, error)
	List(context.Context, entity.Actor, ListInput) (*ListOutput, error)
	Update(context.Context, entity.Actor, UpdateInput) (*entity.LargeWorkItem, error)
	Cancel(context.Context, entity.Actor, int64) (*entity.LargeWorkItem, error)
	GetOverview(context.Context, entity.Actor, int64) (*OverviewOutput, error)
	ReplaceTasks(context.Context, entity.Actor, int64, ReplaceTasksInput) ([]entity.LargeWorkTask, error)
	ListTasks(context.Context, entity.Actor, int64) ([]entity.LargeWorkTask, error)
	MyTodos(context.Context, entity.Actor, MyTodosInput) (*MyTodosOutput, error)
	StartTask(context.Context, entity.Actor, int64) (*entity.LargeWorkTask, error)
	AttachTaskPhoto(context.Context, entity.Actor, int64, string, string) (*entity.LargeWorkTask, error)
	CompleteTask(context.Context, entity.Actor, int64, CompleteTaskInput) (*entity.LargeWorkTask, error)
	BlockTask(context.Context, entity.Actor, int64, string) (*entity.LargeWorkTask, error)
}

type service struct {
	repo repo.Repository
}

func NewService(repository repo.Repository) Service {
	return &service{repo: repository}
}

func (s *service) Create(ctx context.Context, actor entity.Actor, input CreateInput) (*entity.LargeWorkItem, error) {
	if input.OwnerTeamID <= 0 {
		return nil, ErrInvalidOwnerTeam
	}
	if input.Title == "" {
		return nil, ErrInvalidTitle
	}
	if input.LocationText == "" {
		return nil, ErrInvalidLocation
	}
	if input.StartDate.IsZero() {
		return nil, ErrInvalidStartDate
	}
	if input.EndDate != nil && input.EndDate.Before(input.StartDate) {
		return nil, ErrInvalidDateRange
	}
	participantIDs := normalizeTeamIDs(append(input.ParticipantTeamIDs, input.OwnerTeamID))
	if len(participantIDs) < 2 {
		return nil, ErrInvalidParticipantTeams
	}
	if !actor.CanCreateLargeWork(&input.OwnerTeamID) {
		return nil, ErrForbidden
	}
	status := input.Status
	if status == "" {
		status = entity.LargeWorkStatusPlanned
	}
	if !isAllowedStatus(status) {
		return nil, ErrInvalidStatus
	}
	return s.repo.Create(ctx, repo.CreateInput{
		OwnerTeamID:        input.OwnerTeamID,
		CreatedByUserID:    actor.UserID,
		ParticipantTeamIDs: participantIDs,
		Title:              input.Title,
		WorkType:           input.WorkType,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		WorkTime:           input.WorkTime,
		LocationText:       input.LocationText,
		PEAID:              input.PEAID,
		OperationCenterID:  input.OperationCenterID,
		FeederID:           input.FeederID,
		StationID:          input.StationID,
		Notes:              input.Notes,
		Status:             status,
	})
}

func (s *service) GetByID(ctx context.Context, actor entity.Actor, id int64) (*entity.LargeWorkItem, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !actor.CanViewLargeWork(item) {
		return nil, ErrForbidden
	}
	return item, nil
}

func (s *service) List(ctx context.Context, actor entity.Actor, input ListInput) (*ListOutput, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 || limit > 100 {
		limit = 100
	}
	input.TeamID = actor.ScopedTeamID(input.TeamID)
	items, total, err := s.repo.List(ctx, repo.ListQuery{Page: page, Limit: limit, From: input.From, To: input.To, TeamID: input.TeamID, Statuses: input.Statuses})
	if err != nil {
		return nil, err
	}
	visible := make([]entity.LargeWorkItem, 0, len(items))
	for _, item := range items {
		if actor.CanViewLargeWork(&item) {
			visible = append(visible, item)
		}
	}
	return &ListOutput{Items: visible, Page: page, Limit: limit, Total: total}, nil
}

func (s *service) Update(ctx context.Context, actor entity.Actor, input UpdateInput) (*entity.LargeWorkItem, error) {
	if input.ID <= 0 {
		return nil, ErrInvalidID
	}
	existing, err := s.repo.GetByID(ctx, repo.GetQuery{ID: input.ID})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if !isEditableStatus(existing.Status) {
		return nil, ErrInvalidStateTransition
	}
	if !actor.CanUpdateLargeWork(existing) {
		return nil, ErrForbidden
	}
	if input.OwnerTeamID != nil && !actor.CanCreateLargeWork(input.OwnerTeamID) {
		return nil, ErrForbidden
	}
	if input.Status != nil && !isAllowedStatus(*input.Status) {
		return nil, ErrInvalidStatus
	}
	if input.ParticipantTeamIDs != nil {
		participantIDs := normalizeTeamIDs(append(input.ParticipantTeamIDs, func() int64 {
			if input.OwnerTeamID != nil {
				return *input.OwnerTeamID
			}
			return existing.OwnerTeamID
		}()))
		if len(participantIDs) < 2 {
			return nil, ErrInvalidParticipantTeams
		}
		input.ParticipantTeamIDs = participantIDs
	}
	if input.Title != nil && *input.Title == "" {
		return nil, ErrInvalidTitle
	}
	if input.LocationText != nil && *input.LocationText == "" {
		return nil, ErrInvalidLocation
	}
	startDate := existing.StartDate
	if input.StartDate != nil {
		startDate = *input.StartDate
	}
	endDate := existing.EndDate
	if input.EndDate != nil {
		endDate = input.EndDate
	}
	if endDate != nil && endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}
	updated, err := s.repo.Update(ctx, repo.UpdateInput{
		ID:                 input.ID,
		OwnerTeamID:        input.OwnerTeamID,
		ParticipantTeamIDs: input.ParticipantTeamIDs,
		Title:              input.Title,
		WorkType:           input.WorkType,
		StartDate:          input.StartDate,
		EndDate:            input.EndDate,
		WorkTime:           input.WorkTime,
		LocationText:       input.LocationText,
		PEAID:              input.PEAID,
		OperationCenterID:  input.OperationCenterID,
		FeederID:           input.FeederID,
		StationID:          input.StationID,
		Notes:              input.Notes,
		Status:             input.Status,
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNotFound
	}
	return updated, nil
}

func (s *service) Cancel(ctx context.Context, actor entity.Actor, id int64) (*entity.LargeWorkItem, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !isEditableStatus(item.Status) {
		return nil, ErrInvalidStateTransition
	}
	if !actor.CanCancelLargeWork(item) {
		return nil, ErrForbidden
	}
	status := entity.LargeWorkStatusCancelled
	return s.repo.Update(ctx, repo.UpdateInput{ID: id, Status: &status})
}

func isEditableStatus(status string) bool {
	switch status {
	case entity.LargeWorkStatusDraft, entity.LargeWorkStatusPlanned:
		return true
	default:
		return false
	}
}

func normalizeTeamIDs(ids []int64) []int64 {
	seen := map[int64]struct{}{}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func isAllowedStatus(status string) bool {
	switch status {
	case entity.LargeWorkStatusDraft, entity.LargeWorkStatusPlanned, entity.LargeWorkStatusInProgress, entity.LargeWorkStatusCompleted, entity.LargeWorkStatusCancelled:
		return true
	default:
		return false
	}
}

func (s *service) GetOverview(ctx context.Context, actor entity.Actor, id int64) (*OverviewOutput, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: id})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !actor.CanViewLargeWorkOverview() {
		return nil, ErrForbidden
	}
	tasks, err := s.repo.ListTasksByPlan(ctx, repo.ListTasksQuery{LargeWorkItemID: id})
	if err != nil {
		return nil, err
	}
	progress, teamProgress := summarizeTasks(tasks)
	return &OverviewOutput{Plan: *item, Progress: progress, TeamProgress: teamProgress}, nil
}

func (s *service) ReplaceTasks(ctx context.Context, actor entity.Actor, planID int64, input ReplaceTasksInput) ([]entity.LargeWorkTask, error) {
	if planID <= 0 {
		return nil, ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: planID})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !actor.CanAssignLargeWorkTasks(item, 0) {
		return nil, ErrForbidden
	}
	if len(input.Tasks) == 0 {
		return nil, ErrInvalidTask
	}
	participantSet := map[int64]struct{}{item.OwnerTeamID: {}}
	participants := []int64{item.OwnerTeamID}
	for _, team := range item.Teams {
		if _, ok := participantSet[team.ID]; !ok && team.ID > 0 {
			participantSet[team.ID] = struct{}{}
			participants = append(participants, team.ID)
		}
	}
	tasks := make([]repo.TaskInput, 0, len(input.Tasks))
	for i, t := range input.Tasks {
		if t.AssignedTeamID <= 0 || t.PointLabel == "" || t.WorkType == "" {
			return nil, ErrInvalidTask
		}
		if _, ok := participantSet[t.AssignedTeamID]; !ok {
			participantSet[t.AssignedTeamID] = struct{}{}
			participants = append(participants, t.AssignedTeamID)
		}
		seq := t.Sequence
		if seq <= 0 {
			seq = i + 1
		}
		tasks = append(tasks, repo.TaskInput{AssignedTeamID: t.AssignedTeamID, Sequence: seq, PointLabel: t.PointLabel, Latitude: t.Latitude, Longitude: t.Longitude, WorkType: t.WorkType, WorkDetail: t.WorkDetail, PointCount: t.PointCount, TreeCount: t.TreeCount, ItemCount: t.ItemCount, Notes: t.Notes, Status: entity.LargeWorkTaskStatusTodo, Metadata: t.Metadata})
	}
	existingTasks, err := s.repo.ListTasksByPlan(ctx, repo.ListTasksQuery{LargeWorkItemID: planID})
	if err != nil {
		return nil, err
	}
	for _, existingTask := range existingTasks {
		if existingTask.Status != entity.LargeWorkTaskStatusTodo {
			return nil, ErrInvalidStateTransition
		}
	}
	return s.repo.ReplaceTasksAndParticipants(ctx, repo.ReplaceTasksAndParticipantsInput{LargeWorkItemID: planID, OwnerTeamID: item.OwnerTeamID, ParticipantTeamIDs: participants, Tasks: tasks})
}

func (s *service) ListTasks(ctx context.Context, actor entity.Actor, planID int64) ([]entity.LargeWorkTask, error) {
	_, err := s.GetByID(ctx, actor, planID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListTasksByPlan(ctx, repo.ListTasksQuery{LargeWorkItemID: planID})
}

func (s *service) MyTodos(ctx context.Context, actor entity.Actor, input MyTodosInput) (*MyTodosOutput, error) {
	if actor.TeamID == nil {
		return nil, ErrForbidden
	}
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 || limit > 100 {
		limit = 100
	}
	statuses := input.Statuses
	if len(statuses) == 0 {
		statuses = []string{entity.LargeWorkTaskStatusTodo, entity.LargeWorkTaskStatusInProgress, entity.LargeWorkTaskStatusBlocked}
	}
	items, total, err := s.repo.ListAssignedTasks(ctx, repo.ListAssignedTasksQuery{AssignedTeamID: *actor.TeamID, Statuses: statuses, Page: page, Limit: limit})
	if err != nil {
		return nil, err
	}
	return &MyTodosOutput{Items: items, Page: page, Limit: limit, Total: total}, nil
}

func (s *service) StartTask(ctx context.Context, actor entity.Actor, taskID int64) (*entity.LargeWorkTask, error) {
	task, err := s.getOwnTask(ctx, actor, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status != entity.LargeWorkTaskStatusTodo && task.Status != entity.LargeWorkTaskStatusBlocked {
		return nil, ErrInvalidStateTransition
	}
	now := time.Now()
	status := entity.LargeWorkTaskStatusInProgress
	uid := actor.UserID
	return s.repo.UpdateTask(ctx, repo.UpdateTaskInput{ID: taskID, Status: &status, StartedByUserID: &uid, StartedAt: &now})
}

func (s *service) AttachTaskPhoto(ctx context.Context, actor entity.Actor, taskID int64, kind string, url string) (*entity.LargeWorkTask, error) {
	task, err := s.getOwnTask(ctx, actor, taskID)
	if err != nil {
		return nil, err
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, ErrPhotoRequired
	}
	if task.Status == entity.LargeWorkTaskStatusDone || task.Status == entity.LargeWorkTaskStatusCancelled {
		return nil, ErrInvalidStateTransition
	}
	input := repo.UpdateTaskInput{ID: taskID}
	switch kind {
	case PhotoKindBefore:
		input.BeforePhotoURLs = append(append([]string{}, task.BeforePhotoURLs...), url)
	case PhotoKindAfter:
		input.AfterPhotoURLs = append(append([]string{}, task.AfterPhotoURLs...), url)
	default:
		return nil, ErrInvalidTask
	}
	return s.repo.UpdateTask(ctx, input)
}

func (s *service) CompleteTask(ctx context.Context, actor entity.Actor, taskID int64, input CompleteTaskInput) (*entity.LargeWorkTask, error) {
	task, err := s.getOwnTask(ctx, actor, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status != entity.LargeWorkTaskStatusInProgress {
		return nil, ErrInvalidStateTransition
	}
	after := append(append([]string{}, task.AfterPhotoURLs...), input.AfterPhotoURLs...)
	if len(task.BeforePhotoURLs) == 0 || !allNonEmptyStrings(task.BeforePhotoURLs) || len(after) == 0 || !allNonEmptyStrings(after) {
		return nil, ErrPhotoRequired
	}
	note := input.CompletionNote
	if note == "" {
		return nil, ErrInvalidTask
	}
	now := time.Now()
	status := entity.LargeWorkTaskStatusDone
	uid := actor.UserID
	return s.repo.UpdateTask(ctx, repo.UpdateTaskInput{ID: taskID, Status: &status, AfterPhotoURLs: after, CompletionNote: &note, CompletedByUserID: &uid, CompletedAt: &now})
}

func (s *service) BlockTask(ctx context.Context, actor entity.Actor, taskID int64, reason string) (*entity.LargeWorkTask, error) {
	task, err := s.getOwnTask(ctx, actor, taskID)
	if err != nil {
		return nil, err
	}
	if task.Status == entity.LargeWorkTaskStatusDone || task.Status == entity.LargeWorkTaskStatusCancelled {
		return nil, ErrInvalidStateTransition
	}
	if reason == "" {
		return nil, ErrInvalidTask
	}
	status := entity.LargeWorkTaskStatusBlocked
	return s.repo.UpdateTask(ctx, repo.UpdateTaskInput{ID: taskID, Status: &status, Notes: &reason})
}

func (s *service) getOwnTask(ctx context.Context, actor entity.Actor, taskID int64) (*entity.LargeWorkTask, error) {
	if taskID <= 0 {
		return nil, ErrInvalidID
	}
	task, err := s.repo.GetTaskByID(ctx, repo.GetTaskQuery{ID: taskID})
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}
	if actor.TeamID == nil || *actor.TeamID != task.AssignedTeamID {
		return nil, ErrForbidden
	}
	item, err := s.repo.GetByID(ctx, repo.GetQuery{ID: task.LargeWorkItemID})
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrNotFound
	}
	if !isExecutablePlanStatus(item.Status) {
		return nil, ErrInvalidStateTransition
	}
	return task, nil
}

func isExecutablePlanStatus(status string) bool {
	switch status {
	case entity.LargeWorkStatusCancelled, entity.LargeWorkStatusCompleted:
		return false
	default:
		return true
	}
}

func allNonEmptyStrings(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}

func summarizeTasks(tasks []entity.LargeWorkTask) (ProgressSummary, []TeamProgressSummary) {
	progress := ProgressSummary{Total: len(tasks)}
	byTeam := map[int64]*TeamProgressSummary{}
	order := []int64{}
	for _, task := range tasks {
		applyStatus(&progress, task.Status)
		team := byTeam[task.AssignedTeamID]
		if team == nil {
			team = &TeamProgressSummary{AssignedTeamID: task.AssignedTeamID}
			byTeam[task.AssignedTeamID] = team
			order = append(order, task.AssignedTeamID)
		}
		team.Total++
		applyTeamStatus(team, task.Status)
	}
	out := make([]TeamProgressSummary, 0, len(order))
	for _, id := range order {
		out = append(out, *byTeam[id])
	}
	return progress, out
}

func applyStatus(p *ProgressSummary, status string) {
	switch status {
	case entity.LargeWorkTaskStatusDone:
		p.Done++
	case entity.LargeWorkTaskStatusInProgress:
		p.InProgress++
	case entity.LargeWorkTaskStatusBlocked:
		p.Blocked++
	case entity.LargeWorkTaskStatusCancelled:
		p.Cancelled++
	default:
		p.Todo++
	}
}
func applyTeamStatus(p *TeamProgressSummary, status string) {
	switch status {
	case entity.LargeWorkTaskStatusDone:
		p.Done++
	case entity.LargeWorkTaskStatusInProgress:
		p.InProgress++
	case entity.LargeWorkTaskStatusBlocked:
		p.Blocked++
	case entity.LargeWorkTaskStatusCancelled:
		p.Cancelled++
	default:
		p.Todo++
	}
}
