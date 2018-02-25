package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	"server/game"
	"server/gamedata"
	"server/msg"
)

func handleMsg(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func init() {
	//版本
	handleMsg(&msg.C2S_Version{}, handlVersion)
	//注册
	handleMsg(&msg.C2S_Register{}, handleRegister)
	//登录
	handleMsg(&msg.C2S_Login{}, handleLogin)
}

//检测版本
func handlVersion(args []interface{}) {
	m := args[0].(*msg.C2S_Version)
	a := args[1].(gate.Agent)

	log.Debug("client Version %v", m.Ver)

	a.WriteMsg(&msg.S2C_Version_Ret{
		Ret: "1.0.0.1"})
}

//处理注册
func handleRegister(args []interface{}) {
	m := args[0].(*msg.C2S_Register)
	a := args[1].(gate.Agent)

	log.Debug("register AccID:%v PassWD:%v Sex:%v", m.AccID, m.PassWD, m.Sex)

	if len(m.AccID) < gamedata.AccIDMin || len(m.AccID) > gamedata.AccIDMax {
		log.Debug("账号不合法 AccID:%v", m.AccID)
		a.WriteMsg(&msg.S2C_Register_Ret{
			Ret: msg.S2C_Register_AccError})
		return
	}

	game.ChanRPC.Go("UserRegister", a, m.AccID, m.PassWD, m.Sex)
}

//处理登录
func handleLogin(args []interface{}) {
	m := args[0].(*msg.C2S_Login)
	a := args[1].(gate.Agent)

	log.Debug("login AccID:%v PassWD:%v", m.AccID, m.PassWD)

	if len(m.AccID) < gamedata.AccIDMin || len(m.AccID) > gamedata.AccIDMax {
		a.WriteMsg(&msg.S2C_Login_Ret{
			Ret: msg.S2C_Login_AccIDInvalid})
		return
	}

	// login
	game.ChanRPC.Go("UserLogin", a, m.AccID)
}