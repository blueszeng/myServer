package internal

import (
	"github.com/name5566/leaf/gate"
	"reflect"
	"server/msg"
	"github.com/name5566/leaf/log"
)

func handleMsg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), func(args []interface{}) {
		// user
		a := args[1].(gate.Agent)
		//user := users[a.UserData().(*AgentInfo).userID]
		user := UserAccIDMgr.Get(a.UserData().(*AgentInfo).accID)
		if user == nil {
			log.Debug("handleMsg error 账号不存在")
			return
		}

		// agent to user
		args[1] = user.(*User)
		h.(func([]interface{}))(args)
	})
}

func init() {
	handleMsg(&msg.C2S_Game_CreateTable{}, handleCreateTable)
	handleMsg(&msg.C2S_Game_JoinTable{}, handleJoinTable)
	handleMsg(&msg.C2S_Game_LeaveTable{}, handleLeaveTable)

	handleMsg(&msg.C2S_Game_Ready{}, handleReady)
	handleMsg(&msg.C2S_Game_OutCard{}, handleOutCard)
	handleMsg(&msg.C2S_Game_OperateCard{}, handleOperator)
}

//创建桌子
func handleCreateTable(args []interface{}) {
	m := args[0].(*msg.C2S_Game_CreateTable)
	user := args[1].(*User)

	user.createTable(m.Type)
}

//加入桌子
func handleJoinTable(args []interface{}) {
	m := args[0].(*msg.C2S_Game_JoinTable)
	user := args[1].(*User)

	user.joinTable(m.TableNo)
}

//离开桌子
func handleLeaveTable(args []interface{}) {
	m := args[0].(*msg.C2S_Game_LeaveTable)
	user := args[1].(*User)

	user.leaveTable(m.TableNo)
}

//玩家准备
func handleReady(args []interface{}) {
	m := args[0].(*msg.C2S_Game_Ready)
	user := args[1].(*User)

	user.ready(m.TableNo)
}

//玩家出牌
func handleOutCard(args []interface{}) {
	m := args[0].(*msg.C2S_Game_OutCard)
	user := args[1].(*User)

	user.outCard(m.Card)
}

//玩家操作
func handleOperator(args []interface{}) {
	m := args[0].(*msg.C2S_Game_OperateCard)
	user := args[1].(*User)

	user.operatorCard(m.OperateAction, m.OperateCard)
}