package tasks

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	"github.com/radius/radius-backend/ent"
	"github.com/radius/radius-backend/internal/module"
	projectpostgres "github.com/radius/radius-backend/internal/projects/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/shared/middleware"
	"github.com/radius/radius-backend/internal/tasks/application/services"
	"github.com/radius/radius-backend/internal/tasks/domain"
	"github.com/radius/radius-backend/internal/tasks/infrastructure/db/postgres"
	"github.com/radius/radius-backend/internal/tasks/interface/api/rest"
	wspostgres "github.com/radius/radius-backend/internal/workspaces/infrastructure/db/postgres"
)

type Module struct {
	taskSvc       *services.TaskService
	activitySvc   *services.ActivityService
	attachmentSvc *services.AttachmentService
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "tasks"
}

func (m *Module) wire(deps module.Dependencies) {
	if m.taskSvc != nil {
		return
	}

	taskRepo := postgres.NewTaskRepository(deps.Ent)
	subtaskRepo := postgres.NewTaskSubtaskRepository(deps.Ent)
	checklistRepo := postgres.NewTaskChecklistItemRepository(deps.Ent)
	activityRepo := postgres.NewTaskActivityLogRepository(deps.Ent)
	attachmentRepo := postgres.NewTaskAttachmentRepository(deps.Ent)
	projectRepo := projectpostgres.NewProjectRepository(deps.Ent)
	columnRepo := projectpostgres.NewBoardColumnRepository(deps.Ent)
	memberRepo := wspostgres.NewWorkspaceMemberRepository(deps.Ent)

	runTasksTx := func(ctx context.Context, fn func(context.Context, domain.TasksRepositories) error) error {
		return deps.RunInTransaction(ctx, func(ctx context.Context, tx *ent.Client) error {
			return fn(ctx, domain.TasksRepositories{
				Tasks:          postgres.NewTaskRepository(tx),
				Subtasks:       postgres.NewTaskSubtaskRepository(tx),
				ChecklistItems: postgres.NewTaskChecklistItemRepository(tx),
				Attachments:    postgres.NewTaskAttachmentRepository(tx),
				ActivityLogs:   postgres.NewTaskActivityLogRepository(tx),
				Projects:       projectpostgres.NewProjectRepository(tx),
			})
		})
	}

	m.taskSvc = services.NewTaskService(
		taskRepo,
		subtaskRepo,
		checklistRepo,
		activityRepo,
		attachmentRepo,
		projectRepo,
		columnRepo,
		memberRepo,
		runTasksTx,
		deps.ObjectStorage,
	)
	m.activitySvc = services.NewActivityService(taskRepo, activityRepo, m.taskSvc)
	m.attachmentSvc = services.NewAttachmentService(
		taskRepo,
		attachmentRepo,
		activityRepo,
		deps.ObjectStorage,
		m.taskSvc,
	)
}

func (m *Module) RegisterHTTP(_ *echo.Echo, api huma.API, deps module.Dependencies, auth *middleware.AuthMiddleware) {
	m.wire(deps)
	rest.RegisterTasks(api, m.taskSvc, auth, deps.Logger)
	rest.RegisterActivities(api, m.activitySvc, auth, deps.Logger)
	rest.RegisterAttachments(api, m.attachmentSvc, auth, deps.Logger)
}

func (m *Module) StartMessaging(_ context.Context, _ module.Dependencies) (func(), error) {
	return nil, nil
}

var _ module.BoundedContext = (*Module)(nil)
