package main

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

func TimestampPB(t time.Time) *timestamp.Timestamp {
	ts := &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
	return ts
}
