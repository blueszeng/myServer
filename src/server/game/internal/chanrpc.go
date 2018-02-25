package internal

import (
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"server/msg"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
)

type AgentInfo struct {
	accID  string
	userID int
}

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)

	skeleton.RegisterChanRPC("UserLogin", rpcUserLogin)
	skeleton.RegisterChanRPC("UserRegister", rpcUserRegister)
}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	a.SetUserData(new(AgentInfo))
}

//处理注册
func rpcUserRegister(args []interface{}) {
	a := args[0].(gate.Agent)
	accID := args[1].(string)
	passWD := args[2].(string)
	sex := args[3].(int)

	if a.UserData() == nil {
		log.Debug("a.UserData() == nil")
		return
	}

	checkAccIdExist(a, accID, passWD, sex)
}

//处理登录
func rpcUserLogin(args []interface{}) {
	a := args[0].(gate.Agent)
	accID := args[1].(string)

	// network closed
	if a.UserData() == nil {
		m := &msg.S2C_Close{Err: msg.S2C_Close_InnerError}
		a.WriteMsg(m)
		a.Close()
		return
	}

	//检测账号是否存在
	userInfo := new(UserInfo)
	skeleton.Go(func() {
		db := mongoDB.Ref()
		defer mongoDB.UnRef(db)
		err := db.DB("game").C("users").Find(bson.M{"accid": accID}).One(userInfo)
		if err != nil {
			if err != mgo.ErrNotFound {
				log.Error("find error :%v", err)
			} else {
				log.Error("account not found :%v", err)
			}
			return
		}
	}, func() {
		//检测是否是断线重连
		oldUser := UserAccIDMgr.Get(accID)
		if oldUser != nil {
			ou := oldUser.(*User)
			m := &msg.S2C_Close{Err: msg.S2C_Close_LoginRepeated}
			ou.WriteMsg(m)
			ou.Close()
			log.Debug("acc %v login repeated, close old socket", accID)

			//a.WriteMsg(m)
			//a.Close()
			return
		}

		//login
		newUser := new(User)
		newUser.Agent = a
		newUser.LinearContext = skeleton.NewLinearContext()
		newUser.state = userLogin
		newUser.userInfo = userInfo
		a.UserData().(*AgentInfo).accID = accID
		a.UserData().(*AgentInfo).userID = userInfo.UserID

		if UserAccIDMgr.Get(accID) == nil {
			UserAccIDMgr.Set(accID, newUser)
		} else {
			UserAccIDMgr.Del(accID)
			UserAccIDMgr.Set(accID, newUser)
		}

		if UserIDMgr.Get(userInfo.UserID) == nil {
			UserIDMgr.Set(userInfo.UserID, newUser)
		} else {
			UserIDMgr.Del(userInfo.UserID)
			UserIDMgr.Set(userInfo.UserID, newUser)
		}

		newUser.login()
	})
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	accID := a.UserData().(*AgentInfo).accID
	a.SetUserData(nil)

	//user := accIDUsers[accID]
	acc := UserAccIDMgr.Get(accID)
	if acc == nil {
		return
	}

	log.Debug("acc %v logout", accID)

	user := acc.(*User)

	// logout
	if user.state == userLogin {
		user.state = userLogout
	} else {
		user.state = userLogout
		user.logout()
	}
}
