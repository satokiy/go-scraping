package main

import (
	"fmt"

	"gorm.io/gorm"
)

func createLatestItem(items []Item, db *gorm.DB) error {

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&LatestItem{}); err != nil {
		return fmt.Errorf("get latest_items parse error: %w", err)
	}

	if err := db.Exec("TRUNCATE " + stmt.Schema.Table + ";").Error; err != nil {
		return fmt.Errorf("truncate latest_items error: %w", err)
	}

	var latestItems []LatestItem
	for _, item := range items {
		latestItems = append(latestItems, LatestItem{Item: item})
	}

	if err := db.CreateInBatches(latestItems, 100).Error; err != nil {
		return fmt.Errorf("bulk insert to latest items error: %w", err)
	}

	return nil
}

func updateItemMaster(db *gorm.DB) error {
	// 整合性を保つために、データの追加、更新、削除はトランザクション内で実行
	return db.Transaction(func(tx *gorm.DB) error {
		// Insert
		var newItems []LatestItem
		err := tx.Unscoped().Joins("LEFT JOIN item_master ON latest_items.url = item_master.url").Where("item_master.name IS NULL").Find(&newItems).Error
		if err != nil {
			return fmt.Errorf("extract for bulk insert to item_master error: %w", err)
		}

		var insertRecords []ItemMaster
		// initDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.Local)
		for _, newItem := range newItems {
			insertRecords = append(insertRecords, ItemMaster{Item: newItem.Item})
			fmt.Printf("Index item is created: %s\n", newItem.URL)
		}
		if err := tx.CreateInBatches(insertRecords, 100).Error; err != nil {
			return fmt.Errorf("bulk insert to item_master error: %w", err)
		}

		// Update
		var updatedItems []LatestItem
		err = tx.Unscoped().Joins("INNER JOIN item_master ON latest_items.url = item_master.url").Where("latest_items.name <> item_master.name OR latest_items.price <> item_master.price OR item_master.deleted_at IS NOT NULL").Find(&updatedItems).Error
		if err != nil {
			return fmt.Errorf("update error: %w", err)
		}
		for _, updatedItem := range updatedItems {
			err := tx.Unscoped().Model(ItemMaster{}).Where("url = ?", updatedItem.URL).Updates(map[string]interface{}{"name": updatedItem.Name, "price": updatedItem.Price, "deleted_at": nil}).Error
			if err != nil {
				return fmt.Errorf("update error: %w", err)
			}
			fmt.Printf("Index item is updated: %s\n", updatedItem.URL)
		}

		// Delete
		var deletedItems []ItemMaster
		if err := tx.Where("NOT EXISTS(SELECT 1 FROM latest_items li WHERE li.url = item_master.url)").Find(&deletedItems).Error; err != nil {
			return fmt.Errorf("delete error: %w", err)
		}
		var ids []uint
		for _, deletedItem := range deletedItems {
			ids = append(ids, deletedItem.ID)
			// 動作確認のために、ログを出力
			fmt.Printf("Index item is deleted: %s\n", deletedItem.URL)
		}
		if len(ids) > 0 {
			if err := tx.Delete(&deletedItems).Error; err != nil {
				return fmt.Errorf("delete error: %w", err)
			}
		}

		return nil
	})
}
