package models

const (
	Delete  string = "Delete"
	Install        = "Install"
	Update         = "Update"
)

type HelmAction struct {
	ActionName string
	ActionType string
	Spec       HelmSpec
}

type HelmSpec struct {
	ChartName    string
	ChartVersion string
	RepoUrl      string
	RepoName     string
}
