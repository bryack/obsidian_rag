package domain

// AllScope represents no filtering - returns all matching documents.
type AllScope struct{}

func (AllScope) Name() string {
	return "all"
}

// FolderScope filters documents by file path prefix.
type FolderScope struct {
	Path string
}

func (f FolderScope) Name() string {
	return f.Path
}
