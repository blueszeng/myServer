package msg

import (
	"github.com/name5566/leaf/network/json"
)

const (
	MaxCount 			= 14								//最大数量
)

var Processor = json.NewProcessor()

func init() {
	Processor.Register(&C2S_Version{})
	Processor.Register(&S2C_Version_Ret{})

	Processor.Register(&S2C_Close{})

	Processor.Register(&C2S_Register{})
	Processor.Register(&S2C_Register_Ret{})

	Processor.Register(&C2S_Login{})
	Processor.Register(&S2C_Login_Ret{})

	Processor.Register(&C2S_Game_CreateTable{})
	Processor.Register(&S2C_Game_CreateTable_Ret{})

	Processor.Register(&C2S_Game_JoinTable{})
	Processor.Register(&S2C_Game_JoinTable_Ret{})

	Processor.Register(&C2S_Game_LeaveTable{})

	Processor.Register(&S2C_TablePlayerInfo{})

	Processor.Register(&C2S_Game_Ready{})

	Processor.Register(&S2C_GameStart{})

	Processor.Register(&C2S_Game_OutCard{})
	Processor.Register(&C2S_Game_OperateCard{})

	Processor.Register(&S2C_Game_OutCard{})
	Processor.Register(&S2C_Game_SendCard{})
	Processor.Register(&S2C_Game_OperateNotify{})
	Processor.Register(&S2C_Game_OperateResult{})
	Processor.Register(&S2C_Game_End{})
}


const (
	S2C_Close_LoginRepeated 			= 1
	S2C_Close_InnerError    			= 2

	S2C_Register_Ok						= 0
	S2C_Register_ExistError				= 1
	S2C_Register_AccError				= 2

	S2C_Login_Ok						= 0
	S2C_Login_NoAccError				= 1
	S2C_Login_PassWDError				= 2
	S2C_Login_AccIDInvalid				= 3

	S2C_Game_CreateTable_Ok    		 	= 0
	S2C_Game_CreateTable_Error  		= 1

	S2C_Game_JoinTable_Ok				= 0
	S2C_Game_JoinTable_Error			= 1
)

// Close
type S2C_Close struct {
	Err 				int
}

type C2S_Version struct {
	Ver 				string
}
type S2C_Version_Ret struct {
	Ret 				string
}

//Register
type C2S_Register struct {
	AccID				string
	PassWD 				string
	Sex  				int
}
type S2C_Register_Ret struct {
	Ret					int
}

//Login
type C2S_Login struct {
	AccID				string
	PassWD 				string
}
type S2C_Login_Ret struct {
	Ret					int
}

//创建桌子
type C2S_Game_CreateTable struct {
	Type				int
}
type S2C_Game_CreateTable_Ret struct {
	Ret  				int
}

//加入桌子
type C2S_Game_JoinTable struct {
	TableNo     		int
}
type S2C_Game_JoinTable_Ret struct {
	Ret  				int
}

//离开桌子
type C2S_Game_LeaveTable struct {
	TableNo				int
}

//用户信息
type S2C_TablePlayerInfo struct {
	AccID				string
}

//游戏开始
type S2C_GameStart struct {
	Sice1				int									//色子
	Sice2				int									//色子
	GameTax				int									//税收
	BankerUser			int									//庄家
	UserAction			int									//用户动作
	SendCard			int									//发送数据
	CurrentUser			int									//当前用户
	CardIndex			[MaxCount]int						//用户数据
}

//出牌
type S2C_Game_OutCard struct {
	OutUser				int									//出牌玩家
	CardData			int									//出牌数据

}

//发牌
type S2C_Game_SendCard struct {
	SendUser			int									//抓牌玩家
	CurrentUser			int									//当前玩家
	CardData			int									//发牌数据
	UserAction			int									//用户动作
	LeftCount			int									//剩余数量
	Tail				bool								//杠后发牌
}

//操作提示
type S2C_Game_OperateNotify struct{
	ResumeUser			int									//还原玩家
	UserAction			int									//用户动作
	CardData			int									//操作数据
}

//操作结果
type S2C_Game_OperateResult struct {
	OperateUser			int									//操作玩家
	ProvideUser			int									//提供玩家
	OperateAction		int									//用户动作
	OperateCard			int									//操作数据
}

//结算
type S2C_Game_End struct {

}

//用户准备
type C2S_Game_Ready struct {
	TableNo				int									//桌子号码
}

//出牌
type C2S_Game_OutCard struct {
	Card 				int
}

//操作
type C2S_Game_OperateCard struct {
	OperateAction		int									//操作动作
	OperateCard			int									//操作数据
}


