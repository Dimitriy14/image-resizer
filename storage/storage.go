package storage

type Storage interface {
	Upload(fileExt string, content []byte) (link string, err error)
	UploadWithOriginal(filExt string, originalImgContent, resizedImgContent []byte) (string, string, error)
	Download(addr string) (fileContent []byte, err error)
	DeleteImage(addr string) error
}
