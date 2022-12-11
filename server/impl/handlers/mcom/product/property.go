package product

import (
	"sort"

	"go.uber.org/zap"

	"gitlab.kenda.com.tw/kenda/mcom"

	"gitlab.kenda.com.tw/kenda/mui/server/swagger/models"
)

type columnBuilder struct {
	columns     []string
	columnsUnit map[string]string
}

// newColumnBuilder creates new columnBuilder
func newColumnBuilder(columnsUnit map[string]string) *columnBuilder {
	columns := sortColumnName(columnsUnit)

	return &columnBuilder{
		columns:     columns,
		columnsUnit: columnsUnit,
	}
}

func (c *columnBuilder) build() *columnsRowData {
	columnIndex := make(map[string]int, len(c.columns))
	for i, columnName := range c.columns {
		columnIndex[columnName] = i
	}

	indexColumn := make(map[int]string, len(columnIndex))
	for name, index := range columnIndex {
		indexColumn[index] = name
	}

	return &columnsRowData{
		indexColumn: indexColumn,
		columnIndex: columnIndex,
		values:      make([]string, len(columnIndex)),
	}
}

func parseColumnBuilder(steps []*mcom.RecipeProcessStep) *columnBuilder {
	controlsUnit := make(map[string]string) // record total control's name in all step. key: control name, value: control unit
	for _, step := range steps {
		for _, control := range step.Controls {
			controlsUnit[control.Name] = control.Param.Unit
		}
	}

	return newColumnBuilder(controlsUnit)
}

// sort control name for consistency data
func sortColumnName(columnsUnit map[string]string) []string {
	columns := make([]string, len(columnsUnit))
	count := 0
	for i := range columnsUnit {
		columns[count] = i
		count++
	}
	sort.Strings(columns)

	return columns
}

type columnsRowData struct {
	indexColumn map[int]string
	columnIndex map[string]int
	values      []string
}

func (c *columnsRowData) set(key, value string) {
	i, ok := c.columnIndex[key]
	if !ok {
		zap.L().Warn("unexpected not found column",
			zap.String("column_name", key),
			zap.Any("column_map", c.columnIndex))
	}

	c.values[i] = value
}

func (c *columnsRowData) done() []*models.TableRowsItems0 {
	results := make([]*models.TableRowsItems0, len(c.values))
	for i, value := range c.values {
		results[i] = &models.TableRowsItems0{
			Name:  c.indexColumn[i],
			Value: value,
		}
	}

	return results
}

func parseColumns(c columnBuilder) []*models.TableColumnsItems0 {
	results := make([]*models.TableColumnsItems0, len(c.columns))
	for i, column := range c.columns {
		results[i] = &models.TableColumnsItems0{
			Name: column,
			Unit: c.columnsUnit[column],
		}
	}

	return results
}

func parseControlTable(configs []*mcom.RecipeProcessConfig) map[string]models.StationBOMControl {
	stationControlList := make(map[string]models.StationBOMControl)

	for _, config := range configs {
		columnBuilder := parseColumnBuilder(config.Steps)
		stationControl := models.StationBOMControl{
			Step: models.Table{
				Columns: parseColumns(*columnBuilder),
				Rows:    make([][]*models.TableRowsItems0, len(config.Steps)),
			},
		}

		for i, step := range config.Steps {
			columnsRowData := columnBuilder.build()
			for _, control := range step.Controls {
				_, value, _ := getControlValue(control.Param)
				columnsRowData.set(control.Name, value)
			}

			stationControl.Step.Rows[i] = columnsRowData.done()
		}

		stationControl.Common = make([]*models.StationBOMControlCommonItems0, len(config.CommonControls))
		for i, control := range config.CommonControls {
			max, _, min := getControlValue(control.Param)
			stationControl.Common[i] = &models.StationBOMControlCommonItems0{
				MaxValue: max,
				MinValue: min,
				RowName:  control.Name,
				Unit:     control.Param.Unit,
			}
		}

		for _, station := range config.Stations {
			stationControlList[station] = stationControl
		}
	}
	return stationControlList
}

func getControlValue(param *mcom.RecipePropertyParameter) (max string, value string, min string) {
	if v := param.High; v != nil {
		max = v.String()
	}
	if v := param.Low; v != nil {
		min = v.String()
	}
	if v := param.Mid; v != nil {
		value = v.String()
	}
	return
}
