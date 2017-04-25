package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
	"github.com/nlopes/slack"

	"github.com/k33nice/slackbots/dobby/channel"
)

// Watcher - handle watching log files.
type Watcher struct {
	files    map[string][]string
	regexps  map[string]string
	channels map[string]string
}

type watchEntry struct {
	channel  string
	file     string
	name     string
	regexp   string
	errLine  string
	messages chan Message
}

func (w *Watcher) watchAll(new <-chan string) {
	messages := make(chan Message)
	mx := &sync.Mutex{}

	for name, slice := range w.files {
		for _, file := range slice {
			go w.watch(name, file, w.regexps[name], mx, messages)
		}
	}

	go func(name string, mx *sync.Mutex) {
		for {
			select {
			case newFile := <-new:
				go w.watch(name, newFile, w.regexps[name], mx, messages)
			default:
			}
		}
	}("Application scripts", mx)

	for {
		msg := <-messages
		notify(msg.attachment, msg.name, msg.channel)
	}
}

func (w *Watcher) accumulator(line string, fullError *bytes.Buffer) (complete bool) {
	// FIXME: Rethink detection end or error line for accumulation.
	found := strings.Contains(line, "host")

	fullError.WriteString(line)

	if found {
		complete = true
	} else {
		complete = false
	}
	return
}

func (w *Watcher) watch(name string, file string, regexp string, mx *sync.Mutex, messages chan Message) {
	var (
		slackChan string
		fullError bytes.Buffer
		entry     watchEntry
	)

	mx.Lock()
	ch := &channel.Channel{API: client}
	slackChan = ch.GetChannel(w.channels[name], file)
	slackChannel, err := ch.New(slackChan)
	if err != nil {
		Logger.Println(err)
	}
	Logger.Printf("%+v", slackChannel)
	mx.Unlock()

	// Start watching file from the end.
	loc := &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Logger: Logger, Location: loc})
	if err != nil {
		Logger.Panic(err)
	}

TAIL:
	for line := range t.Lines {
		var errLine string
		switch name {
		case "Nginx boxes logs":
			fallthrough
		case "Nginx API log":
			complete := w.accumulator(line.Text, &fullError)

			if !complete {
				continue TAIL
			}

			errLine = fullError.String()
			fullError.Reset()
		default:
			errLine = line.Text
		}

		entry = watchEntry{slackChan, file, name, regexp, errLine, messages}
		w.watchLog(entry)

		if name == "Application scripts" {
			w.watchScript(entry)
		}
	}
}

func (w *Watcher) watchLog(entry watchEntry) {
	var filter = &Filter{entry.regexp, "", nil}
	filter.line = entry.errLine

	filter.parse()
	filter.divide()

	e := filter.err
	attachment := slack.Attachment{
		Color: e.Color,
		Text:  entry.file,
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: e.Type,
				Value: e.Message,
			},
		},
	}
	// TODO: Use filter (e.g. Bloom filter) for filtering notifies.
	if e.Type != "Deprecated" && e.Message != "" {
		entry.messages <- Message{entry.name, entry.channel, attachment}
	}
}

func (w *Watcher) watchScript(entry watchEntry) {
	var process = &Process{entry.name, entry.errLine}
	var val string

	switch {
	case process.isStart():
		val = "START"
	case process.isEnd():
		val = "END"
	default:
		return
	}

	attachment := slack.Attachment{
		Color: GetColor(ColorGreen),
		Text:  entry.file,
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: entry.name,
				Value: val,
			},
		},
	}

	entry.messages <- Message{entry.name, entry.channel, attachment}
}

func (w *Watcher) watchNew(dir string, files chan<- string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					files <- event.Name
				}
			case err := <-watcher.Errors:
				Logger.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(dir)
	if err != nil {
		Logger.Fatal(err)
	}
	<-done
}
