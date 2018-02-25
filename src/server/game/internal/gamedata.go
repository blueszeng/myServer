package internal

import (
	"math/rand"
	"time"
	"github.com/name5566/leaf/log"
)

const (
	MaxRepertory 		= 64								//最大库存
	MaxIndex 			= 42								//最大索引
	MaxCount 			= 14								//最大数量
	MaxWeaveCount		= 4									//最大组合

	MaskColor			= 0xF0								//颜色掩码
	MaskValue			= 0x0F								//数值掩码

	PlayerCount 		= 2									//游戏人数

	EndLeftCount		= 0									//结束数量
)

const (
	Wik_Null			= 0x00								//null
	Wik_Left			= 0x01								//左吃
	Wik_Center			= 0x02								//中吃
	Wik_Right			= 0x04								//右吃
	Wik_Peng			= 0x08								//碰
	Wik_Gang			= 0x10								//杠
	Wik_Hu				= 0x12								//胡
)

var CardDataArray = [MaxRepertory]int {
	1,1,1,1,2,2,2,2,3,3,3,3,4,4,4,4,5,5,5,5,6,6,6,6,7,7,7,7,8,8,8,8,9,9,9,9,
	49,49,49,49,50,50,50,50,51,51,51,51,52,52,52,52,53,53,53,53,54,54,54,54,55,55,55,55,
}

type GameData struct {
	repertoryCard 		[MaxRepertory]int					//牌数据

	sice1				int									//色子
	sice2				int									//色子
	bankerUser			int									//庄家

	cardIndex			[PlayerCount][MaxIndex]int			//用户数据

	outCardUser			int									//出牌用户
	outCardData			int									//出牌数据
	outCardCount		int									//出牌数量
	gangCount			int									//杠牌数量
	discardCount		[PlayerCount]int					//丢弃数量
	discardCard			[PlayerCount][60]int				//丢弃记录

	sendCardData		int									//发送数据
	leftCardCount		int									//剩余数量

	resumeUser			int									//还原用户
	currentUser			int									//当前用户
	provideUser			int									//供应用户
	provideCard			int									//供应数据

	sendStatus			bool								//发牌状态
	gangStatus			bool								//抢杠状态
	forbidHu			[PlayerCount]bool					//禁止胡牌
	forbidCPG			[PlayerCount]bool					//禁止吃碰

	response			[PlayerCount]bool					//响应标志
	userAction			[PlayerCount]int					//用户动作
	operateCard			[PlayerCount]int					//操作数据
	performAction		[PlayerCount]int					//执行动作

	weaveItemCount		[PlayerCount]int					//组合数量
	weaveItems			[PlayerCount]*WeaveItem				//组合扑克

	huCard				int									//胡牌数据
	huUser				int									//胡牌用户
}

type WeaveItem struct {
	weaveKind			int									//组合类型
	centerCard			int									//中心数据
	provideUser			int									//提供用户
	publicCard			bool								//公开标志

}

func NewWeaveItem(kind int, card int, user int, p bool) *WeaveItem{
	item := new(WeaveItem)
	item.Init(kind, card, user, p)

	return item
}

func (item *WeaveItem) Init(kind int, card int, user int, p bool) {
	item.weaveKind = kind
	item.centerCard = card
	item.provideUser = user
	item.publicCard = p
}

func NewGameData() *GameData {
	gamedata := new(GameData)
	gamedata.Clean()

	return gamedata
}

//清空
func (gamedata *GameData) Clean() {
	gamedata.sice1 = 0
	gamedata.sice2 = 0
	gamedata.bankerUser = -1
	gamedata.outCardUser = -1
	gamedata.outCardData = 0
	gamedata.outCardCount = 0
	gamedata.gangCount = 0
	gamedata.sendCardData = 0
	gamedata.leftCardCount = MaxRepertory
	gamedata.resumeUser = -1
	gamedata.currentUser = -1
	gamedata.provideUser = -1
	gamedata.provideCard = 0
	gamedata.sendStatus = false
	gamedata.gangStatus = false
	gamedata.sice1 = 0
	gamedata.sice1 = 0
	gamedata.huCard = 0
	gamedata.huUser = -1

	for i:=0; i<PlayerCount; i++ {
		gamedata.discardCount[i] = 0
		gamedata.forbidHu[i] = false
		gamedata.forbidCPG[i] = false
		gamedata.response[i] = false
		gamedata.userAction[i] = 0
		gamedata.operateCard[i] = 0
		gamedata.performAction[i] = 0
		gamedata.weaveItemCount[i] = 0

		for j:=0; j<60; j++ {
			gamedata.discardCard[i][j] = 0

			if j<MaxIndex {
				gamedata.cardIndex[i][j] = 0
			}
		}
	}

	for i:=0;i<MaxRepertory; i++ {
		gamedata.repertoryCard[i] = 0
	}
}

//游戏开始，初始化数据
func (gamedata *GameData) Init() {
	log.Debug("")
	gamedata.randCard()
	gamedata.chooseBanker()
	gamedata.DispatchStartCards()
}

//混乱数据
func (gamedata *GameData) randCard() {
	rand.Seed(time.Now().UnixNano())

	tempArray := CardDataArray
	randCount := 0
	position := 0

	for {
		position=rand.Intn(MaxRepertory - randCount)
		gamedata.repertoryCard[randCount] = tempArray[position]
		randCount++
		tempArray[position] = tempArray[MaxRepertory - randCount]

		if randCount >= MaxRepertory {
			break
		}
	}

	//log.Debug("%v",gamedata.repertoryCard)
}

//生成色子数值，选择庄家
func (gamedata *GameData) chooseBanker() {
	rand.Seed(time.Now().UnixNano())

	gamedata.sice1 = rand.Intn(6) + 1
	gamedata.sice2 = rand.Intn(6) + 1

	gamedata.bankerUser = rand.Intn(2)
	gamedata.currentUser = gamedata.bankerUser
}

//生成起始数据
func (gamedata *GameData) DispatchStartCards() {
	for i:=0; i<PlayerCount; i++ {
		gamedata.leftCardCount = gamedata.leftCardCount - (MaxCount - 1)

		tempArray := gamedata.repertoryCard[gamedata.leftCardCount : gamedata.leftCardCount+(MaxCount-1)]
		for j:=0; j<(MaxCount-1); j++ {
			index := gamedata.switchToCardIndex(tempArray[j])
			gamedata.cardIndex[i][index]++
		}
	}

	gamedata.leftCardCount--
	index := gamedata.switchToCardIndex(gamedata.repertoryCard[gamedata.leftCardCount])
	gamedata.cardIndex[gamedata.bankerUser][index]++

	gamedata.sendCardData = gamedata.repertoryCard[gamedata.leftCardCount]
	//log.Debug("%v %v", gamedata.repertoryCard[gamedata.leftCardCount], gamedata.leftCardCount)
}

//获取发送的数据
func (gamedata *GameData) GetNextCard(chairID int) int{

	//如果到剩余牌数为EndLeftCount，则结束
	if gamedata.leftCardCount <= EndLeftCount {
		gamedata.huUser = -1
		gamedata.provideUser = -1

		return -1
	}

	gamedata.leftCardCount--
	gamedata.sendCardData = gamedata.repertoryCard[gamedata.leftCardCount]
	gamedata.cardIndex[chairID][gamedata.switchToCardIndex(gamedata.sendCardData)]++
	gamedata.provideUser = gamedata.currentUser
	gamedata.provideCard = gamedata.sendCardData

	return gamedata.sendCardData
}

func (gamedata *GameData) AddGangCount() {
	gamedata.gangCount++
}

//检测玩家动作
func (gamedata *GameData) CheckAction(user int) bool {

	return false
}

//设置下一个玩家
func (gamedata *GameData) SetNextUser(nowUser int) {
	gamedata.currentUser = (nowUser + PlayerCount - 1) % PlayerCount
}

//获取玩家14张牌数据
func (gamedata *GameData) UserCards(userPos int) [MaxCount]int {
	index := 0
	var tmp [MaxCount]int
	for i:=0; i<MaxIndex; i++ {
		if gamedata.cardIndex[userPos][i] == 0 {
			continue
		}

		for j:=0; j<gamedata.cardIndex[userPos][i]; j++ {
			tmp[index] = gamedata.switchToCardData(i)
			index++
		}
	}

	return tmp
}

//游戏快照，游戏当前数据
func (gamedata *GameData) NowSceneData() {

}

//数值转换为索引
func (gamedata *GameData) switchToCardIndex(card int) int{
	return ((card&MaskColor)>>4)*9+(card&MaskValue)-1;
}

//索引转换为数值
func (gamedata *GameData) switchToCardData(index int) int{
	if index < (MaxIndex - 8) { //8是花牌数量
		return ((index/9)<<4)|(index%9+1)
	} else {
		return ((3<<4)|(index - 27 + 1))
	}
}

//校验数据
func (gamedata *GameData) IsValidCard(card int) bool {
	value := card&MaskValue
	color := card&MaskColor

	return (((value>=1)&&(value<=9)&&(color<=2))||((value>=1)&&(value<=15)&&(color==3)))
}

//动作等级
func (gamedata *GameData) ActionRank(action int) int {
	if action&Wik_Hu != 0 {
		return 4
	}

	if action&Wik_Gang != 0 {
		return 3
	}

	if action&Wik_Peng != 0 {
		return 2
	}

	if action&(Wik_Left|Wik_Center|Wik_Right) != 0 {
		return 1
	}

	return 0
}

//检测吃
func (gamedata *GameData) EstimateEat(chairID int, card int) int {
	if card >= 0x31 {
		return Wik_Null
	}

	excursion := [3]int{0, 1, 2}
	itemkind := [3]int{Wik_Left, Wik_Center, Wik_Right}

	eatKind := 0
	firstIndex := 0
	currentIndex := gamedata.switchToCardIndex(card)
	for i:=0; i<3; i++ {
		valueIndex := currentIndex%9
		if (valueIndex>=excursion[i]) && (valueIndex-excursion[i]<=6) {
			firstIndex = currentIndex-excursion[i]

			if (currentIndex!=firstIndex) && (gamedata.cardIndex[chairID][firstIndex]==0) {
				continue
			}
			if (currentIndex!=firstIndex+1) && (gamedata.cardIndex[chairID][firstIndex+1]==0) {
				continue
			}
			if (currentIndex!=firstIndex+2) && (gamedata.cardIndex[chairID][firstIndex+2]==0) {
				continue
			}

			eatKind|=itemkind[i]
		}
	}

	return eatKind
}

//检测碰
func (gamedata *GameData) EstimatePeng(chairID int, card int) int {
	if gamedata.cardIndex[chairID][gamedata.switchToCardIndex(card)] >= 2 {
		return Wik_Peng
	}

	return Wik_Null
}

//检测杠
func (gamedata *GameData) EstimateGang(chairID int, card int) int {
	if card != 0 && gamedata.cardIndex[chairID][gamedata.switchToCardIndex(card)] == 3 {
		return Wik_Gang
	} else if card == 0 {
		gangCard := [4]int{0, 0, 0, 0}
		index := 0
		hasGang := false
		for i:=0; i<MaxIndex; i++ {
			if gamedata.cardIndex[chairID][i] == 4 {
				gangCard[index] = gamedata.switchToCardData(i)
				index++
				hasGang = true
			}
		}

		for i:=0; i<gamedata.weaveItemCount[chairID]; i++ {
			if gamedata.weaveItems[i].weaveKind == Wik_Peng && gamedata.cardIndex[chairID][gamedata.switchToCardIndex(gamedata.weaveItems[i].centerCard)] == 1 {
				gangCard[index] = gamedata.weaveItems[i].centerCard
				index++
				hasGang = true
			}
		}

		if hasGang == true {
			return Wik_Gang
		} else {
			return Wik_Null
		}
	}

	return Wik_Null
}

//检测胡
func (gamedata *GameData) AnalyseHu(chairID int, card int) int {

	cardIndexTemp := gamedata.cardIndex[chairID]

	if card != 0 {
		cardIndexTemp[gamedata.switchToCardIndex(card)]++
	}



	return Wik_Null
}