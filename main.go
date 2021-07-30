package main

import (
	"encoding/json"
	"errors"
	"net"
)

type Msg struct {
	Event  string
	Name   string
	ToName string
	IP     string
	Port   int
	Msg    string
}

type Client struct {
	Name string
	Addr *net.UDPAddr
}

var socket *net.UDPConn
var clients = make(map[string]*Client)

func main() {
	var laddr, _ = net.ResolveUDPAddr("udp", ":13688")
	var err error
	socket, err = net.ListenUDP("udp", laddr)
	if err != nil {
		panic(errors.New("绑定socket失败:" + err.Error()))
	}
	println("启动监听:" + socket.LocalAddr().String())
	var bytes = make([]byte, 1024*10)
	for {
		var len, addr, err = socket.ReadFromUDP(bytes)
		if err != nil {
			println("receive失败", err)
			break
		}
		var content = string(bytes[0:len])
		println("receive", addr.String(), content)
		var msg = new(Msg)
		json.Unmarshal(bytes[0:len], msg)
		processMsg(addr, msg)
	}
}

func processMsg(addr *net.UDPAddr, msg *Msg) {
	switch msg.Event {
	case "reg":
		var client = &Client{
			Name: msg.Name,
			Addr: addr,
		}
		clients[msg.Name] = client
		send(addr, &Msg{
			Event: "reg",
			Name:  msg.Name,
			IP:    addr.IP.String(),
			Port:  addr.Port,
		})
	case "unreg":
		delete(clients, msg.Name)
		send(addr, &Msg{
			Event: "unreg",
		})
	case "touch":
		var client = clients[msg.Name]
		if client == nil {
			send(addr, &Msg{
				Event: "err",
				Msg:   "unreg",
			})
		} else {
			var remote = clients[msg.ToName]
			if remote == nil {
				send(addr, &Msg{
					Event:  "act",
					ToName: msg.ToName,
					Msg:    "offline",
				})
			} else {
				send(addr, &Msg{
					Event:  "touch",
					ToName: msg.ToName,
					IP:     remote.Addr.IP.String(),
					Port:   remote.Addr.Port,
				})
				send(remote.Addr, &Msg{
					Event:  "touch",
					ToName: client.Name,
					IP:     client.Addr.IP.String(),
					Port:   client.Addr.Port,
				})
			}
		}
	case "users":
		var array = make([]map[string]interface{}, len(clients))
		keys := make([]string, 0, len(clients))
		for k := range clients {
			keys = append(keys, k)
		}
		for i := 0; i < len(array); i++ {
			var key = keys[i]
			var client = clients[key]
			var m = make(map[string]interface{})
			m["Name"] = client.Name
			m["IP"] = client.Addr.IP.String()
			m["Port"] = client.Addr.Port
			array[i] = m
		}
		var bytes, _ = json.Marshal(array)
		send(addr, &Msg{
			Event: "users",
			Msg:   string(bytes),
		})
	}
}

func send(addr *net.UDPAddr, msg *Msg) {
	var bytes, _ = json.Marshal(msg)
	socket.WriteToUDP(bytes, addr)
}
