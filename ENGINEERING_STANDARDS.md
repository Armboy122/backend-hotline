# Engineering Standards & Guidelines

> **Role:** Engineering Lead / Tech Lead
> **Purpose:** กำหนดมาตรฐานการพัฒนา architecture decisions และ best practices
> **Audience:** Development team, code reviewers

---

## Code Standards

### 1. Go Code Style

#### Naming Conventions
```go
// ✅ GOOD
type UserRepository struct{}
func (ur *UserRepository) GetUserByID(id uint) (*User, error) {}
var maxRetries = 3

// ❌ BAD
type userRepository struct{}  // private, should use uppercase
func GetUserById() {}  // use ID not Id
var MAX_RETRIES = 3  // constants use camelCase not SNAKE_CASE
```

#### File Organization
```
handlers/
├── v1/
│   ├── task.go              // Handler functions เท่านั้น
│   ├── team.go
│   └── user.go
├── helpers.go               // Common helper functions
└── middleware.go            # Middleware only

models/
├── models.go               // GORM models ทั้งหมด
└── types.go                // Custom types (StringArray, etc.)

dto/
├── request.go              // Request DTOs
├── response.go             // Response DTOs
└── errors.go               // Error definitions

repositories/
├── task_repository.go      // Database access layer
├── user_repository.go
└── interfaces.go           // Repository interfaces

services/
├── task_service.go         // Business logic
├── user_service.go
└── interfaces.go           // Service interfaces
```

#### File Size Guidelines
- Max 300-400 lines per file
- One handler per file (except small related operations)
- Break into smaller files if logic exceeds 400 lines

### 2. Error Handling

#### Error Definition
```go
// ✅ GOOD - Define custom errors
package errors

type ErrorCode string

const (
    ErrInvalidInput    ErrorCode = "INVALID_INPUT"
    ErrNotFound        ErrorCode = "NOT_FOUND"
    ErrUnauthorized    ErrorCode = "UNAUTHORIZED"
    ErrConflict        ErrorCode = "CONFLICT"
    ErrInternalError   ErrorCode = "INTERNAL_ERROR"
)

type APIError struct {
    Code      ErrorCode   `json:"code"`
    Message   string      `json:"message"`
    Details   interface{} `json:"details,omitempty"`
    Timestamp int64       `json:"timestamp"`
}

// ❌ BAD - Return error strings
return fmt.Errorf("user not found")
```

#### Error Handling in Handlers
```go
// ✅ GOOD
func GetUser(c *gin.Context) {
    id := c.Param("id")

    user, err := ur.GetUserByID(id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            c.JSON(404, ErrorResponse{
                Code:    "NOT_FOUND",
                Message: "User not found",
            })
            return
        }
        // Log internal error
        logger.Error("failed to get user", err)
        c.JSON(500, ErrorResponse{
            Code:    "INTERNAL_ERROR",
            Message: "Internal server error",
        })
        return
    }

    c.JSON(200, user)
}

// ❌ BAD
func GetUser(c *gin.Context) {
    user, err := ur.GetUserByID(c.Param("id"))
    if err != nil {
        c.JSON(500, err.Error())  // Generic error
        return
    }
}
```

### 3. Function Guidelines

#### Function Size
- Max 50 lines per function
- One responsibility per function (SRP)
- Break into smaller functions if exceeds 50 lines

#### Function Signature
```go
// ✅ GOOD - Clear return types
func GetUserByID(id uint) (*User, error) {}
func ValidateEmail(email string) (bool, error) {}

// ❌ BAD - Unclear return types
func GetUserByID(id uint) (interface{}, error) {}
func ValidateEmail(email string) interface{} {}
```

#### Comments
```go
// ✅ GOOD - Explain WHY, not WHAT
// GetUserByID retrieves user from database and caches result
// to avoid repeated queries for the same user within 5 minutes
func GetUserByID(id uint) (*User, error) {}

// ❌ BAD - Obvious comments
// Get user by ID
func GetUserByID(id uint) {}

// ❌ BAD - No comments for non-obvious logic
func complexBusinessLogic() {}  // needs comment explaining the algorithm
```

### 4. Constants & Variables

```go
// ✅ GOOD
const (
    DefaultPageSize = 20
    MaxPageSize     = 100
    RequestTimeout  = 30 * time.Second
)

var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidEmail = errors.New("invalid email format")
)

// ❌ BAD
const DEFAULT_PAGE_SIZE = 20  // Should be DefaultPageSize
magic := 42  // What is 42? Use named constant
```

---

## Architecture Decisions

### 1. Project Structure (Clean Architecture)

```
├── handlers/      // HTTP layer (Gin)
├── dto/          // Data Transfer Objects
├── services/     // Business logic
├── repositories/ // Data access layer
├── models/       // Domain models (GORM)
├── middleware/   // HTTP middleware
├── config/       // Configuration
├── database/     // Database setup
└── pkg/          // Shared utilities
```

**Rationale:**
- Separation of concerns
- Easy to test (mock repositories & services)
- Easy to swap implementations
- Clear dependency flow

### 2. Dependency Injection

```go
// ✅ GOOD - Constructor injection
type TaskService struct {
    repo TaskRepository
    logger Logger
}

func NewTaskService(repo TaskRepository, logger Logger) *TaskService {
    return &TaskService{repo, logger}
}

// ❌ BAD - Global singletons
var TaskRepo = &TaskRepository{}  // Hard to test
```

### 3. Database Access Pattern

```go
// Layer 1: Repository (Data access)
type TaskRepository interface {
    GetByID(id uint) (*Task, error)
    Create(task *Task) error
    Update(task *Task) error
    Delete(id uint) error
    List(page, size int) ([]*Task, int64, error)
}

// Layer 2: Service (Business logic)
type TaskService struct {
    repo TaskRepository
}

func (ts *TaskService) CreateTask(input CreateTaskInput) (*Task, error) {
    // Validation
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Convert DTO to model
    task := &Task{
        Title: input.Title,
        // ...
    }

    // Call repository
    if err := ts.repo.Create(task); err != nil {
        return nil, fmt.Errorf("create failed: %w", err)
    }

    return task, nil
}

// Layer 3: Handler (HTTP layer)
func (h *TaskHandler) CreateTask(c *gin.Context) {
    var input CreateTaskInput
    if err := c.BindJSON(&input); err != nil {
        c.JSON(400, ErrorResponse{Code: "INVALID_INPUT"})
        return
    }

    task, err := h.service.CreateTask(input)
    if err != nil {
        // Handle error
        return
    }

    c.JSON(201, task)
}
```

### 4. Request/Response Pattern

```go
// Request DTOs
type CreateTaskRequest struct {
    Title       string    `json:"title" binding:"required,min=3,max=255"`
    Description string    `json:"description" binding:"max=1000"`
    WorkDate    time.Time `json:"workDate" binding:"required"`
    TeamID      uint      `json:"teamId" binding:"required"`
}

type TaskResponse struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    WorkDate    time.Time `json:"workDate"`
    Team        TeamInfo  `json:"team"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

// Consistent response envelope
type SuccessResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Meta    *PageMeta   `json:"meta,omitempty"`
}

type ErrorResponse struct {
    Success bool        `json:"success"`
    Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type PageMeta struct {
    Page      int   `json:"page"`
    Size      int   `json:"size"`
    Total     int64 `json:"total"`
    TotalPage int   `json:"totalPage"`
}
```

### 5. GORM Model Pattern

```go
// ✅ GOOD - Clear relationships
type Task struct {
    ID        uint        `gorm:"primaryKey"`
    Title     string      `gorm:"index"`
    TeamID    uint        `gorm:"index"`
    Team      *Team       `gorm:"foreignKey:TeamID"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Method receivers for business logic
func (t *Task) IsExpired() bool {
    return time.Since(t.CreatedAt) > 30*24*time.Hour
}

// ❌ BAD - Methods in models for complex logic
func (t *Task) ComplexBusinessLogic() {
    // This should be in service, not model
}
```

### 6. Middleware Order

```go
// router.go
func SetupRoutes(engine *gin.Engine) {
    // 1. Global middleware
    engine.Use(middleware.CORSMiddleware())
    engine.Use(middleware.LoggingMiddleware())
    engine.Use(middleware.SecurityHeadersMiddleware())
    engine.Use(middleware.ErrorRecoveryMiddleware())

    // 2. Health check (no auth needed)
    engine.GET("/health", handlers.HealthCheck)

    // 3. Public routes
    public := engine.Group("/api/v1")
    {
        public.POST("/auth/login", handlers.Login)
        public.POST("/auth/register", handlers.Register)
    }

    // 4. Protected routes
    protected := engine.Group("/api/v1")
    protected.Use(middleware.JWTMiddleware())
    {
        protected.POST("/tasks", handlers.CreateTask)
        protected.GET("/tasks/:id", handlers.GetTask)

        // Admin-only routes
        admin := protected.Group("/admin")
        admin.Use(middleware.RoleMiddleware("admin"))
        {
            admin.DELETE("/tasks/:id", handlers.DeleteTask)
        }
    }
}
```

---

## Testing Standards

### 1. Test Naming & Organization

```go
// ✅ GOOD - Table-driven tests
func TestCreateTask(t *testing.T) {
    testCases := []struct {
        name      string
        input     CreateTaskInput
        wantErr   bool
        wantCode  ErrorCode
    }{
        {
            name:      "valid input",
            input:     CreateTaskInput{Title: "Test"},
            wantErr:   false,
        },
        {
            name:      "empty title",
            input:     CreateTaskInput{Title: ""},
            wantErr:   true,
            wantCode:  ErrInvalidInput,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Mock Pattern

```go
// Use interfaces for mocking
type TaskRepository interface {
    Create(task *Task) error
    GetByID(id uint) (*Task, error)
}

// Mock implementation
type MockTaskRepository struct {
    mock.Mock
}

func (m *MockTaskRepository) Create(task *Task) error {
    args := m.Called(task)
    return args.Error(0)
}

// Use in tests
func TestTaskService_Create(t *testing.T) {
    mockRepo := new(MockTaskRepository)
    mockRepo.On("Create", mock.MatchedBy(func(t *Task) bool {
        return t.Title != ""
    })).Return(nil)

    service := NewTaskService(mockRepo)
    // Test
}
```

### 3. Test Coverage Target

- **Minimum:** 70% overall
- **Target:** 80%+ for critical paths
- **Excluded:** Database drivers, third-party integrations
- **Required:** All error paths covered

---

## Database Standards

### 1. Index Strategy

```go
// ✅ GOOD - Only index fields used in WHERE/JOIN
type Task struct {
    ID        uint      `gorm:"primaryKey"`
    WorkDate  time.Time `gorm:"index"`  // Often filtered by date
    TeamID    uint      `gorm:"index"`  // Foreign key
    FeederID  uint      `gorm:"index"`  // Often filtered
    CreatedAt time.Time `gorm:"index"`  // For sorting

    // Composite index for common query
    // SELECT * FROM tasks WHERE team_id = ? AND work_date = ?
}

type Task struct {
    ID       uint
    TeamID   uint
    WorkDate time.Time
}

func (Task) TableName() string {
    return "tasks"
}

func (Task) Indexes() []schema.Index {
    return []schema.Index{
        schema.Index{
            Fields: []schema.Field{{Name: "team_id"}, {Name: "work_date"}},
            Name:   "idx_team_workdate",
        },
    }
}

// ❌ BAD - Index everything
type Task struct {
    ID           uint      `gorm:"index"`
    Title        string    `gorm:"index"`  // Not needed
    Description  string    `gorm:"index"`  // Not needed
    Detail       string    `gorm:"index"`  // Not needed
}
```

### 2. Query Optimization

```go
// ✅ GOOD - Use select to limit columns
db.Select("id", "title", "team_id").
    Where("team_id = ?", teamID).
    Limit(20).
    Find(&tasks)

// ✅ GOOD - Eager load only needed relations
db.Preload("Team").
    Preload("JobType").
    Where("work_date = ?", date).
    Find(&tasks)

// ❌ BAD - Load all columns
db.Where("team_id = ?", teamID).Find(&tasks)

// ❌ BAD - N+1 queries
for _, task := range tasks {
    db.Where("id = ?", task.TeamID).First(&task.Team)  // Loop query!
}
```

### 3. Transaction Usage

```go
// ✅ GOOD - Use transactions for multi-step operations
func (s *TaskService) BulkCreateTasks(tasks []*Task) error {
    return s.db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
        for _, task := range tasks {
            if err := tx.Create(task).Error; err != nil {
                return err
            }
        }
        return nil
    })
}

// ✅ GOOD - Rollback on error
tx := db.BeginTx(ctx, nil)
if err := tx.Create(task).Error; err != nil {
    tx.Rollback()
    return err
}
if err := tx.Commit().Error; err != nil {
    return err
}

// ❌ BAD - No transaction for related operations
db.Create(task)
db.Create(auditLog)  // If this fails, task is orphaned
```

---

## Security Standards

### 1. Input Validation

```go
// ✅ GOOD - Validate all inputs
type CreateTaskRequest struct {
    Title      string    `json:"title" binding:"required,min=3,max=255"`
    WorkDate   time.Time `json:"workDate" binding:"required"`
    TeamID     uint      `json:"teamId" binding:"required,gt=0"`
    FeederID   uint      `json:"feederId" binding:"omitempty,gt=0"`
    Latitude   *float64  `json:"latitude" binding:"omitempty,latitude"`
    Longitude  *float64  `json:"longitude" binding:"omitempty,longitude"`
}

// ❌ BAD - No validation
type CreateTaskRequest struct {
    Title    string
    WorkDate time.Time
    TeamID   uint
}
```

### 2. Sensitive Data Handling

```go
// ✅ GOOD - Never log sensitive data
logger.Info("user login", "username", user.Username)  // OK
// Don't log password!

// ✅ GOOD - Use pointers for password fields
type User struct {
    ID       uint
    Username string
    Password *string  // Never serialize in responses
}

// ❌ BAD - Don't include password in responses
type UserResponse struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"`  // NEVER!
}
```

### 3. Authentication Checks

```go
// ✅ GOOD - Check ownership before operations
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    taskID := c.Param("id")
    userID := GetUserIDFromContext(c)

    // Check authorization
    task := &Task{}
    if err := h.db.First(task, taskID).Error; err != nil {
        c.JSON(404, ErrorResponse{Code: "NOT_FOUND"})
        return
    }

    if task.CreatedByID != userID && !isAdmin(userID) {
        c.JSON(403, ErrorResponse{Code: "FORBIDDEN"})
        return
    }

    // Proceed with update
}

// ❌ BAD - No ownership check
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    var input UpdateTaskInput
    c.BindJSON(&input)
    h.db.Model(&Task{}).Where("id = ?", c.Param("id")).Updates(input)
    c.JSON(200, "OK")
}
```

---

## Documentation Standards

### 1. API Documentation

Every public function should have a comment:

```go
// GetTaskByID retrieves a single task by its ID.
// Returns ErrNotFound if task doesn't exist or is soft-deleted.
func (r *TaskRepository) GetTaskByID(id uint) (*Task, error) {
    // Implementation
}

// CreateTask creates a new task and returns its ID.
// Input validation should be done before calling this method.
// Returns error if database operation fails.
func (r *TaskRepository) CreateTask(task *Task) error {
    // Implementation
}
```

### 2. Handler Documentation

```go
// POST /v1/tasks
// CreateTask creates a new task with images
// Requires: Authorization header with valid JWT
// Request body: CreateTaskRequest
// Response: 201 Created with TaskResponse
func (h *TaskHandler) CreateTask(c *gin.Context) {
    // Implementation
}
```

---

## Code Review Checklist

Reviewers should check:

- [ ] **Code Style**
  - [ ] Follows naming conventions
  - [ ] Functions < 50 lines
  - [ ] Files < 400 lines
  - [ ] Comments explain WHY not WHAT

- [ ] **Architecture**
  - [ ] No circular dependencies
  - [ ] Proper separation of concerns
  - [ ] Dependency injection used
  - [ ] No global state

- [ ] **Error Handling**
  - [ ] All errors handled
  - [ ] Appropriate error codes returned
  - [ ] No sensitive data in errors
  - [ ] Error messages helpful

- [ ] **Testing**
  - [ ] New tests added for new code
  - [ ] Edge cases covered
  - [ ] Tests have clear names
  - [ ] No test is skipped

- [ ] **Security**
  - [ ] Inputs validated
  - [ ] No SQL injection possible
  - [ ] No sensitive data logged
  - [ ] Authorization checks present

- [ ] **Database**
  - [ ] Indexes used appropriately
  - [ ] N+1 queries avoided
  - [ ] Transactions used correctly
  - [ ] Migrations included

- [ ] **Documentation**
  - [ ] Public APIs documented
  - [ ] Complex logic explained
  - [ ] README updated if needed
  - [ ] Changelog updated

---

*Last Updated: 2026-01-26*
*Review Cycle: Every sprint*
