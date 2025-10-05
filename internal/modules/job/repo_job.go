package job

import (
	"context"

	"gorm.io/gorm"
)

type jobRepoGorm struct {
	db *gorm.DB
}

func NewJobRepo(db *gorm.DB) JobRepo {
	err := db.AutoMigrate(&Job{}, &Task{}, &Image{})
	if err != nil {
		panic(err)
	}
	return &jobRepoGorm{db: db}
}

func (r *jobRepoGorm) Create(ctx context.Context, job *Job) (*Job, error) {
	if err := r.db.WithContext(ctx).
		Create(job).Error; err != nil {
		return nil, mapGormError(err)
	}
	return job, nil
}

func (r *jobRepoGorm) GetByID(ctx context.Context, id string) (*Job, error) {
	var job Job
	if err := r.db.WithContext(ctx).
		Model(&Job{}).
		Preload("Tasks").
		Preload("Tasks.InputImage", "image_type = 'input'").
		Preload("Tasks.OutputImage", "image_type = 'output'").
		Where("id = ?", id).
		First(&job).Error; err != nil {
		return nil, mapGormError(err)
	}
	return &job, nil
}

func (r *jobRepoGorm) Update(ctx context.Context, job *Job) (*Job, error) {
	if err := r.db.WithContext(ctx).
		Save(job).Error; err != nil {
		return nil, err
	}
	return job, nil
}

func (r *jobRepoGorm) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&Job{}).Error; err != nil {
		return mapGormError(err)
	}
	return nil
}

func (r *jobRepoGorm) ListByUser(ctx context.Context) ([]*Job, error) {
	var jobs []*Job
	if err := r.db.WithContext(ctx).Model(&Job{}).
		Preload("Tasks").
		Preload("Tasks.InputImage", "image_type = 'input'").
		Preload("Tasks.OutputImage", "image_type = 'output'").
		Find(&jobs).Error; err != nil {
		return nil, mapGormError(err)
	}
	return jobs, nil
}
