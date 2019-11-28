package storage

type Storage interface {
	Uploader
	Downloader
}

type Uploader interface {
	Upload(fileExt string, content []byte) (link string, err error)
	UploadWithOriginal(filExt string, originalImgContent, resizedImgContent []byte) (string, string, error)
}

type Downloader interface {
	Download(url string) (fileContent []byte, err error)
}
