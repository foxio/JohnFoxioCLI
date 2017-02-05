package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"
	notifier "github.com/deckarep/gosx-notifier"
	"github.com/tbruyelle/hipchat-go/hipchat"

	"github.com/foxio/john_foxio_cli/services"
)

var doneChan chan bool
var duration int
var breakDuration int

// PomodoroConfiguration represents the pom config file
type PomodoroConfiguration struct {
	RunTime int
	Break   int
}

// PomodoroStart starts a pom
func PomodoroStart(c *cli.Context, config *Configuration) {
	doneChan = make(chan bool)

	duration = c.Int("duration")
	if duration <= 0 {
		duration = config.Pomodoro.RunTime
	}
	breakDuration = config.Pomodoro.Break

	fmt.Printf("Pom started for %d mintues\n", duration)
	displayNotification(fmt.Sprintf("Pom started for %d mintues", duration))

	go runTimer(duration, pomCompleted, pomTick)

	userPresence := hipchat.UpdateUserPresenceRequest{
		Show:   hipchat.UserPresenceShowDnd,
		Status: fmt.Sprintf("In %dm Pom", duration),
	}
	updateHipChatStatus(userPresence)

	<-doneChan
	fmt.Println("done")
	pomStartBreak()
}

// PomodoroStop stops a pom
func PomodoroStop(c *cli.Context) {
	fmt.Println("Ending pom ...")
	displayNotification("Pom stopped")
}

func pomCompleted() {
	displayNotification("Pom complete.")
	doneChan <- true
}

func pomTick(maxMinutes int, minute int) {
	if minute%2 == 0 {
		userPresence := hipchat.UpdateUserPresenceRequest{
			Show:   hipchat.UserPresenceShowDnd,
			Status: fmt.Sprintf("%dm left in Pom", maxMinutes-minute),
		}
		updateHipChatStatus(userPresence)
	}
}

func breakTick(maxMinutes int, minute int) {

}

func pomStartBreak() {
	fmt.Println("Break starting")

	userPresence := hipchat.UpdateUserPresenceRequest{
		Show:   hipchat.UserPresenceShowChat,
		Status: "",
	}
	updateHipChatStatus(userPresence)

	displayNotification("Break Time!")

	go runTimer(breakDuration, pomBreakOver, breakTick)
	<-doneChan
}

func pomBreakOver() {
	displayNotification("Break Over.")
	doneChan <- true
}

func runTimer(maxMinutes int, callback func(), tickCallback func(maxMinutes int, minute int)) {
	startTime := time.Now()

	fmt.Printf("\r0 minute")
	tick := time.NewTicker(1 * time.Minute)
	for now := range tick.C {

		minute := int(now.Sub(startTime).Minutes())
		fmt.Printf("\r%d minute", minute)

		tickCallback(maxMinutes, minute)

		if minute >= maxMinutes {
			fmt.Printf("\n")
			tick.Stop()
			callback()
		}
	}
}

func displayNotification(message string) {
	note := notifier.Notification{
		Title:   "John Foxio",
		Message: message,
		AppIcon: "./command/notification_icon.png",
		Sound:   notifier.Hero,
	}

	note.Push()
}

func updateHipChatStatus(userPresence hipchat.UpdateUserPresenceRequest) {
	var service services.Servicer
	hipchatService := services.HipchatService{}
	service = hipchatService
	if service.Available() {
		hipchatService.SetStatus(userPresence)
	}
}
