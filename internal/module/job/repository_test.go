package job

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMockDB(t *testing.T) *gorm.DB {
	//nolint:exhaustruct
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm in-memory sqlite database: %v", err)
	}

	sqlDB, err := gormDB.DB()
	require.NoError(t, err, "expected to get sqlDB without error")
	t.Cleanup(func() {
		sqlDB.Close()
	})

	return gormDB
}

func TestNewJobRepositoryGorm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupMockDB(t)

		repo, err := NewJobRepositoryGorm(db)

		assert.NoError(t, err, "should not return error on successful migration")
		assert.NotNil(t, repo, "repository should not be nil")
	})

	t.Run("failure", func(t *testing.T) {
		db := setupMockDB(t)

		sqlDB, err := db.DB()
		require.NoError(t, err, "expected to get sqlDB without error")
		sqlDB.Close()

		repo, err := NewJobRepositoryGorm(db)

		assert.Error(t, err, "should return error when migration fails")
		assert.Nil(t, repo, "repository should be nil on migration failure")
	})
}

func TestJobRepositoryGorm_CreateJob(t *testing.T) {

	configData := map[string]interface{}{"prompt": "a photo of a person"}
	configBytes, _ := json.Marshal(configData)

	//nolint:exhaustruct
	cases := []struct {
		name        string
		id          string
		jobType     string
		modelConfig map[string]interface{}
		imageKey    string
		status      string
		setupFunc   func(t *testing.T, repo JobRepository, db *gorm.DB)
		want        *JobModel
		wantErr     error
	}{
		{
			name:        "success",
			id:          "123",
			jobType:     JobTypeDeselfie,
			modelConfig: configData,
			imageKey:    "jobs/123/image.jpg",
			status:      StatusPending,

			want: &JobModel{
				ID:          "123",
				Type:        JobTypeDeselfie,
				ModelConfig: datatypes.JSON(configBytes),
				Status:      StatusPending,
			},
			wantErr: nil,
		},
		{
			name:        "failed_ununique_primary_key",
			id:          "123",
			jobType:     JobTypeDeselfie,
			modelConfig: configData,
			imageKey:    "jobs/123/image.jpg",
			status:      StatusPending,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", "ignore", map[string]interface{}{"p": "v"}, "ignore")
				require.NoError(t, err)
			},
			want:    nil,
			wantErr: ErrConflict,
		},
		{
			name:        "failed_marshal_model_config",
			id:          "124",
			jobType:     "",
			modelConfig: map[string]interface{}{"prompt": make(chan int)},
			imageKey:    "jobs/124/image.jpg",
			status:      StatusPending,

			want:    nil,
			wantErr: ErrInternal,
		},
		{
			name:        "failed_db_connection",
			id:          "123",
			jobType:     JobTypeDeselfie,
			modelConfig: configData,
			imageKey:    "jobs/123/image.jpg",
			status:      StatusPending,

			want:    nil,
			wantErr: ErrInternal,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := setupMockDB(t)
			repo, err := NewJobRepositoryGorm(db)
			require.NoError(t, err, "NewJobRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}

			got, err := repo.CreateJob(t.Context(), c.id, c.jobType, c.modelConfig, c.status)

			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, c.want.ID, got.ID)
				assert.Equal(t, c.want.Type, got.Type)
				assert.Equal(t, c.want.Status, got.Status)
				assert.JSONEq(t, string(c.want.ModelConfig), string(got.ModelConfig))
				assert.NotEmpty(t, got.CreatedAt)
			}
		})
	}

}

func TestJobRepository_GetJobs(t *testing.T) {

	configData := map[string]interface{}{"prompt": "a photo of a person"}
	configBytes, _ := json.Marshal(configData)

	cases := []struct {
		name      string
		want      []JobModel
		wantErr   error
		setupFunc func(t *testing.T, repo JobRepository, db *gorm.DB)
	}{
		{
			name: "success",
			want: []JobModel{
				{ //nolint:exhaustruct
					ID:          "123",
					Type:        JobTypeDeselfie,
					ModelConfig: datatypes.JSON(configBytes),
					Status:      StatusPending,
				},
				{ //nolint:exhaustruct
					ID:          "124",
					Type:        JobTypeDeselfie,
					ModelConfig: datatypes.JSON(configBytes),
					Status:      StatusPending,
				},
			},
			wantErr: nil,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", JobTypeDeselfie, configData, StatusPending)
				require.NoError(t, err)
				_, err = repo.CreateJob(t.Context(), "124", JobTypeDeselfie, configData, StatusPending)
				require.NoError(t, err)
			},
		},
		{
			name:    "failed_db_connection",
			want:    nil,
			wantErr: ErrInternal,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := setupMockDB(t)
			repo, err := NewJobRepositoryGorm(db)
			require.NoError(t, err, "NewJobRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}
			got, err := repo.GetJobs(t.Context())

			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				for i, want := range c.want {
					assert.Equal(t, want.ID, got[i].ID)
					assert.Equal(t, want.Type, got[i].Type)
					assert.Equal(t, want.Status, got[i].Status)
					assert.JSONEq(t, string(want.ModelConfig), string(got[i].ModelConfig))
					assert.NotEmpty(t, got[i].CreatedAt)
				}
			}
		})
	}

}

func TestJobRepository_GetJobByID(t *testing.T) {
	configData := map[string]interface{}{"prompt": "a photo of a person"}
	configBytes, _ := json.Marshal(configData)
	cases := []struct {
		name      string
		want      *JobModel
		wantErr   error
		setupFunc func(t *testing.T, repo JobRepository, db *gorm.DB)
		id        string
	}{
		{
			name: "success",
			id:   "123",
			want: &JobModel{ //nolint:exhaustruct
				ID:          "123",
				Type:        JobTypeDeselfie,
				ModelConfig: datatypes.JSON(configBytes),
				Status:      StatusPending,
			},
			wantErr: nil,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", JobTypeDeselfie, configData, StatusPending)
				require.NoError(t, err)
			},
		},
		{
			name:    "failed_get_job_by_id",
			id:      "124",
			want:    nil,
			wantErr: ErrNotFound,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", JobTypeDeselfie, configData, StatusPending)
				require.NoError(t, err)
			},
		},
		{
			name:    "failed_db_connection",
			id:      "123",
			want:    nil,
			wantErr: ErrInternal,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := setupMockDB(t)
			repo, err := NewJobRepositoryGorm(db)
			require.NoError(t, err, "NewJobRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}
			got, err := repo.GetJobByID(t.Context(), c.id)

			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, c.want.ID, got.ID)
				assert.Equal(t, c.want.Type, got.Type)
				assert.Equal(t, c.want.Status, got.Status)
				assert.JSONEq(t, string(c.want.ModelConfig), string(got.ModelConfig))
				assert.NotEmpty(t, got.CreatedAt)
			}
		})
	}

}

func TestJobRepository_DeleteJobByID(t *testing.T) {
	configData := map[string]interface{}{"prompt": "a photo of a person"}
	cases := []struct {
		name      string
		wantErr   error
		setupFunc func(t *testing.T, repo JobRepository, db *gorm.DB)
		id        string
	}{
		{
			name:    "success",
			id:      "123",
			wantErr: nil,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", JobTypeDeselfie, configData, StatusPending)
				require.NoError(t, err)
			},
		},
		{
			name: "failed_not_found_delete_job_by_id",
			id:   "124",
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", JobTypeDeselfie, configData, StatusPending)
				require.NoError(t, err)
			},
			wantErr: nil,
		},
		{
			name:    "failed_db_connection",
			id:      "123",
			wantErr: ErrInternal,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := setupMockDB(t)
			repo, err := NewJobRepositoryGorm(db)
			require.NoError(t, err, "NewJobRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}
			err = repo.DeleteJobByID(t.Context(), c.id)
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestJobRepository_UpdateJobStatus(t *testing.T) {

	//nolint:exhaustruct
	cases := []struct {
		name      string
		want      *JobModel
		wantErr   error
		setupFunc func(t *testing.T, repo JobRepository, db *gorm.DB)
		id        string
		status    string
	}{

		{
			name:   "success",
			id:     "123",
			status: StatusFinished,
			want: &JobModel{
				ID:     "123",
				Status: StatusFinished,
			},
			wantErr: nil,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				_, err := repo.CreateJob(t.Context(), "123", JobTypeDeselfie, map[string]interface{}{"prompt": "a photo of a person"}, StatusPending)
				require.NoError(t, err)
			},
		},
		{
			name:    "failed_not_found_update_job_status",
			id:      "id_not_found",
			status:  StatusFinished,
			want:    nil,
			wantErr: ErrNotFound,
		},
		{
			name:    "failed_db_connection",
			id:      "123",
			status:  StatusFinished,
			want:    nil,
			wantErr: ErrInternal,
			setupFunc: func(t *testing.T, repo JobRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			db := setupMockDB(t)
			repo, err := NewJobRepositoryGorm(db)
			require.NoError(t, err, "NewJobRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}

			got, err := repo.UpdateJobStatus(t.Context(), c.id, c.status)
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.ID, got.ID)
				assert.Equal(t, c.want.Status, got.Status)
			}
		})
	}

}

func TestNewJobImageRepositoryGorm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupMockDB(t)

		repo, err := NewJobImageRepositoryGorm(db)

		assert.NoError(t, err, "should not return error on successful migration")
		assert.NotNil(t, repo, "repository should not be nil")
	})

	t.Run("failure", func(t *testing.T) {
		db := setupMockDB(t)

		sqlDB, err := db.DB()
		require.NoError(t, err, "expected to get sqlDB without error")
		sqlDB.Close()

		repo, err := NewJobImageRepositoryGorm(db)

		assert.Error(t, err, "should return error when migration fails")
		assert.Nil(t, repo, "repository should be nil on migration failure")
	})
}

func TestJobImageRepositoryGorm_CreateImage(t *testing.T) {

	//nolint:exhaustruct
	cases := []struct {
		name      string
		want      *JobImageModel
		wantErr   error
		id        string
		jobID     string
		imageKey  string
		setupFunc func(t *testing.T, repo JobImageRepository, db *gorm.DB)
	}{
		{
			name:     "success",
			id:       "abc",
			imageKey: "jobs/123/abc.jpg",
			jobID:    "123",
			want: &JobImageModel{
				ID:       "abc",
				JobID:    "123",
				ImageKey: "jobs/123/abc.jpg",
			},
			wantErr: nil,
		},
		{
			name:     "failed_db_connection",
			id:       "abc",
			imageKey: "jobs/123/abc.jpg",
			jobID:    "123",
			want:     nil,
			wantErr:  ErrInternal,
			setupFunc: func(t *testing.T, repo JobImageRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
		{
			name:     "failed_duplicate_key",
			id:       "abc",
			jobID:    "123",
			imageKey: "jobs/123/abc.jpg",
			wantErr:  ErrConflict,
			want:     nil,
			setupFunc: func(t *testing.T, repo JobImageRepository, db *gorm.DB) {
				_, err := repo.CreateImage(t.Context(), "abc", "123", "jobs/123/abc.jpg", ImageTypeInput)
				require.NoError(t, err)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := setupMockDB(t)
			repo, err := NewJobImageRepositoryGorm(db)
			require.NoError(t, err, "NewJobImageRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}

			got, err := repo.CreateImage(t.Context(), c.id, c.jobID, c.imageKey, ImageTypeInput)
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.ID, got.ID)
				assert.Equal(t, c.want.JobID, got.JobID)
				assert.Equal(t, c.want.ImageKey, got.ImageKey)
			}
		})
	}
}

func TestJobImageRepositoryGorm_GetResultsByJobID(t *testing.T) {
	cases := []struct {
		name      string
		jobID     string
		want      []JobImageModel
		wantErr   error
		setupFunc func(t *testing.T, repo JobImageRepository, db *gorm.DB)
	}{
		{
			name:  "success",
			jobID: "123",
			want: []JobImageModel{
				{ //nolint:exhaustruct
					ID:       "abc",
					JobID:    "123",
					ImageKey: "jobs/123/abc.jpg",
				},
				{ //nolint:exhaustruct
					ID:       "def",
					JobID:    "123",
					ImageKey: "jobs/123/def.jpg",
				},
			},
			wantErr: nil,
			setupFunc: func(t *testing.T, repo JobImageRepository, db *gorm.DB) {
				_, err := repo.CreateImage(t.Context(), "abc", "123", "jobs/123/abc.jpg", ImageTypeInput)
				require.NoError(t, err)
				_, err = repo.CreateImage(t.Context(), "def", "123", "jobs/123/def.jpg", ImageTypeInput)
				require.NoError(t, err)
			},
		},
		{ //nolint:exhaustruct
			name:    "failed_not_found",
			jobID:   "123",
			want:    []JobImageModel{},
			wantErr: nil,
		},
		{
			name:    "failed_db_connection",
			jobID:   "123",
			want:    nil,
			wantErr: ErrInternal,
			setupFunc: func(t *testing.T, repo JobImageRepository, db *gorm.DB) {
				sqlDB, _ := db.DB()
				sqlDB.Close()
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := setupMockDB(t)
			repo, err := NewJobImageRepositoryGorm(db)
			require.NoError(t, err, "NewJobImageRepositoryGorm returned error")

			if c.setupFunc != nil {
				c.setupFunc(t, repo, db)
			}

			got, err := repo.GetImages(t.Context(), c.jobID)
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(c.want), len(got))
				for i := range c.want {
					assert.Equal(t, c.want[i].ID, got[i].ID)
					assert.Equal(t, c.want[i].JobID, got[i].JobID)
					assert.Equal(t, c.want[i].ImageKey, got[i].ImageKey)
				}
			}
		})
	}
}
