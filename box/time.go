package box

import "time"

// TimeNowMs returns current timestamp in milliseconds.
func TimeNowMs() uint64 {
	return TimeNowUs() / 1000
}

// TimeNowUs returns current timestamp in microseconds.
func TimeNowUs() uint64 {
	return uint64(time.Now().UnixNano() / 1000)
}

// TimeNowStr returns current time in a string of
// format : "2006-01-02 15:04:05"
func TimeNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// TimeToStr returns the given time, including seconds and nano seconds,
// in a string of format : "2006-01-02 15:04:05"
func TimeToStr(sec, nsec int64) string {
	return time.Unix(sec, nsec).Format("2006-01-02 15:04:05")
}

// TimeSecToStr returns the given time, represented by seconds,
// in a string of format : "2006-01-02 15:04:05"
func TimeSecToStr(sec int64) string {
	return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
}
