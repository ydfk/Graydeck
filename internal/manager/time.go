package manager

import (
	"os"
	"strings"
	"time"
)

const localTimeLayout = "2006-01-02 15:04:05"

func localNow() time.Time {
	return time.Now().In(localTimeLocation())
}

func formatLocalTime(value time.Time) string {
	return value.In(localTimeLocation()).Format(localTimeLayout)
}

func parseLocalTime(value string) (time.Time, error) {
	return time.ParseInLocation(localTimeLayout, value, localTimeLocation())
}

func localTimeLocation() *time.Location {
	name := strings.TrimSpace(os.Getenv("GRAYDECK_TIMEZONE"))
	if name == "" {
		name = strings.TrimSpace(os.Getenv("TZ"))
	}
	if name == "" {
		return time.Local
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}

	return location
}
