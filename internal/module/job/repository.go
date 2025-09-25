package job

import (
	"fmt"

	"gorm.io/gorm"
)

type jobRepositoryGorm struct {
	db *gorm.DB
}

func NewJobRepositoryGorm(db *gorm.DB) JobRepository {
	return &jobRepositoryGorm{db: db}
}

func (r *jobRepositoryGorm) CreateJob() {
	fmt.Print("create job")
}

func (r *jobRepositoryGorm) GetJobs() {
	fmt.Print("get jobs")
}

func (r *jobRepositoryGorm) GetJobByID() {
	fmt.Print("get job by id")
}

func (r *jobRepositoryGorm) DeleteJobByID() {
	fmt.Print("delete job by id")
}

func (r *jobRepositoryGorm) GetJobResultsByID() {
	fmt.Print("get job results by id")
}

func (r *jobRepositoryGorm) UpdateJobStatus() {
	fmt.Print("update job status")
}

type jobResultRepositoryGorm struct {
	db *gorm.DB
}

func NewJobResultRepositoryGorm(db *gorm.DB) JobResultRepository {
	return &jobResultRepositoryGorm{db: db}
}

func (r *jobResultRepositoryGorm) CreateResult() {
	fmt.Print("create result")
}

func (r *jobResultRepositoryGorm) GetResultsByJobID() {
	fmt.Print("get results by job id")
}
