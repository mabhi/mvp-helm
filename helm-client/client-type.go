package helmclient

import (
	"github.com/mabhi/mimic-helm-mvp/models"
	"helm.sh/helm/v3/pkg/release"
)

type IHelmClient interface {
	DeleteApp(models.HelmAction) error
	InstallApp(models.HelmAction, bool) (*release.Release, error)
}
