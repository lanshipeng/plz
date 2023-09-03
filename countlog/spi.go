package countlog

import (
	"os"

	"github.com/v2pro/plz/countlog/output"
	"github.com/v2pro/plz/countlog/output/hrf"
	"github.com/v2pro/plz/countlog/spi"
	"github.com/v2pro/plz/countlog/stats"
)

var EventWriter spi.EventSink = output.NewEventWriter(output.EventWriterConfig{
	Format: &hrf.Format{},
	Writer: os.Stdout,
})

var EventAggregator spi.EventSink = stats.NewEventAggregator(stats.EventAggregatorConfig{
	Collector: nil, // set Collector to enable stats
})

type LogWriter interface {
	ShouldLog(level int, event string, properties []interface{}) bool
	WriteLog(level int, event string, properties []interface{})
}

type LogFormatter interface {
	FormatLog(event Event) []byte
}

type LogOutput interface {
	OutputLog(level int, timestamp int64, formattedEvent []byte)
	Close()
}

type Event struct {
	Level      int
	Event      string
	Properties []interface{}
}

func (event Event) Get(target string) interface{} {
	for i := 0; i < len(event.Properties); i += 2 {
		k, _ := event.Properties[i].(string)
		if k == target {
			return event.Properties[i+1]
		}
	}
	return nil
}

func (event Event) LevelName() string {
	return getLevelName(event.Level)
}

var LogWriters = []LogWriter{}
