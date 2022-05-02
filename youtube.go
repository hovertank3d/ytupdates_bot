package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

type (
	channelInfo struct {
		name        string
		id          string
		lastVideoId string
		videos      uint64
	}
	youtubeConfig struct {
		Channels []string
		Cooldown time.Duration
		Ytsecret string
	}
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func getChannelId___(service *youtube.Service, part string, forUsername string) string {
	var part_arr = []string{part}

	call := service.Channels.List(part_arr)
	call = call.ForUsername(forUsername)
	response, err := call.Do()
	handleError(err, "")

	return response.Items[0].Id
}

func getChannelId(service *youtube.Service, name string) string {

	var part = []string{"snippet"}

	call := service.Search.List(part)
	call.Type("channel")
	call.MaxResults(1)
	call.Q(name)
	response, err := call.Do()
	handleError(err, "")

	return response.Items[0].Id.ChannelId
}

func getLastVideo(service *youtube.Service, channelId string) string {
	var part_arr []string
	part_arr = append(part_arr, "snippet")

	call := service.Search.List(part_arr)
	call.ChannelId(channelId)
	call.MaxResults(1)
	call.Order("date")
	call.Type("video")

	response, err := call.Do()
	handleError(err, "")
	return response.Items[0].Id.VideoId
}

func initApi(ytconfig youtubeConfig) *youtube.Service {

	ctx := context.Background()

	b, err := ioutil.ReadFile(ytconfig.Ytsecret)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)
	service, err := youtube.New(client)

	handleError(err, "Error creating YouTube client")

	return service
}

func getNewVideos(channels []*channelInfo) ([]*channelInfo, []string) {
	var new_videos []string
	var new_channels []*channelInfo
	var part_arr = []string{"statistics,id"}

	for _, channel := range channels {
		call := yt_service.Channels.List(part_arr)
		call = call.Id(channel.id)

		response, err := call.Do()
		handleError(err, "")

		if channel.videos < response.Items[0].Statistics.VideoCount {
			lastVideo := getLastVideo(yt_service, channel.id)

			for lastVideo == channel.lastVideoId {
				continue
			}

			channel.lastVideoId = lastVideo
			new_channels = append(new_channels, channel)
			new_videos = append(new_videos, lastVideo)
		}
		channel.videos = response.Items[0].Statistics.VideoCount
	}
	return new_channels, new_videos
}

func loadChannelsInfo(service *youtube.Service, config youtubeConfig) []*channelInfo {
	var part_arr = []string{"snippet,contentDetails,statistics,id"}
	var channels []*channelInfo

	for _, id := range config.Channels {
		call := service.Channels.List(part_arr)
		call = call.Id(id)

		response, err := call.Do()
		handleError(err, "")

		var temp = channelInfo{
			videos:      response.Items[0].Statistics.VideoCount,
			name:        response.Items[0].Snippet.Title,
			id:          id,
			lastVideoId: getLastVideo(yt_service, id),
		}
		channels = append(channels, &temp)
	}
	return channels
}
