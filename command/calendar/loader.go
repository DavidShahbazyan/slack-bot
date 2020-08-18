package calendar

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/PuloV/ics-golang"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
)

const (
	storeKey      = "calendar"
	checkInPast   = time.Hour * 2
	checkInterval = time.Minute * 5
)

type event struct {
	CalendarEvent ics.Event
	Calendar      config.Calendar
	Event         config.CalendarEvent
	Params        map[string]string
}

func waitForEvents(calendars []config.Calendar) chan event {
	eventChan := make(chan event)

	lastCheck := time.Now().Add(-checkInPast)
	go func() {
		for {
			for _, calConfig := range loadEvents(calendars) {
				calendarEvent := calConfig.Ical
				// todo fix timezone handling
				startDate := calendarEvent.GetStart().Add(-time.Hour * 2)

				// ignore passed events
				if startDate.Before(lastCheck) {
					continue
				}

				// ignore not started events
				if startDate.After(time.Now()) {
					continue
				}

				var existingEntry string
				storage.Read(storeKey, calendarEvent.GetID(), existingEntry)
				if existingEntry != "" {
					// event was already handled
					continue
				}
				storage.Write(storeKey, calendarEvent.GetID(), time.Now().String())

				for _, eventDefinition := range calConfig.Events {
					re := util.CompileRegexp(eventDefinition.Pattern)
					match := re.FindStringSubmatch(calendarEvent.GetSummary())
					if len(match) == 0 {
						continue
					}

					eventChan <- event{
						CalendarEvent: calendarEvent,
						Calendar:      calConfig,
						Event:         eventDefinition,
						Params:        util.RegexpResultToParams(re, match),
					}
				}
			}

			lastCheck = time.Now()

			time.Sleep(checkInterval)
		}
	}()

	return eventChan
}

func loadEvents(calendars []config.Calendar) []config.Calendar {
	events := make([]config.Calendar, 0)
	for _, calConfig := range calendars {
		calendar := loadCalender(calConfig)
		for _, event := range calendar.GetEvents() {
			calConfig.Ical = event
			events = append(events, calConfig)
		}
	}

	return events
}

// todo load all calenders with one parser
func loadCalender(calendar config.Calendar) *ics.Calendar {
	ics.RepeatRuleApply = true
	parser := ics.New()
	ics.FilePath = os.TempDir() + "/slack-bot"

	if strings.HasPrefix(calendar.Path, "http") {
		parserChan := parser.GetInputChan()
		parserChan <- calendar.Path
	} else {
		content, err := ioutil.ReadFile(calendar.Path)
		if err != nil {
			panic(err) // todo(easypick) use logger
		}
		parser.Load(string(content))
	}

	parser.Wait()
	calendars, _ := parser.GetCalendars()
	cal := calendars[0]

	return cal
}
