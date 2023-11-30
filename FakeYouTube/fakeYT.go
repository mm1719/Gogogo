package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/joho/godotenv"
)

type VideoData struct {
	Title        string
	Id           string
	ChannelTitle string
	LikeCount    int64
	ViewCount    int64
	PublishedAt  string
	CommentCount int64
}

// TODO: Please create a struct to include the information of a video

func YouTubePage(w http.ResponseWriter, r *http.Request) {
	errTemplate, err := template.ParseFiles("error.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// TODO: Get API token from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file:", err)
		errTemplate.Execute(w, map[string]string{"Message": "Error loading .env file"})
		return
	}
	apiToken := os.Getenv("YOUTUBE_API_KEY")
	if apiToken == "" {
		log.Println("API token is missing, make sure it's set in the .env file")
		errTemplate.Execute(w, map[string]string{"Message": "API token is missing"})
		return
	} else {
		log.Println("Using API token:", apiToken) // Just for debugging, remove this in production
	}

	// TODO: Get video ID from URL query `v`
	videoID := r.URL.Query().Get("v")
	if videoID == "" {
		log.Println("Video ID is missing") // Log to console for debugging
		errTemplate.Execute(w, map[string]string{"Message": "Video ID is missing"})
		return
	}

	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?id=%s&key=%s&part=snippet,contentDetails,statistics", videoID, apiToken)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching video data:", err)
		errTemplate.Execute(w, map[string]string{"Message": "Error fetching video data"})
		return
	}
	defer resp.Body.Close()

	// TODO: Get video information from YouTube API
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		errTemplate.Execute(w, map[string]string{"Message": "Error reading response body"})
		return
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Println("Error parsing JSON data:", err)
		errTemplate.Execute(w, map[string]string{"Message": "Error parsing JSON data"})
		return
	}

	if e, ok := result["error"]; ok {
		log.Printf("YouTube API Error: %+v", e)
		errTemplate.Execute(w, map[string]string{"Message": "YouTube API error"})
		return
	}

	items, ok := result["items"].([]interface{})
	if !ok || len(items) == 0 {
		log.Println("No video found") // Log to console for debugging
		errTemplate.Execute(w, map[string]string{"Message": "No video found"})
		return
	}

	// TODO: Parse the JSON response and store the information into a struct
	firstItem := items[0].(map[string]interface{})
	snippet := firstItem["snippet"].(map[string]interface{})
	statistics := firstItem["statistics"].(map[string]interface{})
	likeCount, _ := strconv.ParseInt(statistics["likeCount"].(string), 10, 64)
	viewCount, _ := strconv.ParseInt(statistics["viewCount"].(string), 10, 64)
	commentCount, _ := strconv.ParseInt(statistics["commentCount"].(string), 10, 64)

	// TODO: Display the information in an HTML page through `template`
	videoData := VideoData{
		Title:        snippet["title"].(string),
		Id:           videoID,
		ChannelTitle: snippet["channelTitle"].(string),
		LikeCount:    likeCount,
		ViewCount:    viewCount,
		PublishedAt:  snippet["publishedAt"].(string),
		CommentCount: commentCount,
	}

	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		errTemplate.Execute(w, "Error loading index.html template")
		fmt.Print("Error fetching video data\n")
		return
	}

	if err := tmpl.Execute(w, videoData); err != nil {
		errTemplate.Execute(w, "Error rendering video data")
		fmt.Print("Error rendering video data\n")
	}
}

func main() {
	http.HandleFunc("/", YouTubePage)
	log.Fatal(http.ListenAndServe(":8085", nil))
}
