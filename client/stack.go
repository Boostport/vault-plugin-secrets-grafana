package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Stack struct {
	ID                       int64     `json:"id"`
	OrgID                    int64     `json:"orgId"`
	OrgSlug                  string    `json:"orgSlug"`
	OrgName                  string    `json:"orgName"`
	Name                     string    `json:"name"`
	URL                      string    `json:"url"`
	Slug                     string    `json:"slug"`
	Version                  string    `json:"version"`
	Description              string    `json:"description"`
	Status                   string    `json:"status"`
	Gateway                  string    `json:"gateway"`
	CreatedAt                time.Time `json:"createdAt"`
	CreatedBy                string    `json:"createdBy"`
	UpdatedAt                time.Time `json:"updatedAt"`
	UpdatedBy                string    `json:"updatedBy"`
	Trial                    int       `json:"trial"`
	TrialExpiresAt           time.Time `json:"trialExpiresAt"`
	ClusterID                int       `json:"clusterId"`
	ClusterSlug              string    `json:"clusterSlug"`
	ClusterName              string    `json:"clusterName"`
	Plan                     string    `json:"plan"`
	PlanName                 string    `json:"planName"`
	BillingStartDate         time.Time `json:"billingStartDate"`
	BillingEndDate           time.Time `json:"billingEndDate"`
	BillingActiveUsers       int       `json:"billingActiveUsers"`
	CurrentActiveUsers       int       `json:"currentActiveUsers"`
	CurrentActiveAdminUsers  int       `json:"currentActiveAdminUsers"`
	CurrentActiveEditorUsers int       `json:"currentActiveEditorUsers"`
	CurrentActiveViewerUsers int       `json:"currentActiveViewerUsers"`
	DailyUserCnt             int       `json:"dailyUserCnt"`
	DailyAdminCnt            int       `json:"dailyAdminCnt"`
	DailyEditorCnt           int       `json:"dailyEditorCnt"`
	DailyViewerCnt           int       `json:"dailyViewerCnt"`
	BillableUserCnt          int       `json:"billableUserCnt"`
	DashboardCnt             int       `json:"dashboardCnt"`
	DatasourceCnts           struct {
	} `json:"datasourceCnts"`
	UserQuota                         int     `json:"userQuota"`
	DashboardQuota                    int     `json:"dashboardQuota"`
	AlertQuota                        int     `json:"alertQuota"`
	Ssl                               bool    `json:"ssl"`
	CustomAuth                        bool    `json:"customAuth"`
	CustomDomain                      bool    `json:"customDomain"`
	Support                           bool    `json:"support"`
	RunningVersion                    string  `json:"runningVersion"`
	MachineLearning                   int     `json:"machineLearning"`
	HmInstancePromID                  int     `json:"hmInstancePromId"`
	HmInstancePromURL                 string  `json:"hmInstancePromUrl"`
	HmInstancePromName                string  `json:"hmInstancePromName"`
	HmInstancePromStatus              string  `json:"hmInstancePromStatus"`
	HmInstancePromCurrentUsage        float64 `json:"hmInstancePromCurrentUsage"`
	HmInstancePromCurrentActiveSeries int     `json:"hmInstancePromCurrentActiveSeries"`
	HmInstanceGraphiteID              int     `json:"hmInstanceGraphiteId"`
	HmInstanceGraphiteURL             string  `json:"hmInstanceGraphiteUrl"`
	HmInstanceGraphiteName            string  `json:"hmInstanceGraphiteName"`
	HmInstanceGraphiteType            string  `json:"hmInstanceGraphiteType"`
	HmInstanceGraphiteStatus          string  `json:"hmInstanceGraphiteStatus"`
	HmInstanceGraphiteCurrentUsage    float64 `json:"hmInstanceGraphiteCurrentUsage"`
	HlInstanceID                      int     `json:"hlInstanceId"`
	HlInstanceURL                     string  `json:"hlInstanceUrl"`
	HlInstanceName                    string  `json:"hlInstanceName"`
	HlInstanceStatus                  string  `json:"hlInstanceStatus"`
	HlInstanceCurrentUsage            float64 `json:"hlInstanceCurrentUsage"`
	AmInstanceID                      int     `json:"amInstanceId"`
	AmInstanceName                    string  `json:"amInstanceName"`
	AmInstanceURL                     string  `json:"amInstanceUrl"`
	AmInstanceStatus                  string  `json:"amInstanceStatus"`
	AmInstanceGeneratorURL            string  `json:"amInstanceGeneratorUrl"`
	HtInstanceID                      int     `json:"htInstanceId"`
	HtInstanceURL                     string  `json:"htInstanceUrl"`
	HtInstanceName                    string  `json:"htInstanceName"`
	HtInstanceStatus                  string  `json:"htInstanceStatus"`
	RegionID                          int     `json:"regionId"`
	RegionSlug                        string  `json:"regionSlug"`
	Links                             []struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"links"`
}

func (g *Grafana) StackBySlug(id string) (Stack, error) {
	stack := Stack{}
	err := g.do("GET", fmt.Sprintf("/api/instances/%s", id), nil, nil, &stack)

	if err != nil {
		return stack, fmt.Errorf("error getting stack: %w", err)
	}

	return stack, nil
}

func (g *Grafana) CreateGrafanaServiceAccountFromCloud(stack string, input CreateServiceAccountInput) (*ServiceAccount, error) {

	result := &ServiceAccount{}

	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, fmt.Sprintf("/api/instances/%s/api/serviceaccounts", stack), nil, data, result)

	if err != nil {
		return nil, fmt.Errorf("error creating service account from cloud token: %w", err)
	}

	return result, nil
}

func (g *Grafana) CreateGrafanaServiceAccountTokenFromCloud(stack string, input CreateServiceAccountTokenInput) (*ServiceAccountToken, error) {

	result := &ServiceAccountToken{}

	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("error marshalling input: %w", err)
	}

	err = g.do(http.MethodPost, fmt.Sprintf("/api/instances/%s/api/serviceaccounts/%d/tokens", stack, input.ServiceAccountID), nil, data, result)

	if err != nil {
		return nil, fmt.Errorf("error creating service account token from cloud token: %w", err)
	}

	return result, nil
}

func (g *Grafana) DeleteGrafanaServiceAccountFromCloud(stack string, serviceAccountID int64) error {

	err := g.do(http.MethodDelete, fmt.Sprintf("/api/instances/%s/api/serviceaccounts/%d", stack, serviceAccountID), nil, nil, nil)

	if err != nil {
		return fmt.Errorf("error deleting service account from cloud token: %w", err)
	}

	return nil
}
