package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/nlopes/slack"
	"github.com/spf13/viper"

	"github.com/k33nice/slackbots/dobby/api"
	"github.com/k33nice/slackbots/dobby/fuzzy"
	"github.com/k33nice/slackbots/dobby/service"
)

var (
	version    = "0.0.1"
	config     = viper.New()
	configPath = "/etc/slackbot"
	myLog      = "/var/log/slackbot/general.log"

	bot       *slack.Client
	client    *slack.Client
	logReader io.Reader
	appLog    *os.File

	// Logger - general loger instance.
	Logger *log.Logger
)

// Color - enum for storing color.
type Color uint8

const (
	// ColorRed - uint8 red color.
	ColorRed Color = iota
	// ColorYellow - uint8 yellow color.
	ColorYellow
	// ColorGreen - uint8 green color.
	ColorGreen
	// ColorBlue - uint8 blue color.
	ColorBlue
)

var colors = map[Color]string{
	ColorRed:    "#f44336",
	ColorYellow: "#ffc107",
	ColorGreen:  "#4caf50",
	ColorBlue:   "#368ee6",
}

func init() {
	env, exists := os.LookupEnv("LOG")
	if exists {
		myLog = env
	}

	env, exists = os.LookupEnv("CONFIG")
	if exists {
		configPath = env
	}

	appLog, err := os.OpenFile(myLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	Logger = log.New(appLog, "", log.Ldate|log.Ltime|log.Lshortfile)

	config.SetConfigName("config")
	config.AddConfigPath(configPath)
	config.WatchConfig()
	// TODO: Handle config update.
	config.OnConfigChange(func(e fsnotify.Event) {
		Logger.Println("Config file changed:", e.Name)
	})

	err = config.ReadInConfig()
	if err != nil {
		Logger.Panic(err)
	}

	bot = slack.New(config.GetString("bot.token"))
	client = slack.New(config.GetString("oauth.token"))
}

func main() {
	defer appLog.Close()

	fuzzy := &fuzzy.Fuzzy{Dict: config.GetStringSlice("api.dictionary")}
	fuzzy.SetUp()
	service := &service.Service{Fixer: fuzzy, Log: Logger}
	api := &api.API{
		Log:     Logger,
		Port:    config.GetInt("api.port"),
		Token:   config.GetString("api.token"),
		Service: service,
	}
	go api.Run()

	sub := config.Sub("services")
	logs := &Logs{sub, map[string][]string{}, map[string]string{}, map[string]string{}}
	logs.collectLogs()

	files := make(chan string)
	watcher := &Watcher{logs.logsFiles, logs.logsRegexp, logs.logsChannel}
	go watcher.watchNew("/home/webbylab/itbox.ua/logs/", files)
	watcher.watchAll(files)
}

func notify(attachment slack.Attachment, message string, channel string) {
	params := slack.PostMessageParameters{}
	params.Attachments = []slack.Attachment{attachment}
	hashedChannel := strings.Join([]string{"#", channel}, "")
	channelID, timestamp, err := bot.PostMessage(hashedChannel, message, params)
	if err != nil {
		Logger.Panic(err)
	}
	Logger.Printf("Message sent to channel: %s, at: %s", channelID, timestamp)
}

// GetColor - return hex string for color.
func GetColor(color Color) string {
	return colors[color]
}
