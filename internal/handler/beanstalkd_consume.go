package handler

import (
	"github.com/beanstalkd/go-beanstalk"
)

var BeanstalkdConsumeConn = &beanstalk.Conn{}

func init() {
	//addr := fmt.Sprintf("%s:%d", consts.Conf.Beanstalkd.Host, consts.Conf.Beanstalkd.Port)
	//fmt.Printf("add -->> %s", addr)
	//c, err := beanstalk.Dial("tcp", addr)
	//
	//if err != nil {
	//	fmt.Printf("consume connect beanstalkd err -->>> %s", err.Error())
	//}
	//go func() {
	//	for {
	//		id, body, err := c.Reserve(5 * time.Second)
	//		if err != nil {
	//			fmt.Printf("received beanstalkd err -->>> %s\n body -->>> %s \n  id--->>> %d \n", err.Error(), body, id)
	//			break
	//		} else {
	//			fmt.Printf("id --- >>> %v \n body --->>> %s \n", id, body)
	//			c.Delete(id)
	//		}
	//	}
	//}()
	//defer c.Close()

}
