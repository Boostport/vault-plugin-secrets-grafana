package client

import (
	"fmt"
	"net/http"
)

type DashboardMeta struct {
	IsStarred bool   `json:"isStarred"`
	Slug      string `json:"slug"`
	Folder    int64  `json:"folderId"`
	FolderUID string `json:"folderUid"`
	URL       string `json:"url"`
}

type Dashboard struct {
	Model    map[string]interface{} `json:"dashboard"`
	FolderID int64                  `json:"folderId"`

	// This field is read-only. It is not used when creating a new dashboard.
	Meta DashboardMeta `json:"meta"`
}

func (g *Grafana) GetHomeDashboard() (Dashboard, error) {
	var result Dashboard

	err := g.do(http.MethodGet, "/api/dashboards/home", nil, nil, &result)
	if err != nil {
		return result, fmt.Errorf("error getting home dashboard: %w", err)
	}

	return result, nil
}
