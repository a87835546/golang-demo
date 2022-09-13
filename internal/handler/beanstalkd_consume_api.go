package handler

import (
	"fmt"
	"strings"
	"time"
)

func ReadingMsg() {
	id, body, err := BeanstalkdConsumeConn.Reserve(5 * time.Second)
	if err != nil {
		fmt.Printf("received beanstalkd err -->>> %s\n body -->>> %s \n  id--->>> %d \n", err.Error(), body, id)
		if err != nil && strings.Contains(err.Error(), "reserve-with-timeout: timeout") {
			fmt.Printf("time out")
		}
	} else {
		fmt.Printf("id --- >>> %v \n body --->>> %s \n", id, body)
		BeanstalkdConsumeConn.Delete(id)
	}
}
