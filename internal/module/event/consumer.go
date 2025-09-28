package event

import (
	"selfier/internal/module/aideselfie"
	"selfier/internal/module/job"
	"selfier/internal/module/notification"
)

type eventConsumerImpl struct {
	jobService          job.JobService
	notificationService notification.NotificationService
	aideselfieService   aideselfie.AIdeselfieService
}

func NewEventConsumer(jobService job.JobService, notificationService notification.NotificationService, aideselfieService aideselfie.AIdeselfieService) EventConsumer {
	return &eventConsumerImpl{
		jobService:          jobService,
		notificationService: notificationService,
		aideselfieService:   aideselfieService,
	}
}

func (e *eventConsumerImpl) JobCreatedEvent() {
	// e.jobService.GetJobByID()
	// e.jobService.GetPresignedURL()
	// e.jobService.UpdateJobStatus()
	// e.aideselfieService.GenerateMultiple()
	// e.jobService.UploadImage()
	// e.jobService.UpdateJobStatus()
	// e.notificationService.SendNotification()
}
