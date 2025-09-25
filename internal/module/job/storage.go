package job

import "fmt"

type jobObjectStorageAWS struct {
}

func NewJobObjectStorageAWS() JobObjectStorage {
	return &jobObjectStorageAWS{}
}

func (o *jobObjectStorageAWS) UploadImage() {
	fmt.Print("upload image")
}

func (o *jobObjectStorageAWS) GetPresignedURL() {
	fmt.Print("get presigned url")
}
