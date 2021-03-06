package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"text/template"
)

type UploadData struct {
	FileName string
	FileSize int64
	Events   chan UploadEvent
}

type UploadEvent struct {
	EventSeq  int64
	EventText string
}

func convertFileIntoTemplate(fp *os.File, events chan<- UploadEvent, firstEventTemplate, subsequentEventTemplate *template.Template) {
	defer close(events)

	var fileReader io.ReadCloser
	var err error

	if IsGzip(fp) {
		fileReader, err = gzip.NewReader(fp)
		if err != nil {
			log.Debugf("Error reading file: %s", err.Error())
			MoveFileToDebug(fp.Name())
			return
		}
		defer fileReader.Close()
	} else {
		fileReader = fp
	}

	scanner := bufio.NewScanner(fileReader)
	var i int64

	for scanner.Scan() {
		var b bytes.Buffer
		var err error
		eventText := scanner.Text()

		if len(eventText) == 0 {
			// skip empty lines
			continue
		}

		if config.CommaSeparateEvents {
			if i == 0 {
				err = firstEventTemplate.Execute(&b, eventText)
			} else {
				err = subsequentEventTemplate.Execute(&b, eventText)
			}
			eventText = b.String()
		} else {
			eventText = eventText + "\n"
		}
		if err != nil {
			log.Debug(err)
		}

		events <- UploadEvent{EventText: eventText, EventSeq: i}
		i++

	}

	if err := scanner.Err(); err != nil {
		log.Debug(err)
	}

}
