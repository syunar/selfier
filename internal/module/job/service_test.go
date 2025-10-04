package job

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
)

type mockJobRepo struct {
	mock.Mock
}

//nolint:exhaustruct
func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{}
}

func (m *mockJobRepo) CreateJob(
	ctx context.Context,
	id string,
	jobType string,
	modelConfig map[string]interface{},
	status string,
) (*JobModel, error) {
	args := m.Called(ctx, id, jobType, modelConfig, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JobModel), args.Error(1)
}

func (m *mockJobRepo) GetJobs(ctx context.Context) ([]*JobModel, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*JobModel), args.Error(1)
}

func (m *mockJobRepo) GetJobByID(ctx context.Context, jobID string) (*JobModel, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JobModel), args.Error(1)
}

func (m *mockJobRepo) DeleteJobByID(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockJobRepo) UpdateJobStatus(
	ctx context.Context,
	id string,
	status string,
) (*JobModel, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JobModel), args.Error(1)
}

type mockJobImageRepo struct {
	mock.Mock
}

//nolint:exhaustruct
func newMockJobImageRepo() *mockJobImageRepo {
	return &mockJobImageRepo{}
}

func (m *mockJobImageRepo) GetImages(ctx context.Context, jobID string) ([]*JobImageModel, error) {
	args := m.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*JobImageModel), args.Error(1)
}

func (m *mockJobImageRepo) GetImage(
	ctx context.Context,
	jobID string,
	imageID string,
) (*JobImageModel, error) {
	args := m.Called(ctx, jobID, imageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JobImageModel), args.Error(1)
}

func (m *mockJobImageRepo) CreateImage(
	ctx context.Context,
	imageID string,
	jobID string,
	imageKey string,
	imageType string,
) (*JobImageModel, error) {
	args := m.Called(ctx, imageID, jobID, imageKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JobImageModel), args.Error(1)
}

type mockJobObjectStorage struct {
	mock.Mock
}

//nolint:exhaustruct
func newMockJobObjectStorage() *mockJobObjectStorage {
	return &mockJobObjectStorage{}
}

func (m *mockJobObjectStorage) GetPresignedURL(
	ctx context.Context,
	objectKey string,
) (string, error) {
	args := m.Called(ctx, objectKey)
	return args.String(0), args.Error(1)
}

func (m *mockJobObjectStorage) UploadImage(
	ctx context.Context,
	fileName string,
	file io.Reader,
) (string, error) {
	args := m.Called(ctx, fileName, file)
	return args.String(0), args.Error(1)
}

func (m *mockJobObjectStorage) DeleteImage(ctx context.Context, imageKey string) error {
	args := m.Called(ctx, imageKey)
	return args.Error(0)
}

func TestJobService_CreateJob(t *testing.T) {

	testCases := []struct {
		name          string
		jobType       string
		modelConfig   map[string]interface{}
		imageReader   io.Reader
		imageFilename string
		expected      *Job
		expectedErr   error
		setupFunc     func(t *testing.T, mockRepo *mockJobRepo, mockJobObject *mockJobObjectStorage, mockJobImageRepo *mockJobImageRepo)
	}{
		{
			name:          "success",
			jobType:       JobTypeDeselfie,
			modelConfig:   map[string]interface{}{"prompt": "a photo of a person"},
			imageReader:   nil,
			imageFilename: "",
			expected: &Job{
				ID:   "123",
				Type: JobTypeDeselfie,
				ModelConfig: map[string]interface{}{
					"prompt": "a photo of a person",
				},
				Status:    StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedErr: nil,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobObject *mockJobObjectStorage, mockJobImageRepo *mockJobImageRepo) {
				mockJobObject.On("UploadImage", mock.Anything, mock.Anything, mock.Anything).
					Return("jobs/123/image.jpg", nil)

				mockJobImageRepo.On("CreateImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&JobImageModel{ //nolint:exhaustruct
						ID:       "abc",
						JobID:    "123",
						ImageKey: "jobs/123/image.jpg",
						Type:     ImageTypeInput,
					}, nil)

				mockRepo.On("CreateJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, StatusPending).
					Return(&JobModel{ //nolint:exhaustruct
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "a photo of a person"}`)),
					}, nil)
			},
		},
		{
			name: "failed_getjobfrommodel",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobObject *mockJobObjectStorage, mockJobImageRepo *mockJobImageRepo) {
				mockJobObject.On("UploadImage", mock.Anything, mock.Anything, mock.Anything).
					Return("jobs/123/image.jpg", nil)

				mockJobImageRepo.On("CreateImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&JobImageModel{ //nolint:exhaustruct
						ID:       "abc",
						JobID:    "123",
						ImageKey: "jobs/123/image.jpg",
						Type:     ImageTypeInput,
					}, nil)
				mockRepo.On("CreateJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, StatusPending).
					Return(&JobModel{ //nolint:exhaustruct
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "missing closing brace"`)),
					}, nil)
			},
			expectedErr:   ErrInternal,
			expected:      nil,
			jobType:       JobTypeDeselfie,
			modelConfig:   map[string]interface{}{"prompt": "a photo of a person"},
			imageReader:   nil,
			imageFilename: "",
		},
		{
			name: "failed_jobrepo",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobObject *mockJobObjectStorage, mockJobImageRepo *mockJobImageRepo) {
				mockJobObject.On("UploadImage", mock.Anything, mock.Anything, mock.Anything).
					Return("jobs/123/image.jpg", nil)
				mockJobImageRepo.On("CreateImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&JobImageModel{ //nolint:exhaustruct
						ID:       "abc",
						JobID:    "123",
						ImageKey: "jobs/123/image.jpg",
						Type:     ImageTypeInput,
					}, nil)

				mockRepo.On("CreateJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, StatusPending).
					Return(nil, assert.AnError)
			},
			expectedErr:   ErrInternal,
			expected:      nil,
			jobType:       JobTypeDeselfie,
			modelConfig:   map[string]interface{}{"prompt": "a photo of a person"},
			imageReader:   nil,
			imageFilename: "",
		},
		{
			name: "failed_errconflict",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobObject *mockJobObjectStorage, mockJobImageRepo *mockJobImageRepo) {
				mockJobObject.On("UploadImage", mock.Anything, mock.Anything, mock.Anything).
					Return("jobs/123/image.jpg", nil)
				mockJobImageRepo.On("CreateImage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&JobImageModel{ //nolint:exhaustruct
						ID:       "abc",
						JobID:    "123",
						ImageKey: "jobs/123/image.jpg",
						Type:     ImageTypeInput,
					}, nil)

				mockRepo.On("CreateJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, StatusPending).
					Return(nil, ErrConflict)
			},
			expectedErr:   ErrConflict,
			expected:      nil,
			jobType:       JobTypeDeselfie,
			modelConfig:   map[string]interface{}{"prompt": "a photo of a person"},
			imageReader:   nil,
			imageFilename: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := newMockJobRepo()
			mockJobImageRepo := newMockJobImageRepo()
			mockJobObject := newMockJobObjectStorage()
			tc.setupFunc(t, mockRepo, mockJobObject, mockJobImageRepo)
			jobService := NewJobService(mockRepo, mockJobImageRepo, mockJobObject, nil)
			got, err := jobService.CreateJob(
				context.Background(),
				tc.jobType,
				tc.modelConfig,
				tc.imageReader,
				tc.imageFilename,
			)
			assert.Equal(t, tc.expectedErr, err)
			if tc.expected != nil {
				assert.Equal(t, tc.expected.ID, got.ID)
				assert.Equal(t, tc.expected.Type, got.Type)
				assert.Equal(t, tc.expected.ModelConfig, got.ModelConfig)
				assert.Equal(t, tc.expected.Status, got.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJobService_GetJobs(t *testing.T) {

	testCases := []struct {
		name        string
		expected    []*Job
		expectedErr error
		setupFunc   func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo)
	}{
		{
			name: "success",
			expected: []*Job{
				{
					ID:   "123",
					Type: JobTypeDeselfie,
					ModelConfig: map[string]interface{}{
						"prompt": "a photo of a person",
					},
					Status:    StatusPending,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			expectedErr: nil,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobs", mock.Anything).Return([]*JobModel{ //nolint:exhaustruct
					{
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "a photo of a person"}`)),
					},
				}, nil)
			},
		},
		{
			name: "failed_jobrepo",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobs", mock.Anything).Return(nil, assert.AnError)
			},
			expectedErr: ErrInternal,
			expected:    nil,
		},
		{
			name: "failed_getjobfrommodel",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobs", mock.Anything).Return([]*JobModel{ //nolint:exhaustruct
					{
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "missing closing brace"`)),
					},
				}, nil)
			},
			expectedErr: ErrInternal,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := newMockJobRepo()
			mockJobImageRepo := newMockJobImageRepo()
			tc.setupFunc(t, mockRepo, mockJobImageRepo)
			jobService := NewJobService(mockRepo, mockJobImageRepo, nil, nil)
			got, err := jobService.GetJobs(t.Context())
			assert.Equal(t, tc.expectedErr, err)
			if tc.expected != nil {
				assert.Equal(t, tc.expected[0].ID, got[0].ID)
				assert.Equal(t, tc.expected[0].Type, got[0].Type)
				assert.Equal(t, tc.expected[0].ModelConfig, got[0].ModelConfig)
				assert.Equal(t, tc.expected[0].Status, got[0].Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}

}

func TestJobServiceGetJobByID(t *testing.T) {
	testCases := []struct {
		name        string
		expected    *Job
		expectedErr error
		setupFunc   func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo)
		jobID       string
	}{
		{
			name: "success",
			expected: &Job{
				ID:   "123",
				Type: JobTypeDeselfie,
				ModelConfig: map[string]interface{}{
					"prompt": "a photo of a person",
				},
				Status:    StatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectedErr: nil,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobByID", mock.Anything, mock.Anything).
					Return(&JobModel{ //nolint:exhaustruct
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "a photo of a person"}`)),
					}, nil)
			},
			jobID: "123",
		},
		{
			name: "failed_jobrepo",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobByID", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedErr: ErrInternal,
			expected:    nil,
			jobID:       "123",
		},
		{
			name: "failed_getjobfrommodel",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobByID", mock.Anything, mock.Anything).
					Return(&JobModel{ //nolint:exhaustruct
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "missing closing brace"`)),
					}, nil)
			},
			expectedErr: ErrInternal,
			expected:    nil,
			jobID:       "123",
		},
		{
			name: "failed_errnotfound",
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo, mockJobImageRepo *mockJobImageRepo) {
				mockRepo.On("GetJobByID", mock.Anything, mock.Anything).Return(nil, ErrNotFound)
			},
			expectedErr: ErrNotFound,
			expected:    nil,
			jobID:       "123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := newMockJobRepo()
			mockJobImageRepo := newMockJobImageRepo()
			tc.setupFunc(t, mockRepo, mockJobImageRepo)
			jobService := NewJobService(mockRepo, mockJobImageRepo, nil, nil)
			got, err := jobService.GetJobByID(t.Context(), tc.jobID)
			assert.Equal(t, tc.expectedErr, err)
			if tc.expected != nil {
				assert.Equal(t, tc.expected.ID, got.ID)
				assert.Equal(t, tc.expected.Type, got.Type)
				assert.Equal(t, tc.expected.ModelConfig, got.ModelConfig)
				assert.Equal(t, tc.expected.Status, got.Status)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJobService_DeleteJobByID(t *testing.T) {
	testCases := []struct {
		name      string
		expectErr error
		setupFunc func(t *testing.T, mockRepo *mockJobRepo)
		jobID     string
	}{
		{
			name:      "success",
			expectErr: nil,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo) {
				mockRepo.On("DeleteJobByID", mock.Anything, mock.Anything).Return(nil)
			},
			jobID: "123",
		},
		{
			name:      "failed_jobrepo",
			expectErr: ErrInternal,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo) {
				mockRepo.On("DeleteJobByID", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			jobID: "123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := newMockJobRepo()
			tc.setupFunc(t, mockRepo)
			jobService := NewJobService(mockRepo, nil, nil, nil)
			err := jobService.DeleteJobByID(t.Context(), tc.jobID)
			assert.Equal(t, tc.expectErr, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJobService_GetJobImagesByID(t *testing.T) {
	testCases := []struct {
		name      string
		expect    []*JobImage
		expectErr error
		setupFunc func(t *testing.T, mockRepo *mockJobImageRepo)
		jobID     string
	}{
		{
			name: "success",
			expect: []*JobImage{
				{
					ID:       "abc",
					JobID:    "123",
					ImageKey: "jobs/123/abc.jpg",
				},
			},
			expectErr: nil,
			setupFunc: func(t *testing.T, mockRepo *mockJobImageRepo) {
				mockRepo.On("GetImages", mock.Anything, mock.Anything).Return([]*JobImageModel{
					{
						ID:       "abc",
						JobID:    "123",
						ImageKey: "jobs/123/abc.jpg",
					},
				}, nil)
			},
			jobID: "123",
		},
		{
			name:      "failed_jobrepo",
			expect:    nil,
			expectErr: ErrInternal,
			setupFunc: func(t *testing.T, mockRepo *mockJobImageRepo) {
				mockRepo.On("GetImages", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			jobID: "123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := newMockJobImageRepo()
			tc.setupFunc(t, mockRepo)
			jobService := NewJobService(nil, mockRepo, nil, nil)
			got, err := jobService.GetJobImagesByID(t.Context(), tc.jobID)
			assert.Equal(t, tc.expectErr, err)
			for i := range tc.expect {
				assert.Equal(t, tc.expect[i].ID, got[i].ID)
				assert.Equal(t, tc.expect[i].JobID, got[i].JobID)
				assert.Equal(t, tc.expect[i].ImageKey, got[i].ImageKey)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJobService_UpdateJobStatus(t *testing.T) {
	testCases := []struct {
		name        string
		expected    *Job
		expectedErr error
		setupFunc   func(t *testing.T, mockRepo *mockJobRepo)
		jobID       string
		status      string
	}{
		{
			name: "success",
			expected: &Job{ //nolint:exhaustruct
				ID:          "123",
				Type:        JobTypeDeselfie,
				ModelConfig: map[string]interface{}{"prompt": "a photo of a person"},
				Status:      StatusFinished,
			},
			expectedErr: nil,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo) {
				mockRepo.On("UpdateJobStatus", mock.Anything, mock.Anything, mock.Anything).
					Return(&JobModel{ //nolint:exhaustruct
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusFinished,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "a photo of a person"}`)),
					}, nil)
			},
			jobID:  "123",
			status: StatusFinished,
		},
		{
			name:        "failed_jobrepo",
			expected:    nil,
			expectedErr: ErrInternal,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo) {
				mockRepo.On("UpdateJobStatus", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, assert.AnError)
			},
			jobID:  "123",
			status: StatusFinished,
		},
		{
			name:        "failed_not_found_update_job_status",
			expected:    nil,
			expectedErr: ErrNotFound,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo) {
				mockRepo.On("UpdateJobStatus", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, ErrNotFound)
			},
			jobID:  "123",
			status: StatusFinished,
		},
		{
			name:        "failed_getjobfrommodel",
			expected:    nil,
			expectedErr: ErrInternal,
			setupFunc: func(t *testing.T, mockRepo *mockJobRepo) {
				mockRepo.On("UpdateJobStatus", mock.Anything, mock.Anything, mock.Anything).
					Return(&JobModel{ //nolint:exhaustruct
						ID:          "123",
						Type:        JobTypeDeselfie,
						Status:      StatusPending,
						ModelConfig: datatypes.JSON([]byte(`{"prompt": "missing closing brace"`)),
					}, nil)
			},
			jobID:  "123",
			status: StatusFinished,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := newMockJobRepo()
			tc.setupFunc(t, mockRepo)
			jobService := NewJobService(mockRepo, nil, nil, nil)
			got, err := jobService.UpdateJobStatus(t.Context(), tc.jobID, tc.status)
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expected, got)
			mockRepo.AssertExpectations(t)
		})
	}
}
