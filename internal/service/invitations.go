package service

import (
	"SolProject/internal/model"
	"SolProject/internal/repository"
)

type InvitationsService interface {
	GetInvitationsById(id int64) (*model.Invitations, error)
}

type invitationsService struct {
	*Service
	invitationsRepository repository.InvitationsRepository
}

func NewInvitationsService(service *Service, invitationsRepository repository.InvitationsRepository) InvitationsService {
	return &invitationsService{
		Service:        service,
		invitationsRepository: invitationsRepository,
	}
}

func (s *invitationsService) GetInvitationsById(id int64) (*model.Invitations, error) {
	return s.invitationsRepository.FirstById(id)
}
