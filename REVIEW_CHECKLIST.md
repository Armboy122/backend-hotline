# Code Review Checklist

> **Role:** Code Reviewer / QA Engineer
> **Purpose:** Standardized checklist สำหรับการ review PR และ merge code
> **Audience:** All reviewers, pull request authors

---

## Pre-Review Checklist (Before Starting Review)

- [ ] PR has clear title and description
- [ ] PR links to relevant issue/ticket
- [ ] Branch name follows convention: `feature/*, fix/*, docs/*, refactor/*`
- [ ] No conflicts with base branch
- [ ] CI/CD pipeline passed (or known failures explained)
- [ ] Author filled commit message properly
- [ ] No WIP (Work In Progress) label
- [ ] Reasonable number of commits (not squashed unnecessarily, but organized)

---

## Code Quality Checklist

### C1: Style & Readability

- [ ] **Naming Conventions**
  - [ ] Functions use camelCase and are descriptive
  - [ ] Constants use PascalCase (not SNAKE_CASE)
  - [ ] Package names are lowercase, single word
  - [ ] Exported types/functions start with uppercase
  - [ ] Unexported types/functions start with lowercase

  ```go
  // ✅ GOOD
  func GetUserByID(id uint) {}
  const DefaultPageSize = 20

  // ❌ BAD
  func get_user_by_id(id uint) {}
  const DEFAULT_PAGE_SIZE = 20
  ```

- [ ] **Code Formatting**
  - [ ] Passes `gofmt` / `goimports`
  - [ ] Line length < 120 characters (ideally < 100)
  - [ ] Consistent indentation (tabs)
  - [ ] No trailing whitespace

  ```bash
  # Quick check
  gofmt -l ./...  # Should return empty
  goimports -l ./...  # Should return empty
  ```

- [ ] **Comments & Documentation**
  - [ ] Public functions have comments
  - [ ] Comments explain WHY, not WHAT
  - [ ] No obvious/redundant comments
  - [ ] Complex logic is explained

  ```go
  // ✅ GOOD - Explain WHY
  // Retry mechanism with exponential backoff to handle temporary DB unavailability
  func retryWithBackoff(fn func() error) error { }

  // ❌ BAD - Obvious comment
  // Get user by ID
  func GetUserByID(id uint) { }
  ```

- [ ] **File Organization**
  - [ ] File size < 400 lines
  - [ ] Related code is grouped together
  - [ ] Imports are organized (std, third-party, internal)
  - [ ] No commented-out code blocks (should be deleted)

### C2: Architecture & Design

- [ ] **Separation of Concerns**
  - [ ] Handler only handles HTTP
  - [ ] Service contains business logic
  - [ ] Repository handles data access
  - [ ] Models are pure domain objects
  - [ ] No business logic in handlers

  ```go
  // ✅ GOOD - Clear separation
  // handlers/v1/task.go
  func (h *TaskHandler) CreateTask(c *gin.Context) {
      var req CreateTaskRequest
      c.BindJSON(&req)
      result, err := h.service.CreateTask(req)  // Call service
  }

  // services/task.go
  func (s *TaskService) CreateTask(req CreateTaskRequest) (*Task, error) {
      if err := req.Validate(); err != nil {  // Validation here
          return nil, err
      }
      return s.repo.Create(req.toModel())  // Call repo
  }

  // ❌ BAD - Business logic in handler
  func (h *TaskHandler) CreateTask(c *gin.Context) {
      // Complex validation and business logic here
  }
  ```

- [ ] **No Circular Dependencies**
  - [ ] Handlers don't import other handlers
  - [ ] Services don't import handlers
  - [ ] Use interfaces to break dependency chains
  - [ ] `go mod graph` shows no cycles

- [ ] **Dependency Injection**
  - [ ] Dependencies passed via constructor
  - [ ] No global singletons
  - [ ] Easy to mock for testing

  ```go
  // ✅ GOOD
  type TaskService struct {
      repo TaskRepository
      logger Logger
  }

  func NewTaskService(repo TaskRepository, logger Logger) *TaskService {
      return &TaskService{repo, logger}
  }

  // ❌ BAD
  var globalRepo = &TaskRepository{}  // Hard to test

  type TaskService struct {}

  func (s *TaskService) Create() {
      globalRepo.Save()  // Uses global
  }
  ```

- [ ] **Function Size & Complexity**
  - [ ] Functions < 50 lines (ideally < 30)
  - [ ] Cyclomatic complexity < 10
  - [ ] Single responsibility principle
  - [ ] Extract complex logic into helper functions

  ```bash
  # Check complexity
  go get -u github.com/fzipp/gocyclo/cmd/gocyclo
  gocyclo -over 10 ./...
  ```

### C3: Error Handling

- [ ] **No Unhandled Errors**
  - [ ] Every function that returns error is checked
  - [ ] Errors are not silently ignored
  - [ ] Error handling is consistent

  ```go
  // ✅ GOOD
  if err := db.Create(task).Error; err != nil {
      logger.Error("failed to create task", err)
      return fmt.Errorf("create task failed: %w", err)
  }

  // ❌ BAD
  db.Create(task)  // Ignore error!

  // ❌ BAD
  if err != nil {
      return err  // Lost context
  }
  ```

- [ ] **Proper Error Types**
  - [ ] Uses custom error types (not string comparisons)
  - [ ] Errors are wrapped with context
  - [ ] Uses `fmt.Errorf` with `%w` for wrapping

  ```go
  // ✅ GOOD
  if err != nil {
      return fmt.Errorf("get user by id: %w", err)
  }

  // ❌ BAD
  if err.Error() == "not found" { }  // String comparison

  // ❌ BAD
  return errors.New("failed")  // No context
  ```

- [ ] **User-Friendly Error Messages**
  - [ ] API errors describe what went wrong
  - [ ] No internal implementation details leaked
  - [ ] Error messages are actionable

  ```go
  // ✅ GOOD
  {
      "error": {
          "code": "VALIDATION_ERROR",
          "message": "Email is required",
          "details": { "field": "email" }
      }
  }

  // ❌ BAD
  {
      "error": "CONSTRAINT_VIOLATION: duplicate key value violates unique constraint"
  }
  ```

- [ ] **Panic Usage**
  - [ ] No panic for expected errors
  - [ ] panic only for truly unrecoverable situations
  - [ ] No panic in production code (except main)

### C4: Security

- [ ] **Input Validation**
  - [ ] All external inputs validated
  - [ ] Struct tags have `binding` constraints
  - [ ] Length/format validation present
  - [ ] SQL injection impossible (using parameterized queries)

  ```go
  // ✅ GOOD
  type CreateTaskRequest struct {
      Title     string    `json:"title" binding:"required,min=3,max=255"`
      WorkDate  time.Time `json:"workDate" binding:"required"`
      TeamID    uint      `json:"teamId" binding:"required,gt=0"`
  }

  // ❌ BAD
  type CreateTaskRequest struct {
      Title    string
      WorkDate time.Time
      TeamID   uint
  }
  ```

- [ ] **Sensitive Data Handling**
  - [ ] No passwords in logs
  - [ ] No API keys in code
  - [ ] No PII in error messages
  - [ ] Database credentials use environment variables

  ```go
  // ✅ GOOD
  logger.Info("user login", "username", user.Username)

  // ❌ BAD
  logger.Info("user login", "password", user.Password)
  ```

- [ ] **Authentication & Authorization**
  - [ ] JWT middleware checks present
  - [ ] Ownership/permission checks before operations
  - [ ] Admin operations are protected
  - [ ] Token expiry is reasonable (15-60 min)

  ```go
  // ✅ GOOD
  protected := engine.Group("/api/v1")
  protected.Use(middleware.JWTMiddleware())
  protected.DELETE("/tasks/:id", handlers.DeleteTask)

  // In handler
  userID := GetUserIDFromContext(c)
  if task.CreatedByID != userID && !isAdmin(userID) {
      c.JSON(403, ErrorResponse{Code: "FORBIDDEN"})
      return
  }

  // ❌ BAD - No protection
  engine.DELETE("/api/v1/tasks/:id", handlers.DeleteTask)
  ```

- [ ] **No Hardcoded Secrets**
  - [ ] No passwords in code
  - [ ] No API keys in code
  - [ ] No JWT secrets hardcoded
  - [ ] All secrets come from config/env

  ```bash
  # Check for secrets
  git diff HEAD~1 | grep -E "password|secret|key|token"
  ```

- [ ] **No External Command Injection**
  - [ ] No `os/exec` with user input directly
  - [ ] Arguments properly escaped/quoted

  ```go
  // ❌ BAD - Command injection risk
  cmd := exec.Command("sh", "-c", userInput)

  // ✅ GOOD
  cmd := exec.Command("command", userInput)  // Separate args
  ```

---

## Testing Checklist

### T1: Test Coverage

- [ ] **Tests Added for New Code**
  - [ ] New functions have test cases
  - [ ] New features have integration tests
  - [ ] Edge cases are covered
  - [ ] Error paths are tested

  ```bash
  # Check coverage
  go test -cover ./...
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
  ```

- [ ] **Test Quality**
  - [ ] Test names describe what is tested
  - [ ] Table-driven tests for multiple cases
  - [ ] Tests don't depend on each other
  - [ ] Tests are deterministic (no random failures)

  ```go
  // ✅ GOOD - Clear test names
  func TestCreateTask_WithValidInput_ReturnsTask(t *testing.T) { }
  func TestCreateTask_WithEmptyTitle_ReturnsError(t *testing.T) { }

  // ✅ GOOD - Table-driven
  testCases := []struct {
      name    string
      input   CreateTaskRequest
      wantErr bool
  }{
      {"valid input", validReq, false},
      {"empty title", emptyReq, true},
  }

  // ❌ BAD - Vague test names
  func TestCreateTask(t *testing.T) { }
  ```

- [ ] **Mock Usage**
  - [ ] Real implementations used where possible
  - [ ] Mocks only for external dependencies
  - [ ] Mock behavior is reasonable
  - [ ] Mocks are verified

  ```go
  // ✅ GOOD - Test against real service, mock repo
  mockRepo := NewMockTaskRepository()
  service := NewTaskService(mockRepo)

  // ❌ BAD - Over-mocking
  mockService := NewMockTaskService()
  handler := NewTaskHandler(mockService)  // Mocking the thing being tested!
  ```

### T2: Benchmark Tests

- [ ] **If Performance Critical**
  - [ ] Benchmark tests added for critical paths
  - [ ] Benchmarks don't regress
  - [ ] Results documented

  ```bash
  go test -bench=. -benchmem ./...
  ```

---

## Database Checklist

### D1: GORM Models

- [ ] **Proper Tags**
  - [ ] Primary key defined with `gorm:"primaryKey"`
  - [ ] Foreign keys use `gorm:"foreignKey:FieldName"`
  - [ ] Indexes on frequently queried fields
  - [ ] Soft delete fields: `DeletedAt gorm.DeletedAt`

  ```go
  // ✅ GOOD
  type Task struct {
      ID        uint           `gorm:"primaryKey"`
      Title     string         `gorm:"index"`
      TeamID    uint           `gorm:"index"`
      WorkDate  time.Time      `gorm:"index"`
      CreatedAt time.Time
      UpdatedAt time.Time
      DeletedAt gorm.DeletedAt `gorm:"index"`
  }

  // ❌ BAD
  type Task struct {
      ID        uint
      Title     string         `gorm:"index"`  // Every field indexed!
      Description string       `gorm:"index"`  // No, just used 3-4 places
  }
  ```

- [ ] **Relationships**
  - [ ] Foreign keys are indexed
  - [ ] Relationships are properly defined
  - [ ] Cascade behavior is correct
  - [ ] No data integrity issues

  ```go
  // ✅ GOOD
  type Task struct {
      ID       uint
      TeamID   uint `gorm:"index"`
      Team     *Team `gorm:"foreignKey:TeamID"`
  }

  // ❌ BAD
  type Task struct {
      ID     uint
      TeamID uint  // Not indexed, queried by team frequently
      Team   *Team
  }
  ```

### D2: Queries

- [ ] **No N+1 Queries**
  - [ ] Eager loading used for relationships
  - [ ] Preload only needed fields
  - [ ] No loops doing database queries

  ```go
  // ✅ GOOD - Single query with preload
  var tasks []Task
  db.Preload("Team").Preload("JobType").
      Where("work_date = ?", date).
      Find(&tasks)

  // ❌ BAD - N+1 queries
  var tasks []Task
  db.Find(&tasks)
  for _, task := range tasks {
      db.First(&task.Team, task.TeamID)  // Loop query!
  }
  ```

- [ ] **Query Optimization**
  - [ ] Uses Select to limit columns when appropriate
  - [ ] Pagination limits are reasonable
  - [ ] Index-friendly WHERE clauses
  - [ ] No full table scans

  ```go
  // ✅ GOOD - Use pagination
  db.Where("team_id = ?", teamID).
      Limit(20).Offset((page-1)*20).
      Find(&tasks)

  // ❌ BAD - No limit
  db.Where("team_id = ?", teamID).Find(&tasks)  // Could be thousands

  // ✅ GOOD - Index-friendly
  db.Where("work_date >= ? AND team_id = ?", date, teamID).Find(&tasks)

  // ❌ BAD - Index unfriendly
  db.Where("YEAR(work_date) = ? AND team_id = ?", year, teamID)
  ```

### D3: Transactions

- [ ] **Multi-step Operations Use Transactions**
  - [ ] All-or-nothing semantics
  - [ ] Proper rollback on error
  - [ ] No deadlock risk

  ```go
  // ✅ GOOD
  return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
      if err := tx.Create(task).Error; err != nil {
          return err
      }
      if err := tx.Create(auditLog).Error; err != nil {
          return err  // Automatic rollback
      }
      return nil
  })

  // ❌ BAD - No transaction
  db.Create(task)
  db.Create(auditLog)  // If this fails, task is orphaned
  ```

### D4: Migrations

- [ ] **New Tables/Columns Covered**
  - [ ] Migration files added for schema changes
  - [ ] Migrations are idempotent
  - [ ] Both up and down migrations work
  - [ ] Data migration handled if needed

  ```bash
  # Check migrations work
  go run cmd/migrate/main.go
  ```

---

## Performance Checklist

### P1: Critical Paths

- [ ] **Dashboard Endpoint**
  - [ ] Response time < 2 seconds
  - [ ] Aggregation uses database queries (not in-memory)
  - [ ] Proper indexes on aggregated fields

  ```bash
  # Test performance
  time curl -s http://localhost:8080/v1/dashboard/stats | jq . > /dev/null
  ```

- [ ] **List Endpoints**
  - [ ] Default pagination enforced
  - [ ] Max page size enforced
  - [ ] Response time < 1 second

  ```go
  // ✅ GOOD
  const (
      DefaultPageSize = 20
      MaxPageSize = 100
  )

  size := query.Get("size")
  if size > MaxPageSize {
      size = MaxPageSize
  }
  ```

### P2: Memory Leaks

- [ ] **No Goroutine Leaks**
  - [ ] Background goroutines properly cleaned up
  - [ ] Contexts used correctly
  - [ ] Channels closed properly

  ```bash
  # Check for leaks
  go test -v -race ./...
  ```

- [ ] **No Large Objects in Memory**
  - [ ] No unbounded caching
  - [ ] Large queries paginated
  - [ ] File uploads streamed (not fully loaded)

---

## Documentation Checklist

### D1: Code Documentation

- [ ] **Public API Documented**
  - [ ] All exported functions have comments
  - [ ] Complex logic has explanatory comments
  - [ ] Parameter descriptions clear
  - [ ] Return value documented

  ```go
  // ✅ GOOD
  // GetTasksForTeam returns all tasks for a specific team on a given date.
  // Results are sorted by creation time (newest first).
  // Returns empty slice if no tasks found; never returns nil.
  func GetTasksForTeam(teamID uint, date time.Time) ([]*Task, error) {

  // ❌ BAD - No documentation
  func GetTasksForTeam(teamID uint, date time.Time) ([]*Task, error) {
  ```

### D2: Changelog

- [ ] **CHANGELOG.md Updated**
  - [ ] Change is documented in proper section (Added/Fixed/Changed/Removed)
  - [ ] Breaking changes clearly marked
  - [ ] Version follows semver

  ```markdown
  ## [1.1.0] - 2026-01-26

  ### Added
  - JWT authentication endpoints
  - Rate limiting middleware

  ### Fixed
  - Dashboard response format inconsistency (issue #42)

  ### Changed
  - Database credentials now use environment variables

  ### Removed
  - Deprecated /api/v0 endpoints
  ```

### D3: README Updates

- [ ] **README Updated if Needed**
  - [ ] New features mentioned
  - [ ] Setup instructions updated
  - [ ] Example usage updated
  - [ ] Breaking changes documented

---

## Git & Commit Checklist

### G1: Commit Quality

- [ ] **Clear Commit Messages**
  - [ ] First line < 50 characters
  - [ ] Descriptive body (if needed)
  - [ ] References issue number: `Fixes #42`
  - [ ] Uses imperative mood: "Add feature" not "Added feature"

  ```
  ✅ GOOD
  feat(auth): add JWT authentication endpoints

  Implements login, logout, refresh, and me endpoints
  with proper token rotation and expiry handling.

  Fixes #42

  ❌ BAD
  fix stuff
  updated the things
  asdf
  ```

- [ ] **Logical Commits**
  - [ ] One feature per commit
  - [ ] Related changes grouped together
  - [ ] No unrelated changes mixed in
  - [ ] Reversible (can git revert)

- [ ] **Clean Commit History**
  - [ ] No merge commits (rebase if needed)
  - [ ] No WIP commits
  - [ ] No "fix previous commit" commits (squash instead)

### G2: Branch Hygiene

- [ ] **Branch Convention**
  ```
  feature/auth-jwt              # New feature
  fix/dashboard-response-format # Bug fix
  docs/api-documentation       # Documentation
  refactor/error-handling       # Refactoring
  test/add-integration-tests   # Tests
  chore/update-dependencies    # Maintenance
  ```

---

## Final Approval Checklist

Before approving, verify:

- [ ] **All Checkboxes Reviewed**
  - [ ] Code quality: ✅
  - [ ] Tests: ✅
  - [ ] Database: ✅
  - [ ] Security: ✅
  - [ ] Performance: ✅
  - [ ] Documentation: ✅
  - [ ] Git: ✅

- [ ] **No Major Issues**
  - [ ] No blocker issues found
  - [ ] All requests addressed or explained
  - [ ] Code is production-ready

- [ ] **CI/CD Status**
  - [ ] All GitHub Actions passed
  - [ ] No test failures
  - [ ] Code coverage not decreased

- [ ] **Approval Decision**
  - [ ] ✅ APPROVE - Ready to merge
  - [ ] ⏸️ REQUEST CHANGES - Wait for fixes
  - [ ] 👀 COMMENT - Informational only

---

## Common Issues & Quick Fixes

### Issue: Large PR (>500 lines)
- **Fix:** Ask for smaller, logical PR splits
- **Reason:** Easier to review, test, and revert if needed

### Issue: No tests included
- **Fix:** Request unit/integration tests
- **Reason:** Can't verify correctness, regression risk

### Issue: Unclear commit messages
- **Fix:** Request clarification via comment
- **Reason:** Important for git history and debugging

### Issue: Missing error handling
- **Fix:** Request error checking
- **Reason:** Silent failures are hard to debug

### Issue: N+1 queries
- **Fix:** Request eager loading/optimization
- **Reason:** Performance degradation at scale

### Issue: Credentials in code
- **Fix:** Reject immediately, request environment variables
- **Reason:** Security risk

---

## Review Time Expectations

| PR Size | Estimated Review Time |
|---------|----------------------|
| < 100 lines | 15-30 minutes |
| 100-300 lines | 30-60 minutes |
| 300-500 lines | 60-90 minutes |
| > 500 lines | Ask for split into smaller PRs |

**Best Practice:** Don't review while tired; take breaks every 30 minutes

---

## Post-Approval

- [ ] Set merge strategy (squash/rebase/merge)
- [ ] Confirm CI/CD passes before merge
- [ ] Delete branch after merge
- [ ] Monitor logs for any issues after deployment
- [ ] Close related issues/tickets

---

*Last Updated: 2026-01-26*
*Review Cycle: Every PR*
*Escalation: Contact Tech Lead if major issues found*
