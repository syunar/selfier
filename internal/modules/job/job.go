// Package job contains the ports or interefaces for the job module
package job

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/inngest/inngestgo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ----------
// Ports
// ----------

type HTTPHandler interface {
	CreateJob(ctx context.Context, input *CreateInput) (*JobOutput, error)
	GetJobByID(ctx context.Context, input *GetByIDInput) (*JobOutput, error)
	GetJobsByUser(ctx context.Context, input *struct{}) (*JobsOutput, error)
	Delete(ctx context.Context, input *DeleteInput) (*struct{}, error)
	GetPresignedURL(ctx context.Context, input *GetPresignedURLInput) (*GetPresignedURLOutput, error)
}

type EventHandler interface {
	TaskCreated(ctx context.Context, input inngestgo.Input[TaskCreatedEventData]) (any, error)
}

type Service interface {
	CreateJob(ctx context.Context, options []map[string]any, imageReader io.Reader, imageFilename string) (*Job, error)
	GetJobByID(ctx context.Context, id string) (*Job, error)
	GetJobsByUser(ctx context.Context) ([]*Job, error)
	DeleteJob(ctx context.Context, id string) error
	GetPresignedURL(ctx context.Context, taskID string, imageID string) (string, error)
	UpdateTaskStatus(ctx context.Context, taskID string, status string) (*Task, error)
	UploadOutputImage(ctx context.Context, taskID string, imageReader io.Reader, imageFilename string) (*Image, error)
	GetTaskByID(ctx context.Context, taskID string) (*Task, error)
	GenerateDeselfie(ctx context.Context, imageURL string, options map[string]any) (string, error)
}

type JobRepo interface {
	Create(ctx context.Context, job *Job) (*Job, error)
	GetByID(ctx context.Context, id string) (*Job, error)
	Update(ctx context.Context, job *Job) (*Job, error)
	Delete(ctx context.Context, id string) error
	ListByUser(ctx context.Context) ([]*Job, error)
}

type TaskRepo interface {
	Create(ctx context.Context, task *Task) (*Task, error)
	GetByID(ctx context.Context, id string) (*Task, error)
	Update(ctx context.Context, task *Task) (*Task, error)
	Delete(ctx context.Context, id string) error
	ListByJob(ctx context.Context, jobID string) ([]*Task, error)
}

type ImageRepo interface {
	Create(ctx context.Context, image *Image) (*Image, error)
	GetByID(ctx context.Context, taskID string, imageID string) (*Image, error)
	Update(ctx context.Context, image *Image) (*Image, error)
	Delete(ctx context.Context, id string) error
	ListByTask(ctx context.Context, taskID string) ([]*Image, error)
}

type Storage interface {
	Upload(ctx context.Context, file io.Reader, fileName string) (string, error)
	GetPresignedURL(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type DeselfiePipeline interface {
	Generate(ctx context.Context, imageURL string, options map[string]any) (string, error)
}

type TaskCreatedProducer interface {
	Send(ctx context.Context, event *TaskCreatedEventData) (string, error)
}

// ---------
// Entites
// --------

type Model struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	// DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}

type Job struct {
	Model
	Tasks []Task `json:"tasks" gorm:"foreignKey:JobID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Task struct {
	Model
	JobID       string         `json:"job_id"`
	Status      string         `json:"status"`
	Options     datatypes.JSON `json:"options" gorm:"type:jsonb"`
	InputImage  Image          `json:"input_image" gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	OutputImage *Image         `json:"output_image" gorm:"foreignKey:TaskID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func GetJSONFromMap(dataMap map[string]any) datatypes.JSON {
	data, _ := json.Marshal(dataMap)
	return datatypes.JSON(data)
}

func GetMapFromJSON(dataJSON datatypes.JSON) map[string]any {
	var result map[string]any
	_ = json.Unmarshal([]byte(dataJSON), &result)
	return result
}

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type Image struct {
	Model
	TaskID    string `json:"task_id"`
	ImageType string `json:"-"`
	ImageKey  string `json:"image_key"`
	// FileSize  int64  `json:"file_size"`
	// MimeType  string `json:"mime_type"`
}

const (
	ImageTypeInput  = "input"
	ImageTypeOutput = "output"
)

// ----------
// DTOs
// ----------

type CreateBody struct {
	Image   huma.FormFile `form:"image" required:"true" doc:"The image file to upload."`
	Options string        `form:"options" doc:"A JSON string representing an array of option objects." example:"[{\"prompt\": \"a photo of a person\"}, {\"prompt\": \"a photo of a cat\"}]" required:"true"`
}

type CreateInput struct {
	RawBody huma.MultipartFormFiles[CreateBody]
}

type GetByIDInput struct {
	ID string `path:"id" binding:"required"`
}

type JobOutput struct {
	Body *Job `json:"body"`
}

type JobsOutput struct {
	Body []*Job `json:"body"`
}

type DeleteInput struct {
	ID string `path:"id" binding:"required"`
}

type GetPresignedURLInput struct {
	ID      string `path:"image_id" required:"true" doc:"Image ID" example:"f6b57be8-aa17-4e46-8cca-396cb7f977e9"`
	JobID   string `path:"job_id"   required:"true" doc:"Job ID"   example:"052ef2f5-28dd-44e6-8345-e8fc11235d1c"`
	TaskID  string `path:"task_id" required:"true" doc:"Task ID" example:"052ef2f5-28dd-44e6-8345-e8fc11235d1c"`
	ImageID string `path:"image_id" required:"true" doc:"Image ID" example:"f6b57be8-aa17-4e46-8cca-396cb7f977e9"`
}

type GetPresignedURLOutputBody struct {
	URL string `json:"url"`
}

type GetPresignedURLOutput struct {
	Body *GetPresignedURLOutputBody `json:"body"`
}

// ----------
// Events
// ----------

const TaskCreatedEventTopic = "api/task.created"

type TaskCreatedEventData struct {
	TaskID string `json:"task_id"`
}

// ----------
// Errors
// ----------

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicatedKey  = errors.New("record conflict, probably non-unique ID")
	ErrInternal       = errors.New("internal data access error")
	ErrUnreadableFile = errors.New("failed to read file")
)

func mapGormError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicatedKey
	}

	return ErrInternal
}
