package main

import (
	"fmt"
	"github.com/dhowden/tag"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	//	"golang.org/x/net/proxy"
	"log"
	//	"net/http"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"os"
	"os/exec"
	"strings"
)

func current() string {
	cmd, err := exec.Command("/bin/mpc", "current").Output()
	//	cmd, err := exec.Command("/bin/cat", "nowplaying").Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(cmd)
}
func Path() string {
	fileCmd, err := exec.Command("/bin/mpc", "-f", config.String("musicdir")+"/%file%", "current").Output()
	//	fileCmd, err:= exec.Command("/bin/cat","nowplayingpath").Output()
	if err == nil {
		return strings.TrimSpace(string(fileCmd))
	}
	fmt.Println(err.Error())
	return err.Error()
}
func contains(s []int, str int) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
func main() {
	/*	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:10805", &proxy.Auth{}, proxy.Direct)
		if err != nil {
			log.Fatalln("Error creating dialer: ", err)
		}

		// setup a http client
		httpTransport := &http.Transport{}
		httpClient := &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
		bot, err := tgbotapi.NewBotAPIWithClient(os.Getenv("TELEGRAM_APITOKEN"), tgbotapi.APIEndpoint, httpClient)
	*/
	config.AddDriver(yaml.Driver)
	err := config.LoadFiles("config.yml")
	if err != nil {
		panic(err)
	}
	bot, err := tgbotapi.NewBotAPI(config.String("tgtoken"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyToMessageID = update.Message.MessageID
		checkSender := func() int {
			if update.Message.From.ID != 234357406 {
				msg.Text = "You Are not my master >:("
				return 0
			} else {
				return 1
			}
		}
		switch update.Message.Command() {
		case "nowplaying":
			file, err := os.Open(Path())
			if err != nil {
				log.Fatal(err)
			}
			tagFile, err := tag.ReadFrom(file)
			if err != nil {
				log.Fatal(err)
			}
			if tagFile.Picture() == nil {
				msg.Text = current()
				msg.Text = "sent nowplaying to " + fmt.Sprint(update.Message.Chat.ID)
			} else {
				os.WriteFile("kos", tagFile.Picture().Data, 0777)
				thumb := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FilePath("./kos"))
				thumb.Caption = current()
				bot.Send(thumb)
				msg.AllowSendingWithoutReply = true
				msg.ChatID = -617652211
				msg.Text = "sent nowplaying to @" + update.Message.Chat.UserName
			}

		case "stats":
			stats, err := exec.Command("/bin/mpc", "stats").Output()
			if err != nil {
				log.Panic(err)
			}
			msg.Text = string(stats)
		case "toggle":
			if checkSender() == 1 {
				exec.Command("/bin/mpc", "toggle").Output()
				msg.Text = "Toggled playback of " + current()
			}
		case "next":
			if checkSender() == 1 {
				exec.Command("/bin/mpc", "next").Output()
				msg.Text = "Playing next song : " + current()
			}
		case "sendmusic":
			allowed := config.Ints("masters")
			if contains(allowed, int(update.Message.From.ID)) == true {
				audio := tgbotapi.NewAudio(update.Message.Chat.ID, tgbotapi.FilePath(Path()))
				bot.Send(audio)
			} else {
				if checkSender() == 1 {
					if update.Message.CommandArguments() == "" {
						audio := tgbotapi.NewAudio(update.Message.Chat.ID, tgbotapi.FilePath(Path()))
						bot.Send(audio)
						msg.DisableNotification = true
						msg.Text = "Successfuly sent the file"
						msg.ChatID = -617652211
						msg.AllowSendingWithoutReply = true
						if update.Message.Chat.UserName == "" {
							msg.Text = "sent music to" + update.Message.Chat.Title
						} else {
							msg.Text = "sent music to @" + update.Message.Chat.UserName
						}
					} else {
						usernames := config.IntMap("contacts")
						audio := tgbotapi.NewAudio(int64(usernames[update.Message.CommandArguments()]), tgbotapi.FilePath(Path()))
                                                fmt.Println(int64(usernames[update.Message.CommandArguments()]))
						msg.DisableNotification = true
						bot.Send(audio)
						msg.AllowSendingWithoutReply = true
						msg.Text = "sent music to " + " " + update.Message.CommandArguments()
					}
				}
			}
		case "playmusic":
			if checkSender() == 1 {
				exec.Command("/bin/mpc", "clear").Output()
				searchCmd, err := exec.Command("/bin/mpc", "search", "title", update.Message.CommandArguments()).Output()
				if err != nil {
					log.Panic(err)
				}
				queriedSong := strings.TrimSpace(string(searchCmd))
				exec.Command("/bin/mpc", "add", queriedSong).Output()
				exec.Command("/bin/mpc", "play").Output()
				msg.Text = "playing " + current()
			}

		default:
			msg.Text = "I don't know that command"
		}
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}

	}
}
