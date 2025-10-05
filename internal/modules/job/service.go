package job

import (
	"context"
	"io"
	"log/slog"
	"selfier/pkg/middleware"
)

type serviceImpl struct {
	JobRepo
	TaskRepo
	ImageRepo
	TaskCreatedProducer
	Storage
	DeselfiePipeline
}

func NewService(
	jobRepo JobRepo,
	taskRepo TaskRepo,
	imageRepo ImageRepo,
	taskCreatedProducer TaskCreatedProducer,
	storage Storage,
	deselfiePipeline DeselfiePipeline,
) Service {
	return &serviceImpl{
		JobRepo:             jobRepo,
		TaskRepo:            taskRepo,
		ImageRepo:           imageRepo,
		TaskCreatedProducer: taskCreatedProducer,
		Storage:             storage,
		DeselfiePipeline:    deselfiePipeline,
	}
}

func (s *serviceImpl) CreateJob(ctx context.Context, options []map[string]any, imageReader io.Reader, imageFilename string) (*Job, error) {
	log := middleware.GetLogger(ctx)

	imageKey, err := s.Upload(ctx, imageReader, imageFilename)
	if err != nil {
		log.Error("failed to upload image", slog.String("error", err.Error()))
		return nil, err
	}

	var tasks []Task
	for _, option := range options {
		newTask := Task{
			Status:  StatusPending,
			Options: GetJSONFromMap(option),
			InputImage: Image{
				ImageType: ImageTypeInput,
				ImageKey:  imageKey,
			},
		}
		tasks = append(tasks, newTask)
	}

	newJob := Job{
		Tasks: tasks,
	}

	job, err := s.JobRepo.Create(ctx, &newJob)
	if err != nil {
		log.Error("failed to create job", slog.String("error", err.Error()))
		return nil, err
	}

	for _, task := range job.Tasks {
		if _, err := s.Send(ctx, &TaskCreatedEventData{
			TaskID: task.ID,
		}); err != nil {
			log.Error("failed to send task created event", slog.String("error", err.Error()))
			return nil, err
		}
	}

	return job, nil
}

func (s *serviceImpl) GetJobByID(ctx context.Context, id string) (*Job, error) {
	log := middleware.GetLogger(ctx)

	job, err := s.JobRepo.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get job by ID", slog.String("error", err.Error()))
		return nil, err
	}

	return job, nil
}

func (s *serviceImpl) GetJobsByUser(ctx context.Context) ([]*Job, error) {

	log := middleware.GetLogger(ctx)

	jobs, err := s.ListByUser(ctx)
	if err != nil {
		log.Error("failed to get jobs by user", slog.String("error", err.Error()))
		return nil, err
	}
	return jobs, nil
}

func (s *serviceImpl) DeleteJob(ctx context.Context, id string) error {

	log := middleware.GetLogger(ctx)

	err := s.JobRepo.Delete(ctx, id)
	if err != nil {
		log.Error("failed to delete job by ID", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *serviceImpl) GetPresignedURL(ctx context.Context, taskID string, imageID string) (string, error) {

	log := middleware.GetLogger(ctx)

	image, err := s.ImageRepo.GetByID(ctx, taskID, imageID)
	if err != nil {
		log.Error("failed to get image by ID", slog.String("error", err.Error()))
		return "", err
	}
	imageKey := image.ImageKey

	presignedURL, err := s.Storage.GetPresignedURL(ctx, imageKey)
	if err != nil {
		log.Error("failed to create presigned url", slog.String("error", err.Error()))
		return "", err
	}
	return presignedURL, nil
}

func (s *serviceImpl) UpdateTaskStatus(ctx context.Context, taskID string, status string) (*Task, error) {
	updatedTask := Task{
		Model:  Model{ID: taskID},
		Status: status,
	}

	task, err := s.TaskRepo.Update(ctx, &updatedTask)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// UploadOutputImage
// 1. Upload image to object storage -> ImageKey
// 2. Create image record with ImageTypeOutput
func (s *serviceImpl) UploadOutputImage(ctx context.Context, taskID string, imageReader io.Reader, imageFilename string) (*Image, error) {

	log := middleware.GetLogger(ctx)

	fileKey, err := s.Upload(ctx, imageReader, imageFilename)
	if err != nil {
		log.Error("failed to upload image", slog.String("error", err.Error()))
		return nil, err
	}

	newImage := Image{
		TaskID:    taskID,
		ImageType: ImageTypeOutput,
		ImageKey:  fileKey,
	}
	image, err := s.ImageRepo.Create(ctx, &newImage)
	if err != nil {
		log.Error("failed to create image", slog.String("error", err.Error()))
		return nil, err
	}

	return image, nil
}

// GetTaskByID
// for task created event
func (s *serviceImpl) GetTaskByID(ctx context.Context, taskID string) (*Task, error) {

	log := middleware.GetLogger(ctx)

	task, err := s.TaskRepo.GetByID(ctx, taskID)

	if err != nil {
		log.Error("failed to get task by ID", slog.String("error", err.Error()))
		return nil, err
	}

	return task, nil
}

func (s *serviceImpl) GenerateDeselfie(ctx context.Context, imageURL string, options map[string]any) (string, error) {
	return s.Generate(ctx, imageURL, options)
}
