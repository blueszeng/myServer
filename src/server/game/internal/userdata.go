package internal

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/gate"
	"server/msg"
	"github.com/name5566/leaf/util"
)

var (
	UserAccIDMgr 		= new(util.Map)						//所有在线玩家   map [accID]*User
	UserIDMgr			= new(util.Map)						//所有在线玩家   map [userID]*User
	UsersGameInfo 		= new(util.Map)						//玩家游戏信息   map[accID]*User
)

//userinfo 用户基本信息
type UserInfo struct {
	UserID 				int "_id"
	AccID  				string
	PassWD				string
	Sex					int
	Money				int
	Level				int
}

func (userInfo *UserInfo) initValue(accID string, passWD string, sex int) error {
	userID, err := mongoDBNextSeq("users")
	if err != nil {
		return fmt.Errorf("get next users id error: %v", err)
	}

	userInfo.UserID = userID
	userInfo.AccID = accID
	userInfo.PassWD = passWD
	userInfo.Sex = sex
	userInfo.Money = 1000
	userInfo.Level = 1

	return nil
}

func checkAccIdExist(a gate.Agent, accID string, passWD string , sex int) {
	userInfo := new(UserInfo)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB("game").C("users").Find(bson.M{"accid": accID}).One(userInfo)
		if err != nil {
			if err != mgo.ErrNotFound {
				log.Error("find acc %v data error: %v", accID, err)
				userInfo = nil
				return
			}
		}
	}, func() {
		if userInfo.AccID != "" {
			m := &msg.S2C_Register_Ret{Ret: msg.S2C_Register_ExistError}
			a.WriteMsg(m)
			log.Debug("Acc %v 已经存在", userInfo.AccID)
		} else {
			addAccId2DB(a, accID, passWD, sex)
		}

		userInfo = nil
	})
}

func addAccId2DB(a gate.Agent, accID string, passWD string , sex int) {
	userInfo := new(UserInfo)
	userInfo.initValue(accID, passWD, sex)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)

		err := db.DB("game").C("users").Insert(userInfo)
		if err != nil{
			log.Fatal("err register - %v, err ",err )
			return
		}

	}, func() {
		log.Debug("Acc %v 注册成功", userInfo.AccID)
		m := &msg.S2C_Register_Ret{Ret: msg.S2C_Register_Ok}
		a.WriteMsg(m)
	})
}
