package handler

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttServer struct {
}

var (
	MqttClient mqtt.Client
)

// InitMqtt /**
/**
 * @author 大菠萝
 * @description //TODO 初始化mqtt的连接信息
 * @date 4:34 pm 9/7/22
 * @param
 * @return
 **/
func InitMqtt() {
	options := mqtt.ClientOptions{}
	options.AddBroker("tcp://127.0.0.1:18082")
	options.SetUsername("test")
	options.SetPassword("123456")
	//TODO 初始化连接实列
	MqttClient = mqtt.NewClient(&options)
	//TODO 连接超时的回调函数
	options.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("断开链接 -->>> %s", err.Error())
	}
	//TODO 连接成功的回调函数
	options.OnConnect = func(client mqtt.Client) {
		fmt.Printf("链接成功")
	}
	MqttClient.OptionsReader()
	if token := MqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("订阅失败")
	}
}

// Publish /**
/**
 * @author 大菠萝
 * @description //TODO 生产者发送消息
 * @date 4:37 pm 9/7/22
 * @param
 * @return
 **/
func Publish(msg string) {
	MqttClient.Publish("TOPIC", 2, true, msg)
}

/**
* @author 大菠萝
* @description
    //TODO 消费者订阅消息，并在回调函数里消费MQTT的消息。
    //TODO 这里如果不填回调函数，默认会在DefaultPublishHandler里去消费消息
* @date 4:39 pm 9/7/22
* @param
* @return
**/
func subCribe() {
	MqttClient.Subscribe("TOPIC", 2, func(client mqtt.Client, message mqtt.Message) {
		fmt.Printf("订阅的是topic --- %s %s", client, string(message.Payload()))
	})
}
