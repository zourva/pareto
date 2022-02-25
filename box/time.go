package box

import "time"

func TimeNowMs() uint64 {
	return TimeNowUs() / 1000
}

func TimeNowUs() uint64 {
	return uint64(time.Now().UnixNano() / 1000)
}

// format : "2006-01-02 15:04:05"
func TimeNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func TimeToStr(sec, nsec int64) string {
	return time.Unix(sec, nsec).Format("2006-01-02 15:04:05")
}

func TimeSecToStr(sec int64) string {
	return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
}