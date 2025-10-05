package job

import (
	"context"

	"gorm.io/gorm"
)

type taskRepoGorm struct {
	db *gorm.DB
}

func NewTaskRepo(db *gorm.DB) TaskRepo {
	err := db.AutoMigrate(&Task{}, &Image{})
	if err != nil {
		panic(err)
	}
	return &taskRepoGorm{db: db}
}

func (r *taskRepoGorm) Create(ctx context.Context, task *Task) (*Task, error) {
	if err := r.db.WithContext(ctx).
		Create(task).Error; err != nil {
		return nil, mapGormError(err)
	}
	return task, nil
}

func (r *taskRepoGorm) GetByID(ctx context.Context, id string) (*Task, error) {
	var task Task
	if err := r.db.WithContext(ctx).
		Model(&Task{}).
		Preload("InputImage", "image_type = 'input'").
		Preload("OutputImage", "image_type = 'output'").
		Where("id = ?", id).
		First(&task).Error; err != nil {
		return nil, mapGormError(err)
	}
	return &task, nil
}

func (r *taskRepoGorm) Update(ctx context.Context, task *Task) (*Task, error) {
	if err := r.db.WithContext(ctx).
		Model(&Task{}).
		Where("id = ?", task.ID).
		Update("status", task.Status).Error; err != nil {
		return nil, mapGormError(err)
	}
	return task, nil
}

func (r *taskRepoGorm) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&Task{}).Error; err != nil {
		return mapGormError(err)
	}
	return nil
}

func (r *taskRepoGorm) ListByJob(ctx context.Context, jobID string) ([]*Task, error) {
	var tasks []*Task
	if err := r.db.WithContext(ctx).Model(&Task{}).
		Preload("InputImage", "image_type = 'input'").
		Preload("OutputImage", "image_type = 'output'").
		Where("job_id = ?", jobID).
		Find(&tasks).Error; err != nil {
		return nil, mapGormError(err)
	}
	return tasks, nil
}
