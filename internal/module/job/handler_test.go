package job

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"selfier/pkg/middleware"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --------------------
// Constants
// --------------------
const (
	testJobType   = "deselfie"
	testImageName = "photo.png"
	testImageData = "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"
	testModelJSON = `{"prompt":"a photo of a person"}`
	testPath      = "/jobs"
)

// --------------------
// Mock JobService
// --------------------
type mockJobService struct {
	mock.Mock
}

//nolint:exhaustruct
func newMockJobService() *mockJobService { return &mockJobService{} }

func (s *mockJobService) CreateJob(ctx context.Context, jobType string, modelConfig map[string]interface{}, imageReader io.Reader, imageFilename string) (*Job, error) {
	args := s.Called(ctx, jobType, modelConfig, imageReader, imageFilename)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Job), args.Error(1)
}

func (s *mockJobService) GetJobs(ctx context.Context) ([]*Job, error) {
	args := s.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Job), args.Error(1)
}

func (s *mockJobService) GetJobByID(ctx context.Context, jobID string) (*Job, error) {
	args := s.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Job), args.Error(1)
}

func (s *mockJobService) DeleteJobByID(ctx context.Context, jobID string) error {
	args := s.Called(ctx, jobID)
	return args.Error(0)
}

func (s *mockJobService) GetJobResultsByID(ctx context.Context, jobID string) ([]*JobResult, error) {
	args := s.Called(ctx, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*JobResult), args.Error(1)
}

func (s *mockJobService) UpdateJobStatus(ctx context.Context, jobID string, status string) (*Job, error) {
	args := s.Called(ctx, jobID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Job), args.Error(1)
}

func (s *mockJobService) GetPresignedURL(ctx context.Context, imageKey string) (string, error) {
	args := s.Called(ctx, imageKey)
	return args.String(0), args.Error(1)
}

// --------------------
// Test Helpers
// -------------------

// Creates a multipart form request with fields and an optional file
func createMultipartRequest(t *testing.T, path string, fields map[string]string, fileField, fileName, fileData string) *http.Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}
	if fileData != "" {
		filePart, err := writer.CreateFormFile(fileField, fileName)
		require.NoError(t, err)
		_, err = filePart.Write([]byte(fileData))
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())
	req, err := http.NewRequest("POST", path, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// Sets up router with middleware and registers a single endpoint
func setupRouter() (*http.ServeMux, huma.API) {
	router := http.NewServeMux()
	api := humago.New(router, huma.DefaultConfig("Test API", "1.0.0"))

	api.UseMiddleware(middleware.AuthMiddleware())
	api.UseMiddleware(middleware.RequestIDMiddleware())
	api.UseMiddleware(middleware.LoggerMiddleware(slog.Default()))

	return router, api
}

// --------------------
// CreateJob Tests
// --------------------
func TestJobHTTPHandler_CreateJob(t *testing.T) {
	mockJob := Job{
		ID:          "123",
		Type:        testJobType,
		ModelConfig: map[string]interface{}{"prompt": "a photo of a person"},
		ImageKey:    "jobs/123/image.jpg",
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	cases := []struct {
		name       string
		fields     map[string]string
		fileName   string
		fileData   string
		mockReturn *Job
		mockError  error
		wantStatus int
	}{
		{
			name: "success",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": testModelJSON,
			},
			fileName:   testImageName,
			fileData:   testImageData,
			mockReturn: &mockJob,
			mockError:  nil,
			wantStatus: http.StatusCreated,
		},
		{
			name: "service error",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": testModelJSON,
			},
			fileName:   testImageName,
			fileData:   testImageData,
			mockReturn: nil,
			mockError:  assert.AnError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "missing image",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": testModelJSON,
			},
			fileName:   testImageName,
			fileData:   "",
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid model_config",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": "invalid json",
			},
			fileName:   testImageName,
			fileData:   testImageData,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid file extension",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": testModelJSON,
			},
			fileName:   "invalid.pdf",
			fileData:   testImageData,
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid image size",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": testModelJSON,
			},
			fileName:   testImageName,
			fileData:   strings.Repeat("x", 10*1024*1024+1),
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "cannot read image",
			fields: map[string]string{
				"type":         testJobType,
				"model_config": testModelJSON,
			},
			fileName:   testImageName,
			fileData:   "not image data",
			mockReturn: nil,
			mockError:  nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newMockJobService()
			jobHandler := NewJobHTTPHandler(service)

			if tc.mockReturn != nil || tc.mockError != nil {
				service.On("CreateJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.mockReturn, tc.mockError).Once()
			}

			router, api := setupRouter()

			huma.Register(api, huma.Operation{ //nolint:exhaustruct
				OperationID:   "create-job",
				Method:        http.MethodPost,
				Path:          "/jobs",
				Summary:       "Create a new job",
				Tags:          []string{"Jobs"},
				DefaultStatus: http.StatusCreated,
			}, jobHandler.CreateJob)

			req := createMultipartRequest(t, testPath, tc.fields, "image", tc.fileName, tc.fileData)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusCreated {
				var actual Job
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actual))
				assert.Equal(t, mockJob.ID, actual.ID)
				assert.Equal(t, mockJob.Type, actual.Type)
				assert.Equal(t, mockJob.ModelConfig, actual.ModelConfig)
				assert.Equal(t, mockJob.ImageKey, actual.ImageKey)
				assert.Equal(t, mockJob.Status, actual.Status)
				assert.False(t, actual.CreatedAt.IsZero())
				assert.False(t, actual.UpdatedAt.IsZero())
			}

			service.AssertExpectations(t)
		})
	}
}

// --------------------
// GetJobs Tests
// --------------------
func TestJobHTTPHandler_GetJobs(t *testing.T) {
	//nolint:exhaustruct
	job1 := &Job{ID: "1", Type: testJobType, Status: StatusPending}
	//nolint:exhaustruct
	job2 := &Job{ID: "2", Type: testJobType, Status: StatusFinished}

	cases := []struct {
		name       string
		mockReturn []*Job
		mockError  error
		wantStatus int
		wantCount  int
	}{
		{
			name:       "success",
			mockReturn: []*Job{job1, job2},
			mockError:  nil,
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:       "service error",
			mockReturn: nil,
			mockError:  assert.AnError,
			wantStatus: http.StatusInternalServerError,
			wantCount:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newMockJobService()
			jobHandler := NewJobHTTPHandler(service)

			service.On("GetJobs", mock.Anything).Return(tc.mockReturn, tc.mockError).Once()

			router, api := setupRouter()

			huma.Register(api, huma.Operation{ //nolint:exhaustruct
				OperationID:   "get-jobs",
				Method:        http.MethodGet,
				Path:          "/jobs",
				Summary:       "Get all jobs",
				Tags:          []string{"Jobs"},
				DefaultStatus: http.StatusOK,
			}, jobHandler.GetJobs)

			req, _ := http.NewRequest("GET", testPath, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusOK {
				var jobs []*Job
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &jobs))
				assert.Len(t, jobs, tc.wantCount)
			}

			service.AssertExpectations(t)
		})
	}
}

func TestJobHTTPHandler_GetJobByID(t *testing.T) {
	//nolint:exhaustruct
	job1 := &Job{ID: "1", Type: testJobType, Status: StatusPending}

	cases := []struct {
		name       string
		wantStatus int
		pathParam  string
		setupFunc  func(t *testing.T, mockService *mockJobService)
	}{
		{
			name:       "success",
			wantStatus: http.StatusOK,
			pathParam:  "/1",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("GetJobByID", mock.Anything, mock.Anything).Return(job1, nil).Once()
			},
		},
		{
			name:       "service error",
			wantStatus: http.StatusInternalServerError,
			pathParam:  "/1",
			setupFunc: func(t *testing.T, mockService *mockJobService) {

				mockService.On("GetJobByID", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
			},
		},
		{
			name:       "failed_not_found_get_job_by_id",
			wantStatus: http.StatusNotFound,
			pathParam:  "/1",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("GetJobByID", mock.Anything, mock.Anything).Return(nil, ErrNotFound).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newMockJobService()
			jobHandler := NewJobHTTPHandler(service)

			if tc.setupFunc != nil {
				tc.setupFunc(t, service)
			}

			router, api := setupRouter()

			huma.Register(api, huma.Operation{ //nolint:exhaustruct
				OperationID:   "get-job-by-id",
				Method:        http.MethodGet,
				Path:          "/jobs/{id}",
				Summary:       "Get job by id",
				Tags:          []string{"Jobs"},
				DefaultStatus: http.StatusOK,
			}, jobHandler.GetJobByID)

			req, _ := http.NewRequest("GET", testPath+tc.pathParam, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusCreated {
				var actual Job
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actual))
				assert.Equal(t, *job1, actual)
			}
			service.AssertExpectations(t)
		})
	}
}

func TestJobHTTPHandler_DeleteJobByID(t *testing.T) {

	cases := []struct {
		name       string
		wantStatus int
		path       string
		setupFunc  func(t *testing.T, mockService *mockJobService)
	}{
		{
			name:       "success",
			wantStatus: http.StatusOK,
			path:       "/jobs/1",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("DeleteJobByID", mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
		{
			name:       "service error",
			wantStatus: http.StatusInternalServerError,
			path:       "/jobs/1",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("DeleteJobByID", mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newMockJobService()
			jobHandler := NewJobHTTPHandler(service)

			if tc.setupFunc != nil {
				tc.setupFunc(t, service)
			}

			router, api := setupRouter()

			huma.Register(api, huma.Operation{ //nolint:exhaustruct
				OperationID:   "delete-job-by-id",
				Method:        http.MethodDelete,
				Path:          "/jobs/{id}",
				Summary:       "Delete job by id",
				Tags:          []string{"Jobs"},
				DefaultStatus: http.StatusOK,
			}, jobHandler.DeleteJobByID)

			req, _ := http.NewRequest("DELETE", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)

			service.AssertExpectations(t)
		})
	}
}

func TestJobHTTPHandler_GetJobResultsByID(t *testing.T) {

	//nolint:exhaustruct
	jobResult1 := &JobResult{
		ID:                "1",
		JobID:             "1",
		ImagePresignedURL: "https://example.com/image_1.jpg",
	}
	//nolint:exhaustruct
	jobResult2 := &JobResult{
		ID:                "2",
		JobID:             "1",
		ImagePresignedURL: "https://example.com/image_2.jpg",
	}

	cases := []struct {
		name       string
		path       string
		setupFunc  func(t *testing.T, mockService *mockJobService)
		wantStatus int
		want       []*JobResult
	}{
		{
			name:       "success",
			wantStatus: http.StatusOK,
			path:       "/jobs/1/results",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("GetJobResultsByID", mock.Anything, mock.Anything).Return([]*JobResult{jobResult1, jobResult2}, nil).Once()
			},
			want: []*JobResult{jobResult1, jobResult2},
		},
		{
			name:       "service error",
			wantStatus: http.StatusInternalServerError,
			path:       "/jobs/1/results",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("GetJobResultsByID", mock.Anything, mock.Anything).Return(nil, assert.AnError).Once()
			},
			want: nil,
		},
		{
			name:       "not found",
			wantStatus: http.StatusNotFound,
			path:       "/jobs/1/results",
			setupFunc: func(t *testing.T, mockService *mockJobService) {
				mockService.On("GetJobResultsByID", mock.Anything, mock.Anything).Return(nil, ErrNotFound).Once()
			},
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service := newMockJobService()
			jobHandler := NewJobHTTPHandler(service)

			if tc.setupFunc != nil {
				tc.setupFunc(t, service)
			}

			router, api := setupRouter()

			huma.Register(api, huma.Operation{ //nolint:exhaustruct
				OperationID:   "get-job-results-by-id",
				Method:        http.MethodGet,
				Path:          "/jobs/{id}/results",
				Summary:       "Get job results by id",
				Tags:          []string{"Jobs"},
				DefaultStatus: http.StatusOK,
			}, jobHandler.GetJobResultsByID)

			req, _ := http.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.wantStatus, w.Code)

			if tc.wantStatus == http.StatusOK {
				var actual []*JobResult
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &actual))
				assert.Equal(t, tc.want, actual)
			}
			service.AssertExpectations(t)
		})
	}
}
