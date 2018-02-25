package internal

import (
	"github.com/name5566/leaf/go"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"github.com/name5566/leaf/util"
	"sync"
	"server/msg"
	"time"
	"container/list"
)

//桌子状态
const (
	TableStatus_Free	= 0
	TableStatus_Play	= 1
	TableStatus_End		= 2
)

//超时时间
const (
	OutCardTimerFlag	= 1
	TimeOut_OutCard		= 5

	OperatorTimerFlag	= 2
	TimeOut_Operator	= 5

	ReadyFlag			= 3
	TimeOut_Ready		= 5
)

//检测类型
const (
	Action_OutCard		= 1
	Action_GangCard		= 2
)

//结束类型
const (
	End_Normal			= 1
	End_LiuJu			= 2
)

var (
	gameTableMgr 		*GameTableMgr						//桌子管理
	tableNoMgr 			*TableNoMgr							//号码管理
)

func init () {
	gameTableMgr = new(GameTableMgr)
	gameTableMgr.Init()

	tableNoMgr = new(TableNoMgr)
	tableNoMgr.Init()
}

//号码管理
type TableNoMgr struct {
	TableNos  		 	*list.List							//号码队列
	l		   			sync.Mutex							//锁
}

func (mgr *TableNoMgr) Init() {
	mgr.TableNos = list.New()

	for i:=0; i<1000; i++ {
		mgr.TableNos.PushBack(i+1000)
	}
}

func (mgr *TableNoMgr) PushBack(tableNo int) {
	mgr.l.Lock()
	defer mgr.l.Unlock()

	mgr.TableNos.PushBack(tableNo)
}

func (mgr *TableNoMgr) PopFront() int {
	mgr.l.Lock()
	defer mgr.l.Unlock()

	iter := mgr.TableNos.Back()
	v := iter.Value
	mgr.TableNos.Remove(iter)

	return v.(int)
}

//桌子管理
type GameTableMgr struct {
	Tables 				*util.Map
}

func (mgr *GameTableMgr) Init() {
	mgr.Tables = new(util.Map)
}

//添加桌子
func (mgr *GameTableMgr) AddTable(gameType int, tableNo int) *Table{
	var table *Table
	if mgr.Tables.Get(tableNo) == nil {
		table = NewTable(gameType, tableNo)

		mgr.Tables.Set(tableNo, table)
	} else {
		log.Debug("AddTable tableNo:%v 已经存在", tableNo)

		table = mgr.Tables.Get(tableNo).(*Table)
	}

	return table
}

//获取桌子
func (mgr *GameTableMgr) GetTable(tableNo int) *Table {
	if mgr.Tables.Get(tableNo) != nil {
		return mgr.Tables.Get(tableNo).(*Table)
	} else {
		log.Debug("GetTable tableNo:%v 不存在", tableNo)
		return nil
	}
}

//删除桌子
func (mgr *GameTableMgr) DelTable(tableNo int) {
	if mgr.Tables.Get(tableNo) != nil {
		mgr.Tables.Del(tableNo)
	} else {
		log.Debug("DelTable tableNo:%v 不存在", tableNo)
	}
}

//游戏信息
type GameInfo struct {
	GameType 			int									//游戏类型
	TableNo				int									//桌子号码
	ChairID				int									//座位号码
}

func NewGameInfo(gameType int, tableNo int, chairID int) *GameInfo {
	gameInfo := new(GameInfo)
	gameInfo.GameType = gameType
	gameInfo.TableNo = tableNo
	gameInfo.ChairID = chairID

	return gameInfo
}

//创建桌子
func NewTable(gameType int, tableNo int) *Table {
	table := new(Table)
	table.init(gameType, tableNo)

	return table
}

//桌子结构
type Table struct {
	*g.LinearContext
	users				*util.Map							//玩家列表[pos]*User
	gameData			*GameData							//游戏数据
	tableNo				int									//桌子号码
	gameType 			int									//游戏类型
	playerCount			int									//游戏人数
	nowCount			int									//当前人数
	status          	int									//桌子状态
	readyInfo			[PlayerCount]bool					//准备状态

	outCardTimer		*timer.Timer						//出牌定时器
	operatorCardTimer	[PlayerCount]*timer.Timer			//操作定时器
	readyGameTimer		[PlayerCount]*timer.Timer			//准备定时器
}

func (table *Table) init(gameType int, tableNo int) {
	table.LinearContext = skeleton.NewLinearContext()
	table.users = new(util.Map)
	table.gameData = NewGameData()
	table.tableNo = tableNo
	table.gameType = gameType
	table.playerCount = PlayerCount
	table.status = TableStatus_Free
	table.nowCount = 0

	for i:=0; i<PlayerCount; i++ {
		table.readyInfo[i] = false
	}
}

//玩家进入桌子
func (table *Table) EnterTable(user *User, gameType int, tableNo int) {
	table.Go(func() {
		accID := user.Agent.UserData().(*AgentInfo).accID
		tmp := -1
		for i:=0; i<table.playerCount; i++ {
			if table.users.Get(i) == nil {
				log.Debug("accID:%v EnterTable sucess Pos:%v", accID, i)

				table.users.Set(i, user)
				table.nowCount++
				tmp = i
				break
			}
		}

		if tmp == -1 {
			return
		}

		gameInfo := NewGameInfo(gameType, tableNo, tmp)
		UsersGameInfo.Set(accID, gameInfo)

		//进入玩家的信息发送给其他人
		for i:=0; i<table.playerCount; i++ {
			if i == tmp {
				continue
			}

			if table.users.Get(i) != nil {
				other := table.users.Get(i).(*User)
				m := &msg.S2C_TablePlayerInfo{AccID: accID}
				other.WriteMsg(m)
			}
		}

		//其他人信息发送给进入玩家
		for i:=0; i<table.playerCount; i++ {
			if i == tmp {
				continue
			}

			if table.users.Get(i) != nil {
				other := table.users.Get(i).(*User)
				accID := other.Agent.UserData().(*AgentInfo).accID
				m := &msg.S2C_TablePlayerInfo{AccID: accID}
				user.WriteMsg(m)
			}
		}
	}, nil)
}

//玩家离开桌子
func (table *Table) LeaveTable(user *User) {
	if table.status == TableStatus_Play {
		//游戏过程中无法退出
		return
	}

	table.Go(func() {
		accID := user.Agent.UserData().(*AgentInfo).accID
		gameInfo := UsersGameInfo.Get(accID).(*GameInfo)
		log.Debug("accID:%v LeaveTable sucess Pos:%v", accID, gameInfo.ChairID)

		//广播给其他玩家离开消息
		for i:=0; i<table.playerCount; i++ {
			if i == gameInfo.ChairID {
				continue
			}

			if table.users.Get(i) != nil {
				other := table.users.Get(i).(*User)
				m := &msg.S2C_TablePlayerInfo{AccID: accID}
				other.WriteMsg(m)
			}
		}

		table.users.Del(gameInfo.ChairID)
		UsersGameInfo.Del(accID)
		table.nowCount--
	}, func(){
		if table.nowCount == 0 {
			log.Debug("房间解散")
		}
	})
}

//玩家准备
func (table *Table) ReadyGame(accID string, tableNo int) {
	if tableNo != table.tableNo {
		log.Debug("桌号不匹配，错误 sendNO:%v tableNO:%v", tableNo, table.tableNo)
		return
	}

	table.Go(func() {
		gameInfo := UsersGameInfo.Get(accID).(*GameInfo)
		table.readyInfo[gameInfo.ChairID] = true

		tmp := 0
		for i:=0; i<table.playerCount; i++ {
			if table.readyInfo[i] == true {
				tmp++
			}
		}

		if tmp == table.playerCount {
			table.StartGame()
		}
	}, nil)
}

//game start
func (table *Table) StartGame() {
	table.Go(func() {
		log.Debug("start game tableno:%v", table.tableNo)
		table.status = TableStatus_Play
		table.gameData.Init()

		bankerUser := table.gameData.bankerUser
		table.gameData.userAction[bankerUser] |= table.gameData.EstimateGang(bankerUser, 0)
		table.gameData.userAction[bankerUser] |= table.gameData.AnalyseHu(bankerUser, 0)

		//broadcast start msg
		for i:=0; i<table.playerCount; i++ {
			sendCard := table.gameData.sendCardData
			if i !=bankerUser {
				sendCard = 0
			}
			m := &msg.S2C_GameStart{
				Sice1: table.gameData.sice1,
				Sice2: table.gameData.sice2,
				GameTax: 0,
				BankerUser: bankerUser,
				UserAction: table.gameData.userAction[bankerUser],
				SendCard: sendCard,
				CurrentUser: bankerUser,
				CardIndex: table.gameData.UserCards(i)}

			table.users.Get(i).(*User).WriteMsg(m)
		}

		//set outcard timer
		table.StopTimer(OutCardTimerFlag, -1)
		table.outCardTimer = skeleton.AfterFunc(TimeOut_OutCard * time.Second, func() {
			table.Go(func() {
				//log.Debug("start game 开启出牌定时器 chairID:%v currentID:%v", table.gameData.bankerUser, table.gameData.currentUser)
				table.TimerOut_OutCard(table.gameData.bankerUser)
			}, nil)
		})
	}, nil)
}

//game end
func (table *Table) End(endType int) {
	log.Debug("table end")
}

//outcard timer out
func (table *Table) TimerOut_OutCard(chairID int) {
	if chairID != table.gameData.currentUser {
		log.Error("非法操作，不是该玩家出牌 outuser:%v current:%v", chairID, table.gameData.currentUser)
		return
	}

	card := table.GetACard(chairID)
	log.Debug("出牌定时器 user:%v card:%v", chairID, card)

	table.DoOutCard(chairID, card)
}

//outcard
func (table *Table) DoOutCard(chairID int, card int) {
	table.Go(func() {
		log.Debug("outcard user:%v card:%v", chairID, card)

		if chairID != table.gameData.currentUser {
			log.Error("非法操作，不是该玩家出牌 outuser:%v current:%v", chairID, table.gameData.currentUser)
			return
		}

		if table.gameData.IsValidCard(card) == false {
			log.Error("非法数据 outuser:%v card:%v", chairID, card)
			return
		}

		table.StopTimer(OutCardTimerFlag, -1)
		table.StopTimer(OperatorTimerFlag, -1)

		//broadcast outcard msg
		for i:=0; i<table.playerCount; i++ {
			m := &msg.S2C_Game_OutCard{
				OutUser: chairID,
				CardData: card,
				}
			table.users.Get(i).(*User).WriteMsg(m)
		}

		table.gameData.outCardCount++
		table.gameData.provideCard = card
		table.gameData.provideUser = chairID
		table.gameData.outCardData = card
		table.gameData.outCardUser = chairID
		table.gameData.gangStatus = false
		table.gameData.sendStatus = true

		table.gameData.currentUser = (chairID + PlayerCount - 1) % PlayerCount

		if table.CheckAction(chairID, card, Action_OutCard) == false {
			card := table.gameData.GetNextCard(chairID)
			if card == -1 {
				log.Debug("没有牌了，游戏结束")
				table.End(End_LiuJu)
				return
			}

			outChairID := table.gameData.outCardUser
			outCard := table.gameData.outCardData
			table.gameData.cardIndex[outChairID][table.gameData.switchToCardIndex(outCard)]--
			table.gameData.discardCard[outChairID][table.gameData.discardCount[outChairID]] = outCard
			table.gameData.discardCount[outChairID]++

			//log.Debug("current %v", table.gameData.currentUser)
			table.SendACard(table.gameData.currentUser, card, false)
		} else {
			table.SendOperateNotify()
		}
	}, nil)
}

func (table *Table) TimerOut_OperatorCard(chairID int, action int, card int) {
	log.Debug("操作定时器 user:%v card:%v action:%v", chairID, action, card)

	table.DoOperator(chairID, action, card)
}

func (table *Table) DoOperator(chairID int, action int, card int) {
	table.Go(func() {
		if table.gameData.currentUser != -1 && chairID != table.gameData.currentUser {
			log.Error("非法操作，不是该玩家出牌 outuser:%v current:%v", chairID, table.gameData.currentUser)
			return
		}

		log.Debug("operator user:%v action:%v card:%v", chairID, action, card)

		//其他人操作
		if table.gameData.currentUser == -1 {
			if table.gameData.response[chairID] == true {
				log.Error("table.gameData.response[%v] == true", chairID)
				return
			}

			if table.gameData.userAction[chairID] == 0 {
				log.Error("table.gameData.userAction[%v] == 0", chairID)
				return
			}

			if (action != 0) && (action&table.gameData.userAction[chairID]) == 0 {
				log.Error("action:%v 不存在的操作", action)
				return
			}

			table.StopTimer(OperatorTimerFlag, chairID)

			targetUser := chairID
			targetAction := action

			table.gameData.response[chairID] = true
			table.gameData.performAction[chairID] = action
			table.gameData.operateCard[chairID] = card

			//判断优先级
			for i:=0; i<table.playerCount; i++ {
				tmpAction := 0
				if table.gameData.response[i] == false {
					tmpAction = table.gameData.userAction[i]
				} else {
					tmpAction = table.gameData.performAction[i]
				}

				userActionRank := table.gameData.ActionRank(tmpAction)
				targetActionRank := table.gameData.ActionRank(targetAction)

				if userActionRank > targetActionRank {
					targetUser = i
					targetAction = tmpAction
				}
			}

			if table.gameData.response[targetUser] == false {
				return
			}

			//等待其他玩家的胡牌操作
			if targetAction == Wik_Hu {
				for i:=0; i<table.playerCount; i++ {
					if table.gameData.response[i] == false && table.gameData.userAction[i]&Wik_Hu != 0 {
						return
					}
				}
			}

			if targetAction == Wik_Null {
				card := table.gameData.GetNextCard(chairID)
				if card == -1 {
					log.Debug("没有牌了，游戏结束")
					table.End(End_LiuJu)
					return
				}

				table.SendACard(table.gameData.currentUser, card, false)

				return
			}

			targetCard := table.gameData.operateCard[targetUser]
			table.gameData.outCardData = 0
			table.gameData.outCardUser = -1
			table.gameData.sendStatus = true

			if targetAction == Wik_Hu {
				table.gameData.huCard = targetCard
				table.gameData.huUser = targetUser
				for i:=0; i<table.playerCount; i++ {
					if i == table.gameData.provideUser || table.gameData.performAction[targetUser]&Wik_Hu == 0 {
						continue
					}

					if table.gameData.userAction[i]&targetAction != 0 {
						table.gameData.cardIndex[i][table.gameData.switchToCardIndex(targetCard)]++
						targetUser=i
						table.gameData.huUser = targetUser
					}
				}

				table.End(End_Normal)

				return
			}

			//删除数据
			if targetAction == Wik_Left {
				removeCard1 := targetCard+1
				removeCard2 := targetCard+2
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(removeCard1)]--
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(removeCard2)]--

			} else if targetAction == Wik_Center {
				removeCard1 := targetCard-1
				removeCard2 := targetCard+1
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(removeCard1)]--
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(removeCard2)]--
			} else if targetAction == Wik_Right {
				removeCard1 := targetCard+1
				removeCard2 := targetCard+2
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(removeCard1)]--
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(removeCard2)]--
			} else if targetAction == Wik_Peng {
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(targetCard)]--
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(targetCard)]--
			} else if targetAction == Wik_Gang {
				table.gameData.cardIndex[targetUser][table.gameData.switchToCardIndex(targetCard)] = 0
			}

			if table.gameData.weaveItemCount[targetUser]>=MaxWeaveCount {
				log.Error("table.gameData.weaveItemCount[%v] = %v", targetUser, table.gameData.weaveItemCount[targetUser])
				return
			}

			index := table.gameData.weaveItemCount[targetUser]
			table.gameData.weaveItems[index] = NewWeaveItem(targetAction, targetCard, table.gameData.provideUser, true)
			table.gameData.weaveItemCount[targetUser]++

			for i:=0; i<table.playerCount; i++ {
				m := &msg.S2C_Game_OperateResult{
					OperateUser: targetUser,
					ProvideUser: table.gameData.provideUser,
					OperateAction: targetAction,
					OperateCard: targetCard}

				table.users.Get(i).(*User).WriteMsg(m)
			}

			table.gameData.currentUser = targetUser
			if targetAction==Wik_Gang {
				table.gameData.gangStatus = true

				card := table.gameData.GetNextCard(chairID)
				if card == -1 {
					log.Debug("没有牌了，游戏结束")
					table.End(End_LiuJu)
					return
				}

				table.SendACard(targetUser, card, true)
			} else {
				//设置出牌定时器
				table.outCardTimer = skeleton.AfterFunc(TimeOut_OutCard * time.Second, func() {
					table.Go(func() {
						table.TimerOut_OutCard(chairID)
					}, nil)
				})
			}

			return
		}

		//自己操作
		if table.gameData.currentUser == chairID {
			table.gameData.sendStatus = true
			table.gameData.userAction[chairID] = 0
			table.gameData.performAction[chairID] = 0

			table.StopTimer(OperatorTimerFlag, chairID)
			table.StopTimer(OutCardTimerFlag, -1)

			publicType := true
			if action == Wik_Gang {
				if table.gameData.cardIndex[chairID][table.gameData.switchToCardIndex(card)] == 1 {
					index := -1
					for i:=0; i<table.gameData.weaveItemCount[chairID]; i++ {
						if table.gameData.weaveItems[i].weaveKind == Wik_Peng && table.gameData.weaveItems[i].centerCard == card {
							publicType = true
							index = i
						}
					}

					if index == -1{
						log.Error("不存在的明杠 card:%v", card)
						return
					}

					table.gameData.weaveItems[index].weaveKind = Wik_Gang
					table.gameData.weaveItems[index].publicCard = publicType
				} else {
					if table.gameData.cardIndex[chairID][table.gameData.switchToCardIndex(card)] != 4 {
						log.Error("不存在的暗杠 card:%v", card)
						return
					}
					publicType = false
					index := table.gameData.weaveItemCount[chairID]
					table.gameData.weaveItems[index] = NewWeaveItem(Wik_Gang, card, chairID, false)
					table.gameData.weaveItemCount[chairID]++
				}

				table.gameData.cardIndex[chairID][table.gameData.switchToCardIndex(card)] = 0
				table.gameData.gangStatus = true

				for i:=0; i<table.playerCount; i++ {
					m := &msg.S2C_Game_OperateResult{
						OperateUser: chairID,
						ProvideUser: chairID,
						OperateAction: action,
						OperateCard: card}

					table.users.Get(i).(*User).WriteMsg(m)
				}

				//判断是否有抢杠
				aroseAction := false
				if publicType == true {
					aroseAction = table.CheckAction(chairID, card, Action_GangCard)
				}

				if aroseAction == false {
					table.gameData.gangStatus = true

					card := table.gameData.GetNextCard(chairID)
					if card == -1 {
						log.Debug("没有牌了，游戏结束")
						table.End(End_LiuJu)
						return
					}

					table.SendACard(chairID, card, true)
				}

			} else if action == Wik_Hu {
				table.gameData.huUser = chairID
				table.gameData.huCard = card

				table.End(End_Normal)
			}
		}

	}, nil)
}

func (table *Table) SendACard(chairID int, card int, tail bool) {
	table.Go(func() {
		table.gameData.userAction[chairID] = 0
		if table.gameData.forbidHu[chairID] == false {
			table.gameData.userAction[chairID]|=table.gameData.AnalyseHu(chairID, 0)
		}
		if table.gameData.forbidCPG[chairID] == false && table.gameData.leftCardCount < EndLeftCount {
			table.gameData.userAction[chairID]|=table.gameData.EstimateGang(chairID, 0)
		}

		for i:=0; i<table.playerCount; i++ {
			tmpCard := 0
			tmpAction := 0
			if i == chairID {
				tmpCard = card
				tmpAction = table.gameData.userAction[chairID]
			}

			m := &msg.S2C_Game_SendCard{
				SendUser: chairID,
				CurrentUser: chairID,
				CardData: tmpCard,
				UserAction: tmpAction,
				LeftCount: table.gameData.leftCardCount,
				Tail: tail}

			table.users.Get(i).(*User).WriteMsg(m)
		}
		//设置出牌定时器
		table.StopTimer(OutCardTimerFlag, -1)
		table.StopTimer(OperatorTimerFlag, -1)
		table.outCardTimer = skeleton.AfterFunc(TimeOut_OutCard * time.Second, func() {
			table.Go(func() {
				//log.Debug("SendACard 开启出牌定时器 chairID:%v currentID:%v", chairID, table.gameData.currentUser)
				table.TimerOut_OutCard(chairID)
			}, nil)
		})
	}, nil)
}

//从自己牌中选取一张
func (table *Table) GetACard(chairID int) int {
	cards := table.gameData.UserCards(chairID)

	card := cards[0]
	for i:=0; i<MaxCount; i++ {
		if cards[i] != 0 {
			card =  cards[i]
			break;
		}
	}

	return card
}

func (table *Table) CheckAction(chairID int, card int, actionType int) bool {
	for i:=0; i<table.playerCount; i++ {
		table.gameData.userAction[i] = 0
		table.gameData.performAction[i] = 0
		table.gameData.response[i] = false
	}

	return false

	hasAction := false
	for  i:=0; i<table.playerCount; i++ {
		if i==chairID {
			continue
		}

		if actionType == Action_OutCard {
			if table.gameData.forbidCPG[i] == false {
				table.gameData.userAction[i]|=table.gameData.EstimatePeng(i, card)

				eatUser := (chairID+PlayerCount-1)%PlayerCount
				if eatUser == i {
					table.gameData.userAction[i]|=table.gameData.EstimateEat(i, card)
				}

				if table.gameData.leftCardCount > EndLeftCount {
					table.gameData.userAction[i]|=table.gameData.EstimateGang(i, card)
				}
			}
		}

		if table.gameData.forbidHu[i] == false {
			table.gameData.userAction[i]|=table.gameData.AnalyseHu(i, card)
		}

		if table.gameData.userAction[i] != Wik_Null {
			hasAction = true
			table.gameData.provideUser = chairID
			table.gameData.provideCard = card
			table.gameData.resumeUser = table.gameData.currentUser
			table.gameData.currentUser = -1
		}
	}

	return hasAction
}

func (table *Table) SendOperateNotify(){
	/*for i:=0; i<table.playerCount; i++ {
		if table.gameData.userAction[i] != 0 {
			m := &msg.S2C_Game_OperateNotify{
				ResumeUser: table.gameData.resumeUser,
				UserAction: table.gameData.userAction[i],
				CardData: table.gameData.provideCard}

			table.users.Get(i).(*User).WriteMsg(m)

			//设置操作定时器
			table.StopTimer(OperatorTimerFlag, i)

			log.Debug("1 i:%v", i)
			table.operatorCardTimer[i] = skeleton.AfterFunc(TimeOut_Operator * time.Second, func() {
				table.Go(func() {
					log.Debug("2 i:%v", i)
					table.TimerOut_OperatorCard(i, 0, table.gameData.provideCard)
				}, nil)
			})
		}
	}*/

	//0
	if table.gameData.userAction[0] != 0 {
		m := &msg.S2C_Game_OperateNotify{
			ResumeUser: table.gameData.resumeUser,
			UserAction: table.gameData.userAction[0],
			CardData: table.gameData.provideCard}

		table.users.Get(0).(*User).WriteMsg(m)

		//设置操作定时器
		table.StopTimer(OperatorTimerFlag, 0)

		table.operatorCardTimer[0] = skeleton.AfterFunc(TimeOut_Operator * time.Second, func() {
			table.Go(func() {
				table.TimerOut_OperatorCard(0, 0, table.gameData.provideCard)
			}, nil)
		})
	}

	//1
	if table.gameData.userAction[1] != 0 {
		m := &msg.S2C_Game_OperateNotify{
			ResumeUser: table.gameData.resumeUser,
			UserAction: table.gameData.userAction[1],
			CardData: table.gameData.provideCard}

		table.users.Get(1).(*User).WriteMsg(m)

		//设置操作定时器
		table.StopTimer(OperatorTimerFlag, 1)

		table.operatorCardTimer[1] = skeleton.AfterFunc(TimeOut_Operator * time.Second, func() {
			table.Go(func() {
				table.TimerOut_OperatorCard(1, 0, table.gameData.provideCard)
			}, nil)
		})
	}
}

func (table *Table) StopTimer(timerType int, chairID int) {
	if timerType == OutCardTimerFlag && table.outCardTimer != nil {
		table.outCardTimer.Stop()
	} else if timerType == OperatorTimerFlag {
		if chairID == -1 {
			for i:=0; i<table.playerCount; i++ {
				if table.operatorCardTimer[i] != nil {
					table.operatorCardTimer[i].Stop()
				}
			}
		} else {
			if table.operatorCardTimer[chairID] != nil {
				table.operatorCardTimer[chairID].Stop()
			}
		}
	} else if timerType == ReadyFlag {
		if chairID == -1 {
			for i:=0; i<table.playerCount; i++ {
				if table.readyGameTimer[i] != nil {
					table.readyGameTimer[i].Stop()
				}
			}
		} else {
			if table.readyGameTimer[chairID] != nil {
				table.readyGameTimer[chairID].Stop()
			}
		}
	}
}