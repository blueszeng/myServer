package gate

import (
	"server/login"
	"server/msg"
	"server/game"
)

func init() {
	//login
	msg.Processor.SetRouter(&msg.C2S_Version{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Register{}, login.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Login{}, login.ChanRPC)

	// game
	msg.Processor.SetRouter(&msg.C2S_Game_CreateTable{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Game_JoinTable{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.C2S_Game_Ready{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Game_OutCard{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.C2S_Game_OperateCard{}, game.ChanRPC)
}
