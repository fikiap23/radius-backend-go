package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	userdomain "github.com/radius/radius-backend/internal/users/domain"
	"github.com/radius/radius-backend/internal/workspaces/application/dto"
	wsdomain "github.com/radius/radius-backend/internal/workspaces/domain"
	"go.uber.org/zap"
)

type WorkspaceService struct {
	workspaceRepo wsdomain.WorkspaceRepository
	memberRepo    wsdomain.WorkspaceMemberRepository
	userRepo      userdomain.UserRepository
	runWorkspaces wsdomain.RunWorkspacesInTransactionFunc
	logger        *zap.Logger
}

func NewWorkspaceService(
	workspaceRepo wsdomain.WorkspaceRepository,
	memberRepo wsdomain.WorkspaceMemberRepository,
	userRepo userdomain.UserRepository,
	runWorkspaces wsdomain.RunWorkspacesInTransactionFunc,
	logger *zap.Logger,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
		userRepo:      userRepo,
		runWorkspaces: runWorkspaces,
		logger:        logger,
	}
}

func (s *WorkspaceService) HandleListWorkspaces(ctx context.Context, userID string) ([]dto.WorkspaceResponse, error) {
	rows, err := s.workspaceRepo.FindMany(ctx, wsdomain.WorkspaceQuery{
		Select: wsdomain.WorkspaceFieldsProfile,
		Filter: wsdomain.WorkspaceFilter{UserID: &userID},
	})
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	return dto.MapWorkspaces(rows), nil
}

func (s *WorkspaceService) HandleCreateWorkspace(
	ctx context.Context,
	userID string,
	name, slug string,
) (*dto.WorkspaceResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID, userdomain.FieldsProfile)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return nil, userdomain.ErrUserNotFound
		}
		return nil, fmt.Errorf("load user: %w", err)
	}

	workspace := &wsdomain.Workspace{
		ID:   uuid.NewString(),
		Name: strings.TrimSpace(name),
		Slug: dto.NormalizeSlug(slug),
	}
	owner := &wsdomain.WorkspaceMember{
		ID:          uuid.NewString(),
		WorkspaceID: workspace.ID,
		UserID:      &userID,
		Name:        user.Name,
		Email:       user.Email,
		Role:        wsdomain.MemberRoleOwner,
		Status:      wsdomain.MemberStatusActive,
	}

	if err := s.runWorkspaces(ctx, func(ctx context.Context, repos wsdomain.WorkspacesRepositories) error {
		if err := repos.Workspaces.Create(ctx, workspace); err != nil {
			return err
		}
		return repos.WorkspaceMembers.Create(ctx, owner)
	}); err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceSlugAlreadyExists) {
			return nil, wsdomain.ErrWorkspaceSlugAlreadyExists
		}
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	out := dto.MapWorkspace(workspace)
	return &out, nil
}

func (s *WorkspaceService) HandleUpdateWorkspace(
	ctx context.Context,
	userID, workspaceID string,
	data wsdomain.WorkspaceUpdateData,
) (*dto.WorkspaceResponse, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	if err := s.workspaceRepo.UpdateByID(ctx, workspaceID, data); err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceNotFound) {
			return nil, wsdomain.ErrWorkspaceNotFound
		}
		if errors.Is(err, wsdomain.ErrWorkspaceSlugAlreadyExists) {
			return nil, wsdomain.ErrWorkspaceSlugAlreadyExists
		}
		return nil, fmt.Errorf("update workspace: %w", err)
	}

	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID, wsdomain.WorkspaceFieldsProfile)
	if err != nil {
		return nil, fmt.Errorf("reload workspace: %w", err)
	}
	out := dto.MapWorkspace(workspace)
	return &out, nil
}

func (s *WorkspaceService) HandleListMembers(
	ctx context.Context,
	userID, workspaceID string,
) ([]dto.MemberResponse, error) {
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	rows, err := s.memberRepo.FindMany(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsProfile,
		Filter: wsdomain.WorkspaceMemberFilter{WorkspaceID: &workspaceID},
	})
	if err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}
	return dto.MapMembers(rows), nil
}

func (s *WorkspaceService) HandleInviteMember(
	ctx context.Context,
	userID, workspaceID, email, role string,
) (*dto.MemberResponse, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	memberRole, err := parseInvitableRole(role)
	if err != nil {
		return nil, err
	}

	email = strings.TrimSpace(strings.ToLower(email))

	_, err = s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &workspaceID,
			Email:       &email,
		},
	})
	if err == nil {
		return nil, wsdomain.ErrWorkspaceMemberAlreadyExists
	}
	if !errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
		return nil, fmt.Errorf("check existing member: %w", err)
	}

	var linkedUserID *string
	linkedUser, err := s.userRepo.FindOne(ctx, userdomain.Query{
		Select: userdomain.FieldsProfile,
		Filter: userdomain.Filter{Email: &email},
	})
	if err == nil {
		linkedUserID = &linkedUser.ID
	} else if !errors.Is(err, userdomain.ErrUserNotFound) {
		return nil, fmt.Errorf("lookup user by email: %w", err)
	}

	member := &wsdomain.WorkspaceMember{
		ID:          uuid.NewString(),
		WorkspaceID: workspaceID,
		UserID:      linkedUserID,
		Name:        nameFromEmail(email),
		Email:       email,
		Role:        memberRole,
		Status:      wsdomain.MemberStatusPending,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberAlreadyExists) {
			return nil, wsdomain.ErrWorkspaceMemberAlreadyExists
		}
		return nil, fmt.Errorf("invite workspace member: %w", err)
	}

	out := dto.MapMember(member)
	return &out, nil
}

func (s *WorkspaceService) HandleUpdateMember(
	ctx context.Context,
	userID, workspaceID, memberID, role string,
) (*dto.OkResponse, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	memberRole, err := parseMemberRole(role)
	if err != nil {
		return nil, err
	}

	member, err := s.memberRepo.FindByID(ctx, memberID, wsdomain.WorkspaceMemberFieldsProfile)
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return nil, wsdomain.ErrWorkspaceMemberNotFound
		}
		return nil, fmt.Errorf("load member: %w", err)
	}
	if member.WorkspaceID != workspaceID {
		return nil, wsdomain.ErrWorkspaceMemberNotFound
	}
	if member.Role == wsdomain.MemberRoleOwner && memberRole != wsdomain.MemberRoleOwner {
		return nil, wsdomain.ErrWorkspaceForbidden
	}

	if err := s.memberRepo.UpdateByID(ctx, memberID, wsdomain.WorkspaceMemberUpdateData{
		Role: &memberRole,
	}); err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return nil, wsdomain.ErrWorkspaceMemberNotFound
		}
		return nil, fmt.Errorf("update workspace member: %w", err)
	}

	return &dto.OkResponse{OK: true}, nil
}

func (s *WorkspaceService) HandleDeleteMember(
	ctx context.Context,
	userID, workspaceID, memberID string,
) (*dto.OkResponse, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	member, err := s.memberRepo.FindByID(ctx, memberID, wsdomain.WorkspaceMemberFieldsProfile)
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return nil, wsdomain.ErrWorkspaceMemberNotFound
		}
		return nil, fmt.Errorf("load member: %w", err)
	}
	if member.WorkspaceID != workspaceID {
		return nil, wsdomain.ErrWorkspaceMemberNotFound
	}
	if member.Role == wsdomain.MemberRoleOwner {
		return nil, wsdomain.ErrWorkspaceForbidden
	}

	if err := s.memberRepo.DeleteByID(ctx, memberID); err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return nil, wsdomain.ErrWorkspaceMemberNotFound
		}
		return nil, fmt.Errorf("delete workspace member: %w", err)
	}

	return &dto.OkResponse{OK: true}, nil
}

func (s *WorkspaceService) requireWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	_, err := s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsExists,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &workspaceID,
			UserID:      &userID,
			Status:      statusPtr(wsdomain.MemberStatusActive),
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return wsdomain.ErrWorkspaceForbidden
		}
		return fmt.Errorf("check workspace membership: %w", err)
	}
	return nil
}

func (s *WorkspaceService) requireWorkspaceAdmin(ctx context.Context, workspaceID, userID string) error {
	member, err := s.memberRepo.FindOne(ctx, wsdomain.WorkspaceMemberQuery{
		Select: wsdomain.WorkspaceMemberFieldsProfile,
		Filter: wsdomain.WorkspaceMemberFilter{
			WorkspaceID: &workspaceID,
			UserID:      &userID,
			Status:      statusPtr(wsdomain.MemberStatusActive),
		},
	})
	if err != nil {
		if errors.Is(err, wsdomain.ErrWorkspaceMemberNotFound) {
			return wsdomain.ErrWorkspaceForbidden
		}
		return fmt.Errorf("check workspace admin: %w", err)
	}
	if member.Role != wsdomain.MemberRoleOwner && member.Role != wsdomain.MemberRoleAdmin {
		return wsdomain.ErrWorkspaceForbidden
	}
	return nil
}

func parseInvitableRole(role string) (wsdomain.MemberRole, error) {
	switch wsdomain.MemberRole(strings.TrimSpace(role)) {
	case wsdomain.MemberRoleAdmin, wsdomain.MemberRoleMember, wsdomain.MemberRoleViewer:
		return wsdomain.MemberRole(role), nil
	default:
		return "", wsdomain.ErrInvalidMemberRole
	}
}

func parseMemberRole(role string) (wsdomain.MemberRole, error) {
	switch wsdomain.MemberRole(strings.TrimSpace(role)) {
	case wsdomain.MemberRoleOwner, wsdomain.MemberRoleAdmin, wsdomain.MemberRoleMember, wsdomain.MemberRoleViewer:
		return wsdomain.MemberRole(role), nil
	default:
		return "", wsdomain.ErrInvalidMemberRole
	}
}

func nameFromEmail(email string) string {
	local, _, _ := strings.Cut(email, "@")
	local = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(local)
	return strings.TrimSpace(local)
}

func statusPtr(s wsdomain.MemberStatus) *wsdomain.MemberStatus {
	return &s
}
