package cache

import (
	"errors"

	"github.com/DIGIREN/logsAPI/models"
)

var LogById map[int]*models.LogObject = make(map[int]*models.LogObject)

//Stores a log in memory, or if the user exists, updates the user's login info
func CacheLog(log models.LogObject) error {
	if log.UserID == 0 {
		return errors.New("UserID is required")
	}
	//If we already have a log for this user, we need to append the login stamp to it
	if _, ok := LogById[log.UserID]; !ok {
		LogById[log.UserID] = &log
	} else {
		LogById[log.UserID].Meta.Logins = append(LogById[log.UserID].Meta.Logins, log.Meta.Logins...)
	}
	return nil
}

//Returns all logs in memory
func GetAllLogs() map[int]*models.LogObject {
	return LogById
}

//Removes all logs in memory
func EmptyAll() {
	LogById = make(map[int]*models.LogObject)
}
