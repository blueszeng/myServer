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
				log.Error("查询错误")
			} else {
				log.Error("账号不存在")
				return
			}
		}
	}, func() {
		// login repeated
		//oldUser := accIDUsers[accID]
		oldUser := AllUsers.Get(accID)
		if oldUser != nil {
			old := oldUser.(*User)
			m := &msg.S2C_Close{Err: msg.S2C_Close_LoginRepeated}
			a.WriteMsg(m)
			old.WriteMsg(m)
			a.Close()
			old.Close()
			log.Debug("acc %v login repeated", accID)
			return
		}

		// login
		newUser := new(User)
		newUser.Agent = a
		newUser.LinearContext = skeleton.NewLinearContext()
		newUser.state = userLogin
		newUser.userInfo = userInfo
		a.UserData().(*AgentInfo).accID = accID
		//accIDUsers[accID] = newUser
		if AllUsers.Get(accID) != nil {
			AllUsers.Set(accID, newUser)
		} else {
			AllUsers.Del(accID)
			AllUsers.Set(accID, newUser)
		}

		newUser.login(accID, userInfo)
	})
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	accID := a.UserData().(*AgentInfo).accID
	a.SetUserData(nil)

	//user := accIDUsers[accID]
	acc := AllUsers.Get(accID)
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
		user.logout(accID)
	}
}
