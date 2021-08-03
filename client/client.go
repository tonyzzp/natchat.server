package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Msg struct {
	Event  string
	Name   string
	ToName string
	IP     string
	Port   int
	Msg    string
}

const HELP = `
reg name : 注册
unreg name : 取消注册
users : 用户列表
peers : 本地用户列表
touch name : 联系
send name msg : 发送原始消息
chat name msg : 发消息
`

var serverAddr, _ = net.ResolveUDPAddr("udp", "192.168.199.144:13688")

var socket *net.UDPConn
var clients = make(map[string]string)
var myName = ""

func Start() {
	socket, _ = net.ListenUDP("udp", nil)
	go listen()
	help()
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var cmds = strings.Split(scanner.Text(), " ")
		var msg = Msg{}
		switch cmds[0] {
		case "reg":
			myName = cmds[1]
			msg.Event = "reg"
			msg.Name = myName
			sendToServer(msg)
		case "unreg":
			msg.Event = "unreg"
			msg.Name = cmds[1]
			sendToServer(msg)
		case "users":
			msg.Event = "users"
			sendToServer(msg)
		case "touch":
			msg.Event = "touch"
			msg.Name = myName
			msg.ToName = cmds[1]
			sendToServer(msg)
		case "send":
			sendTo(cmds[1], cmds[2])
		case "chat":
			msg.Event = "chat"
			msg.Name = myName
			msg.Msg = strings.Join(cmds[2:], " ")
			sendToClient(cmds[1], msg)
		case "peers":
			for k, v := range clients {
				fmt.Printf("%s:%s\n", k, v)
			}
		}
	}
}

func listen() {
	for {
		var bytes = make([]byte, 1024*10)
		var len, addr, err = socket.ReadFromUDP(bytes)
		if err != nil {
			fmt.Println("receive错误:" + err.Error())
			break
		}
		var content = string(bytes[:len])
		fmt.Println("receive:" + addr.String() + " " + content)
		var msg = Msg{}
		json.Unmarshal(bytes[:len], &msg)
		switch msg.Event {
		case "touch":
			if msg.Port > 0 {
				clients[msg.Name] = msg.IP + ":" + strconv.Itoa(msg.Port)
				go touch(msg.Name)
			}
		}
	}
}

func touch(name string) {
	var ticker = time.NewTicker(time.Second * 1)
	var times = 0
	for {
		<-ticker.C
		times++
		sendToClient(name, Msg{Event: "touch", Name: myName})
		if times == 2 {
			ticker.Stop()
			break
		}
	}
}

func sendToClient(name string, msg interface{}) {
	var bytes, _ = json.Marshal(msg)
	var addr, _ = net.ResolveUDPAddr("udp", clients[name])
	socket.WriteToUDP(bytes, addr)
	fmt.Println("sendToClient", addr.String(), string(bytes))
}

func sendTo(name string, msg string) {
	var bytes, _ = json.Marshal(msg)
	var addr, _ = net.ResolveUDPAddr("udp", clients[name])
	socket.WriteToUDP(bytes, addr)
}

func sendToServer(msg interface{}) {
	var bytes, _ = json.Marshal(msg)
	socket.WriteToUDP(bytes, serverAddr)
}

func help() {
	fmt.Println(HELP)
}
