package hookers

type RegisterHookRequest struct {
	Provider    string `json:"provider" form:"provider" validate:"required"`
	ProjectPath string `json:"repo_path" form:"repo_path" validate:"required"`
}

type RegisterHookerResponse struct {
	Secret string `json:"secret,omitempty"`
}

type SearchHookerResponse struct {
	Secrets  string `json:"-"`
	RepoPath string `json:"repo_path,omitempty"`
	Provider string `json:"provider"`
}

type FindHookByProvider struct {
	Provider string `json:"provider" query:"provider" db:"provider" validate:"required"`
	RepoPath string `json:"repo_path" query:"repo_path" db:"repo_path"`
	Dive     bool   `json:"-" query:"-" db:"-" form:"-"`
}
