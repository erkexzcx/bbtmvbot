package main

import (
	"fmt"
	"net/http"
)

func defineInfluxHTTP() {

	http.HandleFunc("/influx", func(w http.ResponseWriter, r *http.Request) {
		printInfluxData(&w)
	})

}

func printInfluxData(w *http.ResponseWriter) {
	sql := `
	SELECT 'portal' AS "type", 'alio.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%alio.lt%"
	UNION SELECT 'portal' AS "type", 'aruodas.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%aruodas.lt%"
	UNION SELECT 'portal' AS "type", 'domoplius.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%domoplius.lt%"
	UNION SELECT 'portal' AS "type", 'kampas.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%kampas.lt%"
	UNION SELECT 'portal' AS "type", 'nuomininkai.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%nuomininkai.lt%"
	UNION SELECT 'portal' AS "type", 'rinka.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%rinka.lt%"
	UNION SELECT 'portal' AS "type", 'skelbiu.lt' AS "key", COUNT(*) AS "value" FROM posts WHERE url LIKE "%skelbiu.lt%"
	UNION SELECT 'users' AS "type", 'visited' AS "key", COUNT(*) AS "value" FROM users
	UNION SELECT 'users' AS "type", 'enabled' AS "key", COUNT(*) AS "value" FROM users WHERE enabled = 1
	UNION SELECT 'user_preferences' AS "type", 'avg_price_from' AS "key", (SELECT CAST(AVG(price_from) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'user_preferences' AS "type", 'avg_price_to' AS "key", (SELECT CAST(AVG(price_to) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'user_preferences' AS "type", 'avg_rooms_from' AS "key", (SELECT CAST(AVG(rooms_from) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'user_preferences' AS "type", 'avg_rooms_to' AS "key", (SELECT CAST(AVG(rooms_to) AS INT) FROM users WHERE enabled=1) AS "value"
	UNION SELECT 'posts' AS "type", 'total' AS "key", (SELECT COUNT(*) FROM posts) AS "value"
	UNION SELECT 'posts' AS "type", 'excluded' AS "key", (SELECT COUNT(*) FROM posts WHERE excluded=1) AS "value"
	UNION SELECT 'posts' AS "type", 'sent' AS "key", (SELECT COUNT(*) FROM posts WHERE excluded=0) AS "value"
	UNION SELECT 'posts' AS "type", 'no_price' AS "key", (SELECT COUNT(*) FROM posts WHERE reason="0eur price") AS "value"
	`

	rows, err := db.Query(sql)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var typ, key string
		var value int
		err = rows.Scan(&typ, &key, &value)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Fprintf(*w, "bbtmvbot,type=%s,key=%s value=%d\n", typ, key, value)
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
		return
	}
}
