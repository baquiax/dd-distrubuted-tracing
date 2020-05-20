package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const fileName = "app.log"

func writer() {
	error := ioutil.WriteFile(fileName, []byte{}, 0644)

	if error != nil {
		log.Fatalf("Error creating the log file")
	}

	file, error := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if error != nil {
		log.Fatalf("Error opening the log file")
	}

	defer file.Close()

	for {
		tracer.Start(tracer.WithServiceName("log-writer"))
		span := tracer.StartSpan("write", tracer.ResourceName(fileName))

		traceID := span.Context().TraceID()

		span.SetTag("writer", true)
		span.SetTag(ext.AnalyticsEvent, true)

		time.Sleep(time.Second * 2)

		_, error := file.WriteString(fmt.Sprintf("%d\n", traceID))

		if error != nil {
			log.Fatalf("Error writing to the file: %s", error)
		}

		span.Finish()
		tracer.Stop()
	}
}

func main() {
	file, error := tail.TailFile(fileName, tail.Config{
		Follow: true,
	})

	if error != nil {
		log.Fatalf("Error reading %s\n", error)
	}

	go writer()

	for line := range file.Lines {
		tracer.Start(tracer.WithServiceName("log-reader"))

		log.Printf("My shared traceID %s\n", line.Text)

		mapCarrier := tracer.TextMapCarrier{
			tracer.DefaultParentIDHeader: line.Text,
			tracer.DefaultTraceIDHeader:  line.Text,
		}

		sctx, error := tracer.Extract(mapCarrier)

		if error != nil {
			log.Println("Invalid parent context")
			continue
		}

		span := tracer.StartSpan("read", tracer.ChildOf(sctx))

		span.SetTag("reader", true)
		span.SetTag(ext.AnalyticsEvent, true)

		time.Sleep(time.Second * 2)

		span.Finish()
		tracer.Stop()
	}
}
