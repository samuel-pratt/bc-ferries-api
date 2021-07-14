package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/robfig/cron"
)

func updateSchedule() {
	response := scraper()
	jsonString, _ := json.Marshal(response)
	err := ioutil.WriteFile("sailings.json", []byte(jsonString), 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print("Updated sailings.json at: ")
	fmt.Println(time.Now())
}

func getDataFromFile(w http.ResponseWriter, r *http.Request) {
	data, err2 := ioutil.ReadFile("sailings.json")
	if err2 != nil {
		fmt.Println(err2)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func main() {
	c := cron.New()
	c.AddFunc("@every 1m", updateSchedule)
	c.Start()

	http.HandleFunc("/api/", getDataFromFile)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(os.Getenv("PORT")	, nil)
}
