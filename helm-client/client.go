package helmclient

import (
	"context"

	"github.com/mabhi/mimic-helm-mvp/models"
	hcl "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/client-go/rest"
)

type HelmClient struct {
	client hcl.Client
	ns     string
}

func NewClient(restConfig *rest.Config, namespace string) (helmClient *HelmClient) {
	opt := &hcl.RestConfClientOptions{
		Options: &hcl.Options{
			Namespace:        namespace, // Change this to the namespace you wish the client to operate in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
			},
		},
		RestConfig: restConfig,
	}

	hc, err := hcl.NewClientFromRestConf(opt)
	if err != nil {
		panic(err)
	}
	return &HelmClient{
		client: hc,
		ns:     namespace,
	}
}

func (hc *HelmClient) InstallApp(model models.HelmAction, isUpgrage bool) (res *release.Release, err error) {
	if !isUpgrage {
		// Define a public chart repository.
		chartRepo := repo.Entry{
			Name: model.Spec.RepoName,
			URL:  model.Spec.RepoUrl,
		}

		// Add a chart-repository to the client.
		if err = hc.client.AddOrUpdateChartRepo(chartRepo); err != nil {
			return
		}
	}

	// Define the chart to be installed / upgrade
	chartSpec := hcl.ChartSpec{
		ReleaseName:     model.ActionName,
		ChartName:       model.Spec.ChartName,
		Version:         model.Spec.ChartVersion,
		Namespace:       hc.ns,
		UpgradeCRDs:     true,
		Wait:            false,
		CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	res, err = hc.client.InstallOrUpgradeChart(context.Background(), &chartSpec, nil)

	return
}

func (hc *HelmClient) DeleteApp(model models.HelmAction) (err error) {
	// Define the released chart to be installed.
	chartSpec := hcl.ChartSpec{

		ReleaseName: model.ActionName,
		ChartName:   model.Spec.ChartName,
		Version:     model.Spec.ChartVersion,
		Namespace:   hc.ns,
		Wait:        true,
	}

	// Uninstall the chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	err = hc.client.UninstallRelease(&chartSpec)
	return
}
