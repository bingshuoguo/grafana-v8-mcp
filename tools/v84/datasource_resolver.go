package v84

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana-openapi-client-go/models"
)

type resolvedDatasource struct {
	ResolvedBy string
	Datasource DatasourceModel
}

func resolveDatasourceRef(ctx context.Context, ref DatasourceRef) (*resolvedDatasource, error) {
	if ref.ID == nil && strings.TrimSpace(ref.UID) == "" && strings.TrimSpace(ref.Name) == "" {
		return nil, fmt.Errorf("one of id, uid, or name is required")
	}

	gc, err := getGrafanaClient(ctx)
	if err != nil {
		return nil, err
	}

	allDS, err := gc.Datasources.GetDataSources()
	if err != nil {
		return nil, fmt.Errorf("list datasources: %w", wrapOpenAPIError(err))
	}

	var (
		matched    *models.DataSourceListItemDTO
		resolvedBy string
	)

	if ref.ID != nil {
		matched = findDatasourceByID(allDS.Payload, *ref.ID)
		resolvedBy = "id"
	} else if ref.UID != "" {
		matched = findDatasourceByUID(allDS.Payload, ref.UID)
		resolvedBy = "uid"
	} else {
		matched = findDatasourceByName(allDS.Payload, ref.Name)
		resolvedBy = "name"
	}

	if matched == nil {
		return nil, fmt.Errorf("datasource not found")
	}

	full, err := gc.Datasources.GetDataSourceByID(strconv.FormatInt(matched.ID, 10))
	if err == nil && full != nil && full.Payload != nil {
		return &resolvedDatasource{ResolvedBy: resolvedBy, Datasource: dataSourceToModel(full.Payload)}, nil
	}

	// Fallback to list item fields when GetDatasourceByID is unavailable or failed.
	return &resolvedDatasource{ResolvedBy: resolvedBy, Datasource: listItemToModel(matched)}, nil
}

func findDatasourceByID(items models.DataSourceList, id int64) *models.DataSourceListItemDTO {
	for _, ds := range items {
		if ds != nil && ds.ID == id {
			return ds
		}
	}
	return nil
}

func findDatasourceByUID(items models.DataSourceList, uid string) *models.DataSourceListItemDTO {
	for _, ds := range items {
		if ds != nil && ds.UID == uid {
			return ds
		}
	}
	return nil
}

func findDatasourceByName(items models.DataSourceList, name string) *models.DataSourceListItemDTO {
	for _, ds := range items {
		if ds != nil && ds.Name == name {
			return ds
		}
	}

	lower := strings.ToLower(name)
	for _, ds := range items {
		if ds != nil && strings.ToLower(ds.Name) == lower {
			return ds
		}
	}

	return nil
}

func dataSourceToModel(ds *models.DataSource) DatasourceModel {
	if ds == nil {
		return DatasourceModel{}
	}
	var jsonData map[string]any
	if ds.JSONData != nil {
		if m, ok := ds.JSONData.(map[string]any); ok {
			jsonData = m
		}
	}
	return DatasourceModel{
		ID:               ds.ID,
		UID:              ds.UID,
		Name:             ds.Name,
		Type:             ds.Type,
		TypeLogoURL:      ds.TypeLogoURL,
		URL:              ds.URL,
		Access:           string(ds.Access),
		Database:         ds.Database,
		User:             ds.User,
		OrgID:            ds.OrgID,
		IsDefault:        ds.IsDefault,
		BasicAuth:        ds.BasicAuth,
		WithCredentials:  ds.WithCredentials,
		ReadOnly:         ds.ReadOnly,
		Version:          ds.Version,
		JSONData:         jsonData,
		SecureJSONFields: ds.SecureJSONFields,
	}
}

func listItemToModel(ds *models.DataSourceListItemDTO) DatasourceModel {
	if ds == nil {
		return DatasourceModel{}
	}
	var jsonData map[string]any
	if ds.JSONData != nil {
		if m, ok := ds.JSONData.(map[string]any); ok {
			jsonData = m
		}
	}
	return DatasourceModel{
		ID:          ds.ID,
		UID:         ds.UID,
		Name:        ds.Name,
		Type:        ds.Type,
		TypeName:    ds.TypeName,
		TypeLogoURL: ds.TypeLogoURL,
		URL:         ds.URL,
		Access:      string(ds.Access),
		Database:    ds.Database,
		User:        ds.User,
		OrgID:       ds.OrgID,
		IsDefault:   ds.IsDefault,
		BasicAuth:   ds.BasicAuth,
		ReadOnly:    ds.ReadOnly,
		JSONData:    jsonData,
	}
}
