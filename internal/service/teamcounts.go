package service

import (
	"SolProject/internal/model"
	"SolProject/internal/repository"
)

type TeamcountsService interface {
	GetTeamcountsById(id int64) (*model.Teamcounts, error)
}

type teamcountsService struct {
	*Service
	teamcountsRepository repository.TeamcountsRepository
}

func NewTeamcountsService(service *Service, teamcountsRepository repository.TeamcountsRepository) TeamcountsService {
	return &teamcountsService{
		Service:        service,
		teamcountsRepository: teamcountsRepository,
	}
}

func (s *teamcountsService) GetTeamcountsById(id int64) (*model.Teamcounts, error) {
	return s.teamcountsRepository.FirstById(id)
}
