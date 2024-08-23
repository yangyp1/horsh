package server

import (
	"SolProject/internal/service"
	"SolProject/pkg/log"
	"context"
	"fmt"
)

type Job struct {
	log *log.Logger
	userservice service.UserService
}

func NewJob(
	log *log.Logger,
	userservice service.UserService,
) *Job {
	return &Job{
		log: log,
		userservice: userservice,
	}
}
func (j *Job) Start(ctx context.Context) error {
    go j.userservice.TASK(ctx)
	return nil
}
func (j *Job) Stop(ctx context.Context) error {
	fmt.Printf("Job Stop")
	return nil
}
