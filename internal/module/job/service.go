package job

type jobServiceImpl struct {
	jobRepository       JobRepository
	jobResultRepository JobResultRepository
	jobEventPublisher   JobEventPublisher
	jobObjectStorage    JobObjectStorage
}

func NewJobService(jobRepository JobRepository, jobResultRepository JobResultRepository, jobEventPublisher JobEventPublisher, jobObjectStorage JobObjectStorage) JobService {
	return &jobServiceImpl{
		jobRepository:       jobRepository,
		jobResultRepository: jobResultRepository,
		jobEventPublisher:   jobEventPublisher,
		jobObjectStorage:    jobObjectStorage,
	}
}

func (s *jobServiceImpl) CreateJob() {
	s.jobRepository.CreateJob()
	s.jobEventPublisher.PublishCreatedJobEvent()
}

func (s *jobServiceImpl) GetJobs() {
	s.jobRepository.GetJobs()
}

func (s *jobServiceImpl) GetJobByID() {
	s.jobRepository.GetJobByID()
}

func (s *jobServiceImpl) DeleteJobByID() {
	s.jobRepository.DeleteJobByID()
}

func (s *jobServiceImpl) GetJobResultsByID() {
	s.jobResultRepository.GetResultsByJobID()
}

func (s *jobServiceImpl) UpdateJobStatus() {
	s.jobRepository.UpdateJobStatus()
}

func (s *jobServiceImpl) UploadImage() {
	s.jobObjectStorage.UploadImage()
}

func (s *jobServiceImpl) GetPresignedURL() {
	s.jobObjectStorage.GetPresignedURL()
}
