package xtime

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}

func ProtoTimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

func TimeToProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func ProtoTimestampToTimestamp(ts *timestamppb.Timestamp) int64 {
	if ts == nil {
		return 0
	}
	return ts.AsTime().Unix()
}
func TimestampToProtoTimestamp(timestamp int64) *timestamppb.Timestamp {
	return timestamppb.New(TimestampToTime(timestamp))
}
