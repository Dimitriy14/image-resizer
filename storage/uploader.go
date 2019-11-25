package storage

type Uploader interface {
	Upload(ext string, content []byte) (string, error)
}
