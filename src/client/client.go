package main

import (
	"fmt"
	"encoding/binary"
	"net"
	"os"
)

var name string

func main () {
	fmt.Println("Client Start")

	fmt.Print("Input name : ")
	fmt.Scanln(&name)

	conn, err  := net.Dial("tcp", "127.0.0.1:3563")
	if err != nil {
		fmt.Print(err)
	}

	go DoOperator(conn)

	go SendVersion(conn)

	buf := make([]byte, 1024)

	for {
		len, err := conn.Read(buf)
		if  err != nil {
			conn.Close()
			fmt.Println("Server is dead ...ByeBye")
			os.Exit(0)
		}

		fmt.Println(string(buf[2:len]))
	}
}

func DoOperator(conn net.Conn) {
	var input string
	//username := conn.LocalAddr().String()

	for {
		fmt.Scanln(&input)

		if input == "Q" {
			fmt.Println("Quit")
			conn.Close()
			os.Exit(0)
		}

		if  input == "R" {
			fmt.Println("Register")

			go SendRegister(conn)
		}

		if input == "L" {
			fmt.Println("Login")

			go SendLoginMsg(conn)
		}

		if  input =="C" {
			fmt.Println("CreateTable")

			go SendCreateTable(conn)
		}

		if input == "J" {
			fmt.Println("JoinTable")

			go SendJoinTable(conn)
		}

		if input == "P" {
			fmt.Println("ReadyGame")

			go SendReady(conn)
		}

		if  input == "Out" {
			fmt.Println("OutCard")

			go SendOutCard(conn)
		}
	}
}

func SendRegister(conn net.Conn) {
	data := []byte(`{
		"C2S_Register": {
			"AccID": "zcwtop",
			"PassWD": "123456",
			"Sex": 1
		}
	}`)

	m := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(m, uint16(len(data)))
	copy(m[2:], data)
	conn.Write(m)
}

func SendLoginMsg(conn net.Conn) {
	if name == "0" {
		fmt.Println("zcwtop")

		data := []byte(`{
		"C2S_Login": {
			"AccID": "zcwtop",
			"PassWD": "123456"
		}
	}`)

		m := make([]byte, 2+len(data))

		binary.BigEndian.PutUint16(m, uint16(len(data)))

		copy(m[2:], data)

		conn.Write(m)
	} else {
		fmt.Println("abc")

		data := []byte(`{
		"C2S_Login": {
			"AccID": "abc",
			"PassWD": "123456"
		}
	}`)

		m := make([]byte, 2+len(data))
		binary.BigEndian.PutUint16(m, uint16(len(data)))
		copy(m[2:], data)
		conn.Write(m)
	}
}

func SendVersion(conn net.Conn) {
	data := []byte(`{
		"C2S_Version": {
			"Ver": "1.0.0.0"
		}
	}`)

	m := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(m, uint16(len(data)))
	copy(m[2:], data)
	conn.Write(m)
}

func SendCreateTable(conn net.Conn) {
	data := []byte(`{
		"C2S_Game_CreateTable": {
			"Type": 1
		}
	}`)
	m := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(m, uint16(len(data)))
	copy(m[2:], data)
	conn.Write(m)
}

func SendJoinTable(conn net.Conn) {
	data := []byte(`{
		"C2S_Game_JoinTable": {
			"TableNo": 1000
		}
	}`)

	m := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(m, uint16(len(data)))
	copy(m[2:], data)
	conn.Write(m)
}

func SendReady(conn net.Conn) {
	data := []byte(`{
		"C2S_Game_Ready": {
			"TableNo": 1000
		}
	}`)

	m := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(m, uint16(len(data)))
	copy(m[2:], data)
	conn.Write(m)
}

func SendOutCard(conn net.Conn) {
	data := []byte(`{
		"C2S_Game_OutCard": {
			"Card": 1
		}
	}`)

	m := make([]byte, 2+len(data))
	binary.BigEndian.PutUint16(m, uint16(len(data)))
	copy(m[2:], data)
	conn.Write(m)
}