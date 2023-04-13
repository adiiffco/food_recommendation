package utils

import "time"

func getCurrentTimeIST() time.Time {
	loc, _ := time.LoadLocation("Asia/Kolkata")
	return time.Now().In(loc)
}
func GetTimeDifferenceInHour(t time.Time) int64 {
	return int64(getCurrentTimeIST().Sub(t).Hours())
}
