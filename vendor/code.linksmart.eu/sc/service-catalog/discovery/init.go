// Copyright 2014-2016 Fraunhofer Institute for Applied Information Technology FIT

package discovery

import (
	"log"
	"os"
	"strconv"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout, "[discovery] ", 0)

	v, err := strconv.Atoi(os.Getenv("DEBUG"))
	if err == nil && v == 1 {
		logger.SetFlags(log.Ltime | log.Lshortfile)
	}
}
