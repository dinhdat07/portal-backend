package services

import (
	"context"
	"encoding/json"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	appLogger "log"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLogService struct {
	repo *repositories.AuditLogRepository
}

func NewAuditLogService(repo *repositories.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

func (s *AuditLogService) Log(ctx context.Context, meta *domain.AuditMeta, action enum.ActionName, actor *domain.AuditUser, target *domain.AuditUser) error {
	log := &models.AuditLog{Action: action}

	if actor != nil {
		log.ActorUserID = &actor.ID
		log.ActorUsername = &actor.Username
		log.ActorEmail = &actor.Email
		log.ActorRole = &actor.RoleCode
	}

	if target != nil {
		log.TargetUserID = &target.ID
		log.TargetUsername = &target.Username
		log.TargetEmail = &target.Email
		log.TargetRole = &target.RoleCode
	}

	if meta != nil {
		log.IPAddress = &meta.IPAddress
		log.UserAgent = &meta.UserAgent
	}

	err := s.repo.Create(ctx, log)
	if err != nil {
		appLogger.Println(err)
	}
	return err
}

func (s *AuditLogService) List(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// validate time range
	if filter.From != nil && filter.To != nil {
		if filter.From.After(*filter.To) {
			return nil, 0, ErrInvalidTimeRange
		}
	}

	logs, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (svc *AuditLogService) LogWithMetadata(ctx context.Context, meta *domain.AuditMeta, action enum.ActionName, actor *domain.AuditUser, target *domain.AuditUser, data map[string]any) error {
	var metadata *datatypes.JSON
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		m := datatypes.JSON(b)
		metadata = &m
	}

	log := &models.AuditLog{
		Action: action,
	}

	if actor != nil {
		log.ActorUserID = &actor.ID
		log.ActorUsername = &actor.Username
		log.ActorEmail = &actor.Email
		log.ActorRole = &actor.RoleCode
	}

	if target != nil {
		log.TargetUserID = &target.ID
		log.TargetUsername = &target.Username
		log.TargetEmail = &target.Email
		log.TargetRole = &target.RoleCode
	}

	if meta != nil {
		log.IPAddress = &meta.IPAddress
		log.UserAgent = &meta.UserAgent
	}

	log.Metadata = metadata

	err := svc.repo.Create(ctx, log)

	if err != nil {
		appLogger.Println(err)
	}
	return err
}

func (s *AuditLogService) WithTx(tx *gorm.DB) *AuditLogService {
	return &AuditLogService{
		repo: s.repo.WithTx(tx),
	}
}
