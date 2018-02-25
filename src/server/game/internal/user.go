package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/go"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/timer"
	"github.com/name5566/leaf/util"
	"server/msg"
	"time"
)

const (
	userLogin = iota
	userLogout
	userGame
	userOffline
)

type User struct {
	gate.Agent
	*g.LinearContext
	state       		int
	userInfo    		*UserInfo
	saveDBTimer 		*timer.Timer
}

func (user *User) login() {
	// network closed
	if user.state == userLogout {
		user.logout()
		return
	}

	user.onLogin()
	user.autoSaveDB()
}

func (user *User) logout() {
	if user.userInfo != nil {
		user.saveDBTimer.Stop()
		user.onLogout()
	}

	// save
	data := util.DeepClone(user.userInfo)
	user.Go(func() {
		if data != nil {
			db := mongoDB.Ref()
			defer mongoDB.UnRef(db)
			userID := data.(*UserInfo).UserID
			_, err := db.DB("game").C("users").UpsertId(userID, data)
			if err != nil {
				log.Error("save user %v data error: %v", userID, err)
			}
		}

		UserAccIDMgr.Del(user.userInfo.AccID)
		UserIDMgr.Del(user.userInfo.UserID)
	}, nil)
}

func (user *User) autoSaveDB() {
	const duration = 5 * time.Minute

	// save
	user.saveDBTimer = skeleton.AfterFunc(duration, func() {
		data := util.DeepClone(user.userInfo)
		user.Go(func() {
			db := mongoDB.Ref()
			defer mongoDB.UnRef(db)
			userID := data.(*UserInfo).UserID
			_, err := db.DB("game").C("users").UpsertId(userID, data)
			if err != nil {
				log.Error("save user %v data error: %v", userID, err)
			}
		}, func() {
			user.autoSaveDB()
		})
	})
}

func (user *User) isOffline() bool {
	return user.state == userLogout
}

func (user *User) onLogin() {
	log.Debug("login sucess: %v", user.userInfo.AccID)
	m := &msg.S2C_Login_Ret{Ret: msg.S2C_Login_Ok}
	user.WriteMsg(m)
}

func (user *User) onLogout() {
	log.Debug("loginout sucess: %v", user.userInfo.AccID)
}

//玩家创建桌子
func (user *User) createTable(gameType int) {
	user.Go(func() {
		tableNo := tableNoMgr.PopBack()
		table := gameTableMgr.AddTable(gameType, tableNo)

		log.Debug("createTable gameType:%v tableNo:%v", gameType, tableNo)

		log.Debug("createTable sucess")
		user.WriteMsg(&msg.S2C_Game_CreateTable_Ret{Ret: msg.S2C_Game_CreateTable_Ok})

		table.EnterTable(user, gameType, tableNo)
	}, nil)
}

//玩家加入桌子
func (user *User) joinTable(tableNo int) {
	user.Go(func() {
		log.Debug("joinTable tableNo:%v", tableNo)
		table := gameTableMgr.GetTable(tableNo)

		if table == nil {
			log.Debug("joinTable error")
			user.WriteMsg(&msg.S2C_Game_JoinTable_Ret{Ret: msg.S2C_Game_JoinTable_Error})
			return
		}

		log.Debug("joinTable sucess")
		user.WriteMsg(&msg.S2C_Game_JoinTable_Ret{Ret: msg.S2C_Game_JoinTable_Ok})

		table.EnterTable(user, 1, tableNo)
	}, nil)
}

//离开桌子
func (user *User) leaveTable(tableNo int) {
	user.Go(func() {
		no := 0
		gameInfo := UsersGameInfo.Get(user.userInfo.AccID).(*GameInfo)
		if gameInfo != nil {
			no = gameInfo.TableNo
			log.Debug("leaveTable tableNo:%v", no)
		} else {
			log.Debug("leaveTable error tableNo:%v", no)
			return
		}
		table := gameTableMgr.GetTable(no)

		if table == nil {
			log.Debug("leaveTable error 不在游戏中")
			return
		}

		table.LeaveTable(user)
	}, nil)
}

//玩家准备
func (user *User) ready(tableNO int) {
	user.Go(func() {
		table := gameTableMgr.GetTable(tableNO)
		if table == nil {
			log.Debug("table == nil error")
			return
		}

		table.ReadyGame(user.userInfo.AccID, tableNO)
	}, nil)
}

//玩家出牌
func (user *User) outCard(card int) {
	user.Go(func() {
		gameInfo := UsersGameInfo.Get(user.userInfo.AccID).(*GameInfo)
		table := gameTableMgr.GetTable(gameInfo.TableNo)
		if table == nil {
			log.Debug("table == nil error")
			return
		}

		table.DoOutCard(gameInfo.ChairID, card)
	}, nil)
}

//玩家操作
func (user *User) operatorCard(action int, card int) {
	user.Go(func() {
		gameInfo := UsersGameInfo.Get(user.userInfo.AccID).(*GameInfo)
		table := gameTableMgr.GetTable(gameInfo.TableNo)
		if table == nil {
			log.Debug("table == nil error")
			return
		}

		table.DoOperator(gameInfo.ChairID, action, card)
	}, nil)
}

