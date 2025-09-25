package job

import "fmt"

type jobEventPublisherInngest struct {
}

func NewJobEventPublisherInngest() JobEventPublisher {
	return &jobEventPublisherInngest{}
}

func (j *jobEventPublisherInngest) PublishCreatedJobEvent() {
	fmt.Print("publish created job event")
}
