package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mxmkiv/subscriptions-service/internal/domain"
	"github.com/mxmkiv/subscriptions-service/internal/repository"
)

type Service interface {
	Create(ctx context.Context, dto CreateDTO) (*domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateDTO) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]domain.Subscription, error)
	SumByPeriod(ctx context.Context, filter SumFilter) (int, error)
}

type subscriptionService struct {
	repo repository.Repository
}

func New(repo repository.Repository) Service {
	return &subscriptionService{repo: repo}
}

// DTO
type CreateDTO struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   string
	EndDate     *string
}

type UpdateDTO struct {
	ServiceName string
	Price       int
	UserID      uuid.UUID
	StartDate   string
	EndDate     *string
}

// filter struct

type ListFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
}

type SumFilter struct {
	UserID      uuid.UUID
	ServiceName *string
	StartDate   string
	EndDate     string
}

func (s *subscriptionService) Create(ctx context.Context, dto CreateDTO) (*domain.Subscription, error) {

	startDate, err := time.Parse("01-2006", dto.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed parse date: %w", err)
	}

	var endDate *time.Time
	if dto.EndDate != nil {
		parse, err := time.Parse("01-2006", *dto.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}

		endDate = &parse
	}

	repoDTO := repository.CreateDTO{
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	sub, err := s.repo.Create(ctx, repoDTO)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {

	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, dto UpdateDTO) (*domain.Subscription, error) {

	startDate, err := time.Parse("01-2006", dto.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed parse date: %w", err)
	}

	var endDate *time.Time
	if dto.EndDate != nil {
		parse, err := time.Parse("01-2006", *dto.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed parse date: %w", err)
		}

		endDate = &parse
	}

	repoDTO := repository.UpdateDTO{
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserID:      dto.UserID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	sub, err := s.repo.Update(ctx, id, repoDTO)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *subscriptionService) Delete(ctx context.Context, id uuid.UUID) error {

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil

}

func (s *subscriptionService) List(ctx context.Context, filter ListFilter) ([]domain.Subscription, error) {

	repoFilter := repository.ListFilter{
		UserID:      filter.UserID,
		ServiceName: filter.ServiceName,
	}

	subs, err := s.repo.List(ctx, repoFilter)
	if err != nil {
		return nil, err
	}

	return subs, nil

}

func (s *subscriptionService) SumByPeriod(ctx context.Context, filter SumFilter) (int, error) {

	startDate, err := time.Parse("01-2006", filter.StartDate)
	if err != nil {
		return 0, fmt.Errorf("failed parse date: %w", err)
	}

	endDate, err := time.Parse("01-2006", filter.EndDate)
	if err != nil {
		return 0, fmt.Errorf("failed parse date: %w", err)
	}

	repoFilter := repository.SumFilter{
		UserID:      filter.UserID,
		ServiceName: filter.ServiceName,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	sum, err := s.repo.SumByPeriod(ctx, repoFilter)
	if err != nil {
		return 0, err
	}

	return sum, nil

}
