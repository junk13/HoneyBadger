/*
 *    metadata_attack_logger.go - HoneyBadger core library for detecting TCP attacks
 *    such as handshake-hijack, segment veto and sloppy injection.
 *
 *    Copyright (C) 2014  David Stainton
 *
 *    This program is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    This program is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package logging

import (
	"encoding/json"
	"fmt"
	"github.com/david415/HoneyBadger/types"
	"io"
	"os"
	"path/filepath"
)

// AttackMetadataJsonLogger is responsible for recording all attack reports as JSON objects in a file.
// This attack logger only logs metadata... but ouch code duplication.
type AttackMetadataJsonLogger struct {
	writer           io.WriteCloser
	LogDir           string
	stopChan         chan bool
	attackReportChan chan *types.Event
}

// NewAttackMetadataJsonLogger returns a pointer to a AttackMetadataJsonLogger struct
func NewAttackMetadataJsonLogger(logDir string) *AttackMetadataJsonLogger {
	a := AttackMetadataJsonLogger{
		LogDir:           logDir,
		stopChan:         make(chan bool),
		attackReportChan: make(chan *types.Event),
	}
	return &a
}

func (a *AttackMetadataJsonLogger) Start() {
	go a.receiveReports()
}

func (a *AttackMetadataJsonLogger) Stop() {
	a.stopChan <- true
}

func (a *AttackMetadataJsonLogger) receiveReports() {
	for {
		select {
		case <-a.stopChan:
			return
		case event := <-a.attackReportChan:
			a.SerializeAndWrite(event)
		}
	}
}

func (a *AttackMetadataJsonLogger) Log(event *types.Event) {
	a.attackReportChan <- event
}

func (a *AttackMetadataJsonLogger) SerializeAndWrite(event *types.Event) {
	publishableEvent := &serializedEvent{
		Type:         event.Type,
		Flow:         event.Flow.String(),
		HijackSeq:    event.HijackSeq,
		HijackAck:    event.HijackAck,
		Time:         event.Time,
		Start:        event.StartSequence,
		End:          event.EndSequence,
		OverlapStart: event.OverlapStart,
		OverlapEnd:   event.OverlapEnd,
	}
	a.Publish(publishableEvent)
}

// Publish writes a JSON report to the attack-report file for that flow.
func (a *AttackMetadataJsonLogger) Publish(event *serializedEvent) {
	b, err := json.Marshal(*event)
	a.writer, err = os.OpenFile(filepath.Join(a.LogDir, fmt.Sprintf("%s.metadata-attackreport.json", event.Flow)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}
	defer a.writer.Close()
	a.writer.Write([]byte(fmt.Sprintf("%s\n", string(b))))
}
