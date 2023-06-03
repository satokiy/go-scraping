package main

func main() {
	// db
	db, err := connDB()
	if err != nil {
		panic(err)
	}
	err = migrateDB(db)
	if err != nil {
		panic(err)
	}

	// scraping
	baseURL := "http://localhost:5000"
	resp, err := fetch(baseURL)
	if err != nil {
		panic(err)
	}
	indexItems, err := parseList(resp)
	if err != nil {
		panic(err)
	}

	// data update
	if err := createLatestItem(indexItems, db); err != nil {
		panic(err)
	}

	if err := updateItemMaster(db); err != nil {
		panic(err)
	}

}
