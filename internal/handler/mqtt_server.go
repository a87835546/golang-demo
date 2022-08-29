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

func InitMqtt() {
	options := mqtt.ClientOptions{}
	options.AddBroker("tcp://127.0.0.1:18082")
	options.SetUsername("test")
	options.SetPassword("123456")
	MqttClient = mqtt.NewClient(&options)
	options.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("断开链接 -->>> %s", err.Error())
	}
	options.OnConnect = func(client mqtt.Client) {
		fmt.Printf("链接成功")
	}
	MqttClient.OptionsReader()
	if token := MqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Printf("订阅失败")
	}
}

func Publish(msg string) {
	MqttClient.Publish("TOPIC", 2, true, msg)
}
func subCribe() {
	MqttClient.Subscribe("TOPIC", 2, func(client mqtt.Client, message mqtt.Message) {
		fmt.Printf("订阅的是topic --- %s %s", client, string(message.Payload()))
	})
}
