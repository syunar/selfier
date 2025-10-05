package job

import (
	"context"

	"gorm.io/gorm"
)

type imageRepoGorm struct {
	db *gorm.DB
}

func NewImageRepo(db *gorm.DB) ImageRepo {
	err := db.AutoMigrate(&Image{})
	if err != nil {
		panic(err)
	}
	return &imageRepoGorm{db: db}
}

func (r *imageRepoGorm) Create(ctx context.Context, image *Image) (*Image, error) {
	if err := r.db.WithContext(ctx).
		Create(image).Error; err != nil {
		return nil, mapGormError(err)
	}
	return image, nil
}

func (r *imageRepoGorm) GetByID(ctx context.Context, taskID string, imageID string) (*Image, error) {
	var image Image
	if err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Where("id = ?", imageID).
		First(&image).Error; err != nil {
		return nil, mapGormError(err)
	}
	return &image, nil
}

func (r *imageRepoGorm) Update(ctx context.Context, image *Image) (*Image, error) {
	if err := r.db.WithContext(ctx).
		Save(image).Error; err != nil {
		return nil, mapGormError(err)
	}
	return image, nil
}

func (r *imageRepoGorm) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&Image{}).Error; err != nil {
		return mapGormError(err)
	}
	return nil
}

func (r *imageRepoGorm) ListByTask(ctx context.Context, taskID string) ([]*Image, error) {
	var images []*Image
	if err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Find(&images).Error; err != nil {
		return nil, mapGormError(err)
	}
	return images, nil
}
