package main

import "forum/utils"

func main() {
	db := utils.InitialiseDB()
	if db != nil {
		defer db.Close()
	}
}
