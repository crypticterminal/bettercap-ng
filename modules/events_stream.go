package modules

import (
	"fmt"
	"strings"

	"github.com/evilsocket/bettercap-ng/core"
	"github.com/evilsocket/bettercap-ng/session"
)

type EventsStream struct {
	session.SessionModule
	filter string
	quit   chan bool
}

func NewEventsStream(s *session.Session) *EventsStream {
	stream := &EventsStream{
		SessionModule: session.NewSessionModule("events.stream", s),
		filter:        "",
		quit:          make(chan bool),
	}

	stream.AddParam(session.NewStringParameter("events.stream.filter",
		"",
		"",
		"If filled, filter events by this prefix type."))

	stream.AddHandler(session.NewModuleHandler("events.stream on", "",
		"Start events stream.",
		func(args []string) error {
			return stream.Start()
		}))

	stream.AddHandler(session.NewModuleHandler("events.stream off", "",
		"Stop events stream.",
		func(args []string) error {
			return stream.Stop()
		}))

	stream.AddHandler(session.NewModuleHandler("events.clear", "",
		"Clear events stream.",
		func(args []string) error {
			stream.Session.Events.Clear()
			return nil
		}))

	return stream
}

func (s EventsStream) Name() string {
	return "events.stream"
}

func (s EventsStream) Description() string {
	return "Print events as a continuous stream."
}

func (s EventsStream) Author() string {
	return "Simone Margaritelli <evilsocket@protonmail.com>"
}

func (s *EventsStream) Configure() error {
	var err error
	if err, s.filter = s.StringParam("events.stream.filter"); err != nil {
		return err
	}
	return nil
}

func (s *EventsStream) Start() error {
	if s.Running() == true {
		return session.ErrAlreadyStarted
	} else if err := s.Configure(); err != nil {
		return err
	}

	s.SetRunning(true)

	go func() {
		for {
			var e session.Event
			select {
			case e = <-s.Session.Events.NewEvents:
				if s.filter == "" || strings.Contains(e.Tag, s.filter) {
					tm := e.Time.Format("2006-01-02 15:04:05")

					if e.Tag == "sys.log" {
						fmt.Printf("[%s] [%s] (%s) %s\n", tm, core.Green(e.Tag), e.Label(), e.Data.(session.LogMessage).Message)
					} else {
						fmt.Printf("[%s] [%s] %v\n", tm, core.Green(e.Tag), e.Data)
					}

					s.Session.Refresh()
				}
				break

			case <-s.quit:
				return
			}
		}
	}()

	return nil
}

func (s *EventsStream) Stop() error {
	if s.Running() == false {
		return session.ErrAlreadyStopped
	}
	s.SetRunning(false)
	s.quit <- true
	return nil
}
