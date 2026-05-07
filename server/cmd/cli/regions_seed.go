package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/xushop/xu-shop/internal/modules/address"
)

//go:embed assets/regions.json
var regionSeedFS embed.FS

type regionSeedRow struct {
	Code       string `json:"code"`
	ParentCode string `json:"parent_code"`
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Sort       int    `json:"sort"`
}

type regionLayerNode struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func loadSeedRegions(args []string) ([]regionSeedRow, string, error) {
	if len(args) == 0 {
		data, err := regionSeedFS.ReadFile("assets/regions.json")
		if err != nil {
			return nil, "", fmt.Errorf("读取内置区划数据失败: %w", err)
		}
		rows, err := decodeFlatRegionRows(data)
		if err != nil {
			return nil, "", err
		}
		return rows, "embedded assets/regions.json", nil
	}

	target := args[0]
	info, err := os.Stat(target)
	if err != nil {
		return nil, "", fmt.Errorf("读取路径失败: %w", err)
	}

	if info.IsDir() {
		rows, err := loadLayeredRegionDir(target)
		if err != nil {
			return nil, "", err
		}
		return rows, target, nil
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return nil, "", fmt.Errorf("读取文件失败: %w", err)
	}
	rows, err := decodeFlatRegionRows(data)
	if err != nil {
		return nil, "", err
	}
	return rows, target, nil
}

func decodeFlatRegionRows(data []byte) ([]regionSeedRow, error) {
	var rows []regionSeedRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}
	return rows, nil
}

func loadLayeredRegionDir(dir string) ([]regionSeedRow, error) {
	readNodes := func(name string, target any) error {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %w", name, err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("解析 %s 失败: %w", name, err)
		}
		return nil
	}

	var provinces []regionLayerNode
	cities := map[string][]regionLayerNode{}
	areas := map[string][]regionLayerNode{}
	streets := map[string][]regionLayerNode{}

	if err := readNodes("province.json", &provinces); err != nil {
		return nil, err
	}
	if err := readNodes("city.json", &cities); err != nil {
		return nil, err
	}
	if err := readNodes("area.json", &areas); err != nil {
		return nil, err
	}
	if err := readNodes("street.json", &streets); err != nil {
		return nil, err
	}

	rows := make([]regionSeedRow, 0, len(provinces))
	appendRows := func(nodes []regionLayerNode, parentCode string, level int) {
		for idx, node := range nodes {
			rows = append(rows, regionSeedRow{
				Code:       fmt.Sprintf("%d", node.ID),
				ParentCode: parentCode,
				Name:       node.Name,
				Level:      level,
				Sort:       idx + 1,
			})
		}
	}

	appendRows(provinces, "", 1)
	for parentCode, nodes := range cities {
		appendRows(nodes, parentCode, 2)
	}
	for parentCode, nodes := range areas {
		appendRows(nodes, parentCode, 3)
	}
	for parentCode, nodes := range streets {
		appendRows(nodes, parentCode, 4)
	}

	return dedupeRegionRows(rows), nil
}

func dedupeRegionRows(rows []regionSeedRow) []regionSeedRow {
	seen := make(map[string]struct{}, len(rows))
	deduped := make([]regionSeedRow, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.Code]; ok {
			continue
		}
		seen[row.Code] = struct{}{}
		deduped = append(deduped, row)
	}
	return deduped
}

func seedRegionsIntoDB(db *gorm.DB, rows []regionSeedRow) error {
	const batchSize = 1000

	for start := 0; start < len(rows); start += batchSize {
		end := start + batchSize
		if end > len(rows) {
			end = len(rows)
		}

		batch := make([]address.Region, 0, end-start)
		for _, row := range rows[start:end] {
			item := address.Region{
				Code:  row.Code,
				Name:  row.Name,
				Level: row.Level,
				Sort:  row.Sort,
			}
			if row.ParentCode != "" {
				item.ParentCode = &row.ParentCode
			}
			batch = append(batch, item)
		}

		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "code"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"parent_code", "name", "level", "sort",
			}),
		}).Create(&batch).Error; err != nil {
			return fmt.Errorf("upsert regions 失败: %w", err)
		}
	}

	return nil
}
