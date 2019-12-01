package storage

//go:generate mockgen -destination=../mocks/mock-storage.go -mock_names=Storage=MockStorage -package=mocks github.com/Dimitriy14/image-resizing/storage Storage
type Storage interface {
	Upload(fileExt string, content []byte) (link string, err error)
	UploadWithOriginal(filExt string, originalImgContent, resizedImgContent []byte) (string, string, error)
	Download(addr string) (fileContent []byte, err error)
	DeleteImage(addr string) error
}
