/*
 * Copyright (C) 2019 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package bandwidth allows us to keep track of the consumer side connection speed.
package bandwidth

import (
	"testing"
	"time"

	"github.com/mysteriumnetwork/node/datasize"
	"github.com/stretchr/testify/assert"

	"github.com/mysteriumnetwork/node/core/connection"
)

func Test_ThroughputStringOutput(t *testing.T) {

}

func Test_ConsumeSessionEvent_ResetsOnConnect(t *testing.T) {
	tracker := Tracker{
		previous: connection.Statistics{
			At:            time.Now(),
			BytesReceived: 1,
			BytesSent:     1,
		},
	}
	tracker.ConsumeSessionEvent(connection.SessionEvent{
		Status: connection.SessionCreatedStatus,
	})

	assert.True(t, tracker.previous.At.IsZero())
	assert.Zero(t, tracker.previous.BytesReceived)
	assert.Zero(t, tracker.previous.BytesSent)
}

func Test_ConsumeSessionEvent_ResetsOnDisconnect(t *testing.T) {
	tracker := Tracker{
		previous: connection.Statistics{
			At:            time.Now(),
			BytesReceived: 1,
			BytesSent:     1,
		},
	}
	tracker.ConsumeSessionEvent(connection.SessionEvent{
		Status: connection.SessionEndedStatus,
	})

	assert.True(t, tracker.previous.At.IsZero())
	assert.Zero(t, tracker.previous.BytesReceived)
	assert.Zero(t, tracker.previous.BytesSent)
}

func Test_ConsumeStatisticsEvent_SkipsOnZero(t *testing.T) {
	tracker := Tracker{}
	e := connection.SessionStatsEvent{
		Stats: connection.Statistics{
			At:            time.Now(),
			BytesReceived: 1,
			BytesSent:     2,
		},
	}
	tracker.ConsumeStatisticsEvent(e)
	assert.False(t, tracker.previous.At.IsZero())
	assert.Equal(t, e.Stats.BytesReceived, tracker.previous.BytesReceived)
	assert.Equal(t, e.Stats.BytesSent, tracker.previous.BytesSent)
	assert.Zero(t, tracker.Get().Down.BitsPerSecond)
}

func Test_ConsumeStatisticsEvent_Regression_1674_InsaneSpeedReports(t *testing.T) {
	tracker := Tracker{}
	tracker.ConsumeStatisticsEvent(connection.SessionStatsEvent{
		Stats: connection.Statistics{
			At:            time.Now(),
			BytesSent:     0,
			BytesReceived: 0,
		},
	})
	tracker.ConsumeStatisticsEvent(connection.SessionStatsEvent{
		Stats: connection.Statistics{
			At:            time.Now(),
			BytesSent:     2048,
			BytesReceived: 2048,
		},
	})
	down := datasize.BitSize(tracker.Get().Down.BitsPerSecond)
	assert.Zero(t, down.Bytes())

	time.Sleep(time.Second)
	tracker.ConsumeStatisticsEvent(connection.SessionStatsEvent{
		Stats: connection.Statistics{
			At:            time.Now(),
			BytesSent:     4096,
			BytesReceived: 4096,
		},
	})
	down = datasize.BitSize(tracker.Get().Down.BitsPerSecond)
	assert.InDelta(t, 4096, down.Bytes(), 1024)
}
