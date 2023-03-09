package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"settlement_go/internal/consts"
	"settlement_go/internal/helper"
	"settlement_go/internal/logic"
	"settlement_go/internal/model"
	"settlement_go/internal/util"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"ph-gitlab.vipsroom.net/bingo_backend/common_core/common"
	ch "ph-gitlab.vipsroom.net/bingo_backend/common_core/helper"
	cmodel "ph-gitlab.vipsroom.net/bingo_backend/common_core/tables"
)

const SettleNum = 10000

type SettleMega struct {
	PeriodNum        string //期号
	MerchantServices map[string]*SettleMegaMerchant
	MerchantConfig   map[string]*model.MerchantConfig
	PrizeNumber      []int32                      // 开奖号码
	PrizeNumberStr   string                       // 开奖号码
	InitStatus       map[string]consts.DataStatus // key=periodNum  1 initData 2 loadData
	OddMap           map[string]string            //extra patterns 的赔率 key:图案 value:赔率
	JackpotCount     int
	BingoCount       int
	PrizePool        model.PrizePot
	EsPrizeClient    *helper.EsPrizeClient
	Flag             bool
	RecordList       []*model.JackpotRecordModel
	HistoryList      []*model.RankHistoryModel
	Mu               sync.Mutex
}
type SettleMegaMerchant struct {
	Id                  string
	Config              *model.MerchantConfig
	mu                  sync.RWMutex
	ExtraPatternsCards  []*CardN                        //所有extra patterns卡片的集合
	BingoCards          []*CardN                        //所有 bingo and jackpot卡片的集合
	OrderMap            map[string]*cmodel.TbBingoOrder //所有订单的集合 key=wallet-id value=order
	BingoOrderMap       map[string]*cmodel.TbBingoOrder //所有订单的集合 key=wallet-id value=order
	ExtraOrderMap       map[string]*cmodel.TbBingoOrder //所有订单的集合 key=wallet-id value=order
	ExtraPatternsOrders []*cmodel.TbBingoOrder          // 这一期中读到内存中一部分数据
	BingoOrders         []*cmodel.TbBingoOrder          // 这一期中读到内存中一部分数据
	CardsMap            map[string]*CardN
	JackpotPrizeCards   []*CardN
	BingoPrizeCards     []*CardN
	PrizeMap            sync.Map
	//Orders                 []*cmodel.TbBingoOrder
	//Transfer               []*model.TransferRequestBody
	ExtraPatternsCardCount sync.Map // 每个用户对应所有的卡片总数 key = wallet id value = count
	BingoCardCount         sync.Map // 每个用户对应所有的卡片总数 key = wallet id value = count
	CardsName              sync.Map // 每个用户对应所有的卡片总数 key = wallet id value = name
}
type PrizeModel struct {
	Extra      float64
	Jackpot    float64
	Bingo      float64
	OneTG      float64
	TwoTG      float64
	PeriodNum  string
	WalletId   string
	MerchantId string
	IsRush     bool
}

func NewSettleMega() *SettleMega {
	InitPatterns()
	return &SettleMega{
		InitStatus:       make(map[string]consts.DataStatus, 0),
		MerchantServices: make(map[string]*SettleMegaMerchant, 0),
		MerchantConfig:   make(map[string]*model.MerchantConfig, 0),
		OddMap:           map[string]string{},
		PrizePool:        model.PrizePot{},
		JackpotCount:     0,
		BingoCount:       0,
		Flag:             false,
		HistoryList:      make([]*model.RankHistoryModel, 0),
		RecordList:       make([]*model.JackpotRecordModel, 0),
		EsPrizeClient:    helper.NewEsPrizeClient(logic.EsClient),
		Mu:               sync.Mutex{},
	}
}

func (m *SettleMega) GameEvent(event *model.GameRoundEvent) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	zap.S().Info("发球数据--->>", event)
	ge := time.Now()
	defer zap.S().Info("GameEvent time=", time.Since(ge))
	status := event.GetStatus()
	if status == "" {
		zap.S().Error("bad data")
		return
	}
	m.PeriodNum = event.GameRoundId

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	now := time.Now()
	switch status {
	case consts.Begin:
		m.Clean()
		m.Flag = false
		//zap.S().Info("Begin")
		m.ReadingPeriod(ctx)
		err := m.InitData(ctx, event.GameRoundId)
		m.ReadingOdd(ctx, []string{"BGM:100", "BGM:101"})
		if err != nil {
			zap.S().Error("Begin InitData fail err=", err)
			return
		}
		break
	case consts.PrizeOnGoing:
		m.PrizeNumberStr = event.PrizeNumber
		if len(m.MerchantServices) == 0 {
			t3 := time.Now()
			m.ReadingPeriod(ctx)
			zap.S().Info("ReadingPeriod time=", time.Since(t3))
			_ = m.InitData(ctx, event.GameRoundId)
			m.ReadingOdd(ctx, []string{"BGM:100", "BGM:101"})
			m.ReadingPrizePool()
			zap.S().Info("P0 time=", time.Since(t3))
		}
		m.PrizeOnGoing(event)
		m.Count()
		break
	case consts.Finish:
		m.Payment()
		m.CleanOrderFromRedis()
		break
	//批量取消逻辑 转到订单服务处理
	case consts.Cancel: //按照期号取消订单
		//zap.S().Info("cancel")
		//m.ReadingPeriod(ctx)
		//_ = m.InitData(ctx, event.GameRoundId)
		//m.Cancel()
		if m.PeriodNum == event.GameRoundId {
			m.Clean()
		} else {
			m.EsPrizeClient.DeleteByCondition(event.GameRoundId)
		}
		break
	}
	if time.Now().UnixMilli()-now.UnixMilli() > 1000 {
		zap.S().Info("settle time spend ", time.Since(now), "发球状态", event)
	}
}

func (m *SettleMega) InitData(ctx context.Context, periodNum string) error {
	res, err := logic.Rdb.HGetAll(ctx, consts.MerchantRedisKey).Result()
	if err != nil {
		zap.S().Info("读取redis 商户失败", err.Error())
		return errors.Wrap(err, " Rdb.HGetAll fail")
	} else {
		for _, v := range res {
			temp := model.MerchantConfig{}
			err := jsonOwn.Unmarshal([]byte(v), &temp)
			if err != nil {
				zap.S().Info("解析数据异常", err.Error())
				return errors.Wrap(err, " Unmarshal fail")
			}
			m.MerchantServices[temp.MerchantID] = &SettleMegaMerchant{}
			m.MerchantConfig[temp.MerchantID] = &temp
		}
	}
	for key, merchant := range m.MerchantServices {
		merchant.InitData(key, m.MerchantConfig[key])
	}
	m.InitStatus[periodNum] = consts.InitData
	return nil
}

func (m *SettleMega) ReadingPeriod(ctx context.Context) error {
	res, err := logic.Rdb.HGet(ctx, consts.MerchantPeriodRedis, "BGM").Result()
	if err != nil {
		zap.S().Info("读取redis 商户失败", err.Error())
		return errors.Wrap(err, " HGet fail")
	} else {
		mp := model.GameRoundEvent{}
		err := jsonOwn.Unmarshal([]byte(res), &mp)
		if err != nil {
			zap.S().Info("解析期号出错", err.Error())
			return errors.Wrap(err, " Unmarshal fail")
		} else {
			m.PeriodNum = mp.GameRoundId
		}
	}
	return nil
}

func (m *SettleMega) ReadOrders(ctx context.Context, periodNum string) {
	for index, config := range m.MerchantConfig {
		k1 := fmt.Sprintf("%s%s:BingoMega:%s", RedisPrefix, periodNum, config.MerchantID)
		res1, err := logic.Rdb.HGetAll(ctx, k1).Result()
		if err != nil {
			zap.S().Info("读取redis 订单失败", err.Error())
		} else if len(res1) == 0 {
			//zap.S().Info("当前商户", index, "的订单数为0. key is --->>>>", k1)
			continue
		} else {
			zap.S().Info("读取redis 订单总数", len(res1), "商户号", index)
			wg := sync.WaitGroup{}
			for key, v := range res1 {
				ss := m.MerchantServices[index]
				wg.Add(1)
				go func(k, val string, memoryService *SettleMegaMerchant) {
					temp := cmodel.TbBingoOrder{}
					err := jsonOwn.Unmarshal([]byte(val), &temp)
					if err != nil {
						zap.S().Info("解析数据异常", err.Error(), val)
					} else {
						memoryService.AddOrderMap(&temp)
						for _, card := range temp.BetContent {
							memoryService.AddCards(card, &temp)
						}
					}
					wg.Done()
				}(key, v, ss)
			}
			wg.Wait()
			for id, merchant := range m.MerchantServices {
				if len(merchant.BingoOrders)+len(merchant.ExtraPatternsOrders) > 0 {
					zap.S().Info("商户", id, "  卡片订单数 ", len(merchant.BingoOrders)+len(merchant.ExtraPatternsOrders))
				}
			}
		}
	}
	m.InitStatus[periodNum] = consts.LoadData
}

// ReadingOdd 读取赔率的数据
func (m *SettleMega) ReadingOdd(ctx context.Context, gameType []string) (err error) {
	if len(gameType) == 0 {
		gameType = []string{"BGM:100"}
	}
	res, err := logic.Rdb.HGetAll(ctx, consts.MerchantOddRedis).Result()
	if err != nil {
		zap.S().Info("读取redis 赔率失败", err.Error())
	} else {
		for val, v := range res {
			if util.ArraysHasItem(gameType, val) {
				temp := model.GameCategoryListResp{}
				err := json.Unmarshal([]byte(v), &temp)
				if err != nil {
					zap.S().Info("解析数据异常", err.Error())
				}
				for _, category := range temp.List {
					m.OddMap[category.BingoPatten] = category.Bonus
					if len(category.BonusPotMin) > 1 {
						m.OddMap[category.BingoPatten] = category.BonusPotMin
					}
				}
			}
		}
	}
	return
}
func (m *SettleMega) ReadingPrizePool() {
	res, err := logic.Rdb.HGet(context.Background(), consts.MerchantPrizePoolRedis, "BGM").Result()
	if err != nil {
		zap.S().Info("读取redis 奖池失败", err.Error())
	} else {
		json.Unmarshal([]byte(res), &m.PrizePool)
		if err != nil {
			zap.S().Info("读取奖池信息出错", err.Error())
		} else {
			zap.S().Info("读取奖池信息", m.PrizePool.String())
		}
	}
}
func (m *SettleMega) UpdatePrizePool() {
	res, _ := json.Marshal(m.PrizePool)
	_, err := logic.Rdb.HSet(context.Background(), consts.MerchantPrizePoolRedis, "BGM", res).Result()
	if err != nil {
		zap.S().Info("更新奖池信息失败", err.Error())
	}
}
func (m *SettleMega) PrizeOnGoing(event *model.GameRoundEvent) {

	numbers, err := event.GetPrizeNumber()
	if err != nil {
		zap.S().Error("GetPrizeNumber fail err=", err)
		return
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	status, ok := m.InitStatus[event.GameRoundId]
	if !ok {
		err := m.InitData(ctx, event.GameRoundId)
		if err != nil {
			zap.S().Error("PrizeOnGoing InitData fail err=", err)
			return
		}
		m.ReadOrders(ctx, event.GameRoundId)
	} else {
		if status == consts.UnKnownStatus {
			zap.S().Error("UnKnownStatus")
			return
		}
		if status == consts.InitData {
			m.ReadOrders(ctx, event.GameRoundId)
			m.ReadingPrizePool()
		}
		verifyNumbers, b := m.VerifyNumbers(numbers)
		if !b {
			zap.S().Info("verifyNumbers=", verifyNumbers)
			for _, number := range verifyNumbers {
				bingoTime := len(m.PrizeNumber)+1 > 44
				extraPatternsCards := make([]*CardN, 0)
				bingoCards := make([]*CardN, 0)
				var (
					e  int
					bb int
				)

				for _, merchant := range m.MerchantServices {
					e += len(merchant.ExtraPatternsCards)
					bb += len(merchant.BingoCards)
					extraPatternsCards = append(extraPatternsCards, merchant.ExtraPatternsCards...)
					bingoCards = append(bingoCards, merchant.BingoCards...)
				}
				if len(extraPatternsCards) > 0 && len(m.PrizeNumber) < 45 {
					if len(extraPatternsCards) != e {
						zap.S().Info("extraPatternsCards not eq e=", e, " len=", len(extraPatternsCards))
					}
					m.ExtraPatternsSettle(extraPatternsCards, int(number))
				}
				if len(bingoCards) > 0 {
					if len(bingoCards) != bb {
						zap.S().Info("bingoCards not eq bb=", bb, " len=", len(bingoCards))
					}
					m.BingoJackPotSettle(bingoCards, int(number), bingoTime)
				}
				m.PrizeNumber = append(m.PrizeNumber, number)
				if len(m.PrizeNumber) == 44 {
					m.M(true)
					m.PaymentJackpot()
					m.PaymentExtraPatterns()
				}
				if len(m.PrizeNumber) == 49 {
					m.M(false)
					m.PaymentBingo()
				}
			}
		} else {
			number, err := event.GetLastPrizeNumber()
			if err != nil {
				zap.S().Error("GetLastPrizeNumber fail err=", err)
				return
			}
			bingoTime := len(numbers) > 44
			extraPatternsCards := make([]*CardN, 0)
			bingoCards := make([]*CardN, 0)
			for _, merchant := range m.MerchantServices {
				extraPatternsCards = append(extraPatternsCards, merchant.ExtraPatternsCards...)
				bingoCards = append(bingoCards, merchant.BingoCards...)
			}
			if len(extraPatternsCards) > 0 && len(m.PrizeNumber) < 45 {
				m.ExtraPatternsSettle(extraPatternsCards, int(number))
			}
			if len(bingoCards) > 0 {
				m.BingoJackPotSettle(bingoCards, int(number), bingoTime)
			}
			m.PrizeNumber = append(m.PrizeNumber, number)
			if len(m.PrizeNumber) == 44 {
				m.M(true)
				m.PaymentJackpot()
				m.PaymentExtraPatterns()
			}
			if len(m.PrizeNumber) == 49 {
				m.M(false)
				m.PaymentBingo()
			}
		}
	}
}

//	func (m *SettleMega) Cancel() {
//		if m == nil || len(m.PeriodNum) == 0 {
//			return
//		}
//		zap.S().Info("<<<<<<批量取消订单start 期号>>>>>>", m.PeriodNum)
//		err := OrderStub.RpcCancelOrder(m.PeriodNum, consts.BGM)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//		zap.S().Info("<<<<<<批量取消订单finish 期号>>>>>>", m.PeriodNum)
//	}
func (m *SettleMega) M(isJackpot bool) {
	for _, merchant := range m.MerchantServices {
		for _, card := range merchant.BingoCards {
			if len(card.Index) >= 24 {
				if card.Jackpot && isJackpot {
					m.JackpotCount = m.JackpotCount + 1
					merchant.JackpotPrizeCards = append(merchant.JackpotPrizeCards, card)
				} else if card.Bingo && !isJackpot {
					m.BingoCount = m.BingoCount + 1
					merchant.BingoPrizeCards = append(merchant.BingoPrizeCards, card)
				}
			}
		}
	}
	zap.S().Info("当前期中奖jackpot的总数 --->>>> ", m.JackpotCount, "当前期中奖bingo的总数 --->>>> ", m.BingoCount)
}

func (m *SettleMega) Payment() {

	var (
		wg = sync.WaitGroup{}
	)

	for _, merchant := range m.MerchantServices {
		wg.Add(1)
		go func(ms *SettleMegaMerchant) {
			now := time.Now()
			defer wg.Done()
			ms.mu.Lock()
			defer ms.mu.Unlock()
			ordersLen := len(ms.BingoOrders) + len(ms.ExtraPatternsOrders)
			mss := ms.Config
			if mss == nil {
				return
			}
			orders := make([]*cmodel.TbBingoOrder, 0, len(ms.BingoOrders))
			ordersExtra := make([]*cmodel.TbBingoOrder, 0, len(ms.ExtraPatternsOrders))
			err1 := copier.CopyWithOption(&orders, &ms.BingoOrders, copier.Option{DeepCopy: true})
			err2 := copier.CopyWithOption(&ordersExtra, &ms.ExtraPatternsOrders, copier.Option{DeepCopy: true})
			if err1 != nil || err2 != nil {
				orders = make([]*cmodel.TbBingoOrder, 0, len(ms.BingoOrders)+len(ms.ExtraPatternsOrders))
				orders = append(orders, ms.BingoOrders...)
				orders = append(orders, ms.ExtraPatternsOrders...)
			}
			orders = append(orders, ordersExtra...)
			zap.S().Info("结算的订单总数--->>", ordersLen, " bing=", len(ms.BingoOrders), " extra=", len(ms.ExtraPatternsOrders), "商户表名", mss.SettledTableName, " 期号=", m.PeriodNum)
			go BatchSaveOrder(orders, mss.SettledTableName)
			//if len(ms.Transfer) > 0 {
			//	zap.S().Info("中奖的卡片总数--->>", len(ms.Transfer))
			//	go Payout(ms.Transfer)
			//}
			if time.Now().UnixMilli()-now.UnixMilli() > 1000 {
				zap.S().Info("结算派彩的时长商户 ", time.Since(now))
			}
		}(merchant)
	}
	wg.Wait()
	if (m.BingoCount > 0 || m.Flag) && len(m.RecordList) > 0 {
		m.CheckBingoPrize(m.RecordList)
	}
	if len(m.HistoryList) > 0 {
		zap.S().Info("插入历史中奖记录长度", len(m.HistoryList))
		m.EsPrizeClient.Save(m.HistoryList)
	}
	if m.JackpotCount > 0 || m.BingoCount > 0 {
		m.UpdatePrizePool()
	}
}

func (m *SettleMega) PaymentBingo() {
	bamount := m.GetJackpotAmount(false)
	now := time.Now()
	wg := sync.WaitGroup{}
	for _, merchant := range m.MerchantServices {
		wg.Add(1)
		go func(ms *SettleMegaMerchant, w *sync.WaitGroup) {
			defer w.Done()
			ms.mu.Lock()
			for _, order := range ms.BingoOrders {
				total := float64(0)
				bingo := float64(0)
				one := float64(0)
				two := float64(0)
				for _, card := range order.BetContent {
					val, ok := ms.CardsMap[fmt.Sprintf("%s-%s", card.Id, order.ID)]
					if ok {
						amount := float64(0)
						if m.BingoCount > 0 {
							if val.Bingo {
								amount = amount + bamount
								//_, t := m.ConfigCardRecord(amount, bamount, 2, card, order)
								//ms.Transfer = append(ms.Transfer, t)
								total = total + bamount
								m.RecordList = append(m.RecordList, ConfigBingoPrizePool(2, bamount, m.PeriodNum))
								bingo += bamount
								card.PrizesAmount = FormatFloat(bamount, 2)
								card.PrizesType = 2
								card.PrizesName = "bingo"
							}
						}

						if val.IsOneTG() {
							card.PrizesType = 2
							m.Flag = true
							card.PrizesName = "1TG"
							if v, ok1 := m.OddMap["1TG"]; ok1 {
								if m.BingoCount > 0 {
									card.PrizesAmount = "0"
								} else {
									aa, _ := strconv.ParseInt(v, 10, 64)
									bb, _ := strconv.ParseFloat(m.PrizePool.Bingo, 0)
									aa = aa / 100
									amount = amount + float64(aa)
									m.PrizePool.Bingo = fmt.Sprintf("%.2f", bb-float64(aa))
									//ms.Transfer = append(ms.Transfer, m.ConfigOrderRecording(amount, order))
									card.PrizesAmount = fmt.Sprintf("%d", aa)
									total = total + amount
									m.RecordList = append(m.RecordList, ConfigBingoPrizePool(2, amount, m.PeriodNum))
									one += amount
								}
							}
						}
						if val.IsTwoTG() {
							m.Flag = true
							card.PrizesType = 2
							card.PrizesName = "2TG"
							if v, ok1 := m.OddMap["2TG"]; ok1 {
								if m.BingoCount > 0 {
									card.PrizesAmount = "0"
								} else {
									aa, _ := strconv.ParseInt(v, 10, 64)
									bb, _ := strconv.ParseFloat(m.PrizePool.Bingo, 0)
									aa = aa / 100
									amount = amount + float64(aa)
									m.PrizePool.Bingo = fmt.Sprintf("%.2f", bb-float64(aa))
									//ms.Transfer = append(ms.Transfer, m.ConfigOrderRecording(amount, order))
									card.PrizesAmount = fmt.Sprintf("%d", aa)
									total = total + amount
									m.RecordList = append(m.RecordList, ConfigBingoPrizePool(2, amount, m.PeriodNum))
									two += amount
								}
							}
						}
						card.PrizesContent = m.PrizeNumberStr
					}
				}
				if total > 0 {
					if bingo > 0 || one > 0 || two > 0 {
						m.HistoryList = append(m.HistoryList, ConfigRankHistoryModel(fmt.Sprintf("%.2f", bingo+one+two), 2, m.PeriodNum, order))
					}
					if val, ok := ms.PrizeMap.Load(order.WalletId); ok {
						if bingo > 0 {
							pm := val.(*PrizeModel)
							pm.Bingo = pm.Bingo + bingo
							ms.PrizeMap.Store(order.WalletId, pm)
						} else {
							if one > 0 && two > 0 {
								pm := val.(*PrizeModel)
								pm.TwoTG = pm.TwoTG + two
								pm.OneTG = pm.OneTG + one
								ms.PrizeMap.Store(order.WalletId, pm)
							} else {
								if one > 0 {
									pm := val.(*PrizeModel)
									pm.OneTG = pm.OneTG + one
									ms.PrizeMap.Store(order.WalletId, pm)
								}
								if two > 0 {
									pm := val.(*PrizeModel)
									pm.TwoTG = pm.TwoTG + two
									ms.PrizeMap.Store(order.WalletId, pm)
								}
							}
						}
					} else {
						if bingo > 0 {
							ms.PrizeMap.Store(order.WalletId, &PrizeModel{Bingo: bingo, MerchantId: ms.Id, PeriodNum: m.PeriodNum, IsRush: false})
						} else {
							if one > 0 {
								ms.PrizeMap.Store(order.WalletId, &PrizeModel{OneTG: one, MerchantId: ms.Id, PeriodNum: m.PeriodNum, IsRush: false})
							}
							if two > 0 {
								ms.PrizeMap.Store(order.WalletId, &PrizeModel{TwoTG: two, MerchantId: ms.Id, PeriodNum: m.PeriodNum, IsRush: false})
							}
							if one > 0 && two > 0 {
								ms.PrizeMap.Store(order.WalletId, &PrizeModel{TwoTG: two, OneTG: one, MerchantId: ms.Id, PeriodNum: m.PeriodNum, IsRush: false})
							}
						}
					}
				}
				order.Status = 2
				order.LotteryContent = m.PrizeNumberStr
				am, _ := strconv.ParseFloat(order.AwardAmount, 0)
				total = total + am
				if total > 0 {
					order.WinLose = fmt.Sprintf("+%.2f", total)
				} else if total == 0 {
					order.WinLose = fmt.Sprintf("%.2f", total)
				} else {
					order.WinLose = fmt.Sprintf("-%.2f", total)
				}
				order.AwardAmount = FormatFloat(total, 2)
				order.UpdatedAt = time.Now().UnixMilli()
			}
			tmp := ConvertMapToArray(&ms.PrizeMap)
			if len(tmp) > 0 {
				PublishFinalResult(tmp)
			}
			ms.mu.Unlock()
		}(merchant, &wg)
	}
	wg.Wait()
	zap.S().Info("PaymentBingo spent time=", time.Since(now))
}

func (m *SettleMega) PaymentJackpot() {
	now := time.Now()
	jamount := m.GetJackpotAmount(true)
	wg := sync.WaitGroup{}
	for _, merchant := range m.MerchantServices {
		wg.Add(1)
		go func(ms *SettleMegaMerchant, w *sync.WaitGroup) {
			defer w.Done()
			ms.mu.Lock()
			defer ms.mu.Unlock()
			for _, order := range ms.BingoOrders {
				total := float64(0)
				jackpot := float64(0)
				for _, card := range order.BetContent {
					val, ok := ms.CardsMap[fmt.Sprintf("%s-%s", card.Id, order.ID)]
					if ok {
						amount := float64(0)
						if m.JackpotCount > 0 && val.Jackpot {
							//m.Flag = true
							amount = amount + jamount
							//_, t := m.ConfigCardRecord(amount, jamount, 1, card, order)
							//ms.Transfer = append(ms.Transfer, t)
							jackpot += jamount
							total = total + jamount
							m.RecordList = append(m.RecordList, ConfigBingoPrizePool(1, jamount, m.PeriodNum))
							card.PrizesAmount = FormatFloat(jamount, 2)
							zap.S().Info("bingo mega 中奖 jackpot -->>", len(m.RecordList))
							card.PrizesType = 1
							card.PrizesName = "jackpot"
						}
						//card.PrizesContent = m.PrizeNumberStr
					}
				}
				if total > 0 {
					if jackpot > 0 {
						m.HistoryList = append(m.HistoryList, ConfigRankHistoryModel(fmt.Sprintf("%.2f", jackpot), 1, m.PeriodNum, order))
					}
					if val, ok := ms.PrizeMap.Load(order.WalletId); ok {
						if jackpot > 0 {
							pm := val.(*PrizeModel)
							pm.Jackpot = pm.Jackpot + total
							ms.PrizeMap.Store(order.WalletId, pm)
						}
					} else {
						if jackpot > 0 {
							ms.PrizeMap.Store(order.WalletId, &PrizeModel{Jackpot: total, MerchantId: ms.Id, PeriodNum: m.PeriodNum, IsRush: false})
						}
					}
				}
				order.Status = 2
				if total > 0 {
					order.WinLose = fmt.Sprintf("+%.2f", total)
				} else if total == 0 {
					order.WinLose = fmt.Sprintf("%.2f", total)
				} else {
					order.WinLose = fmt.Sprintf("-%.2f", total)
				}
				order.AwardAmount = FormatFloat(total, 2)
				order.UpdatedAt = time.Now().UnixMilli()
			}
		}(merchant, &wg)
	}
	wg.Wait()
	zap.S().Info("PaymentJackpot spent time=", time.Since(now))
}
func (m *SettleMega) PaymentExtraPatterns() {
	wg := sync.WaitGroup{}
	now := time.Now()
	for _, merchant := range m.MerchantServices {
		wg.Add(1)
		go func(ms *SettleMegaMerchant, w *sync.WaitGroup) {
			defer w.Done()
			ms.mu.Lock()
			defer ms.mu.Unlock()
			for _, order := range ms.ExtraPatternsOrders {
				total := float64(0)
				for _, card := range order.BetContent {
					val, ok := ms.CardsMap[fmt.Sprintf("%s-%s", card.Id, order.ID)]
					if ok {
						//amount := float64(0)
						if len(val.Prize) > 0 {
							// extra patterns
							name, amount1 := ms.CalculatePrizeAmount(val, order, m, false)
							total = total + amount1
							card.PrizesType = 3
							card.PrizesName = name
							//amount = amount1
							card.PrizesAmount = FormatFloat(amount1, 2)
						} else {
							card.PrizesType = 0
							card.PrizesAmount = fmt.Sprintf("%d", 0)
							card.PrizesName = "没有中奖(测试)"
						}
						//if amount > 0 {
						//	ms.Transfer = append(ms.Transfer, m.ConfigOrderRecording(amount, order))
						//}
						card.PrizesContent = m.PrizeNumberStr
					}
				}
				if total > 0 {
					m.HistoryList = append(m.HistoryList, ConfigRankHistoryModel(FormatFloat(total, 2), 0, m.PeriodNum, order))
					if val, ok := ms.PrizeMap.Load(order.WalletId); ok {
						pm := val.(*PrizeModel)
						pm.Extra = pm.Extra + total
						ms.PrizeMap.Store(order.WalletId, pm)
					} else {
						ms.PrizeMap.Store(order.WalletId, &PrizeModel{Extra: total, MerchantId: ms.Id, PeriodNum: m.PeriodNum, IsRush: false})
					}
				}
				order.Status = 2
				order.LotteryContent = m.PrizeNumberStr
				if total > 0 {
					order.WinLose = fmt.Sprintf("+%.2f", total)
				} else if total == 0 {
					order.WinLose = fmt.Sprintf("%.2f", total)
				} else {
					order.WinLose = fmt.Sprintf("-%.2f", total)
				}
				order.AwardAmount = FormatFloat(total, 2)
				order.UpdatedAt = time.Now().UnixMilli()
			}
			tmp := ConvertMapToArray(&ms.PrizeMap)
			if len(tmp) > 0 {
				PublishFinalResult(tmp)
			}
		}(merchant, &wg)
	}
	wg.Wait()
	zap.S().Info("PaymentExtraPatterns spent time=", time.Since(now))
}
func (m *SettleMega) HasItemInTransferRecord(transfer []*model.TransferRequestBody, t *cmodel.TbBingoOrder) (val float64, flag bool) {
	flag = false
	for _, body := range transfer {
		if t.WalletId == body.WalletId {
			flag = true
			val += float64(body.Amount)
		}
	}
	return
}
func (m *SettleMega) CheckBingoPrize(list []*model.JackpotRecordModel) {
	bb, _ := strconv.ParseFloat(m.PrizePool.Bingo, 0)
	aa, _ := strconv.ParseFloat(m.OddMap["49B"], 0)
	if len(list) > 0 {
		zap.S().Info("bingo mega bingo:", m.BingoCount, " 1tg 2tg 中奖个数", len(list)-m.BingoCount)
		zap.S().Info("奖池账变的长度 ", len(list))
		PublishBingoPrizePool(list)
	}
	if bb < aa {
		cc := aa - bb
		mm := ConfigBingoPrizePool(2, aa-bb, m.PeriodNum)
		mm.Remark = "池底注水"
		mm.Amount = decimal.NewFromFloat(cc)
		zap.S().Info("奖池账变上分结构", mm)
		m.PrizePool.Bingo = fmt.Sprintf("%.2f", aa)
		m.UpdatePrizePool()
		mm.InOut = 1
		time.AfterFunc(3*time.Second, func() {
			PublishBingoPrizePool([]*model.JackpotRecordModel{mm})
			zap.S().Info("奖池账变的长度 ", 1)
		})
	}
}

// ConfigCardRecord
func (m *SettleMega) ConfigCardRecord(amount, jamount, t float64, card *cmodel.TbCard, order *cmodel.TbBingoOrder) (am float64, body *model.TransferRequestBody) {
	card.PrizesAmount = fmt.Sprintf("%.2f", jamount)
	card.PrizesType = int(t)
	amount = jamount
	return amount, m.ConfigOrderRecording(amount, order)
}
func (m *SettleMega) ConfigOrderRecording(amount float64, order *cmodel.TbBingoOrder) *model.TransferRequestBody {
	t := m.MerchantConfig[order.MerchantID]
	return &model.TransferRequestBody{
		Amount:       int64(amount * 100),
		Type:         common.Prize,
		IsDeposit:    true,
		TableName:    t.WalletRecord,
		Nickname:     order.Nickname,
		IssueNumber:  order.IssueNumber,
		OrderId:      order.ID,
		UserId:       order.MerchantUserId,
		MerchantId:   order.MerchantID,
		WalletId:     order.WalletId,
		MerchantName: order.MerchantName,
		CreatedAt:    time.Now().UnixMilli(),
	}
}
func (m *SettleMega) VerifyNumbers(numbers []int32) ([]int32, bool) {
	i := len(numbers) - len(m.PrizeNumber)
	if i == 1 {
		return nil, true
	}
	if len(m.PrizeNumber) == 0 {
		return numbers, false
	}
	array := util.DiffArray(numbers, m.PrizeNumber)
	return array, false
}

func (m *SettleMega) ExtraPatternsSettle(extraPatternsCards []*CardN, num int) {
	cards, err := ch.SplitArray[*CardN](extraPatternsCards, SettleNum)
	if err != nil {
		zap.S().Error("ExtraPatternsSettle切分订单数组异常", err.Error())
		return
	}
	group := sync.WaitGroup{}
	for _, card := range cards {
		group.Add(1)
		go func(cs []*CardN, n int, g *sync.WaitGroup) {
			g.Done()
			for _, i4 := range cs {
				if i4.MarkNumber(n) {
					i4.CheckExtraPatternNew()
				}
			}
		}(card, num, &group)
	}
	group.Wait()
}
func (m *SettleMega) BingoJackPotSettle(bingoCards []*CardN, num int, bingo bool) {
	cards, err := ch.SplitArray[*CardN](bingoCards, SettleNum)
	if err != nil {
		zap.S().Error("BingoJackPotSettle切分订单数组异常", err.Error())
		return
	}
	group := sync.WaitGroup{}
	for _, card := range cards {
		group.Add(1)
		go func(cs []*CardN, n int, g *sync.WaitGroup) {
			g.Done()
			for _, i4 := range cs {
				if i4.MarkNumber(n) {
					i4.BingoJackPotSettle(bingo)
				}
			}
		}(card, num, &group)
	}
	group.Wait()
}

func (m *SettleMega) Clean() {
	m.PeriodNum = ""
	for _, merchant := range m.MerchantServices {
		merchant.Clean()
		merchant.CleanRedis()
	}
	m.PrizeNumber = nil
	m.JackpotCount = 0
	m.BingoCount = 0
	m.PrizeNumberStr = ""
	m.HistoryList = make([]*model.RankHistoryModel, 0)
	m.RecordList = make([]*model.JackpotRecordModel, 0)
}

func (m *SettleMega) Count() {
	for _, merchant := range m.MerchantServices {
		go func(ms *SettleMegaMerchant) {
			ms.Count(m)
		}(merchant)
	}
}

func (m *SettleMega) GetJackpotAmount(jackpot bool) (amount float64) {
	if jackpot {
		if m.JackpotCount > 0 {
			total, _ := strconv.ParseFloat(m.PrizePool.JackpotBGM, 0)
			amount = total / float64(m.JackpotCount)
			// TODO
			m.PrizePool.JackpotBGM = m.OddMap["44B"]
			PublishPrizePool(m.PrizePool, m.OddMap, 1, m.PeriodNum)
			zap.S().Info("中奖jackpot 的个数--->>>", m.JackpotCount, "金额--->>", amount, "总金额金额--->>", total, "奖池数据--->>", m.PrizePool.String())
		}
	} else {
		if m.BingoCount > 0 {
			total, _ := strconv.ParseFloat(m.PrizePool.Bingo, 0)
			amount = total / float64(m.BingoCount)
			m.PrizePool.Bingo = m.OddMap["49B"]
			PublishPrizePool(m.PrizePool, m.OddMap, 2, m.PeriodNum)
			zap.S().Info("中奖bingo 的个数--->>>", m.BingoCount, "金额--->>", amount, "总金额金额--->>", total)
		}
	}
	return
}

func (s *SettleMegaMerchant) InitData(id string, config *model.MerchantConfig) {
	s.Id = id
	s.mu = sync.RWMutex{}
	s.ExtraPatternsCards = make([]*CardN, 0)
	s.BingoCards = make([]*CardN, 0)
	s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
	s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
	s.CardsMap = make(map[string]*CardN, 0)
	s.JackpotPrizeCards = make([]*CardN, 0)
	s.BingoPrizeCards = make([]*CardN, 0)
	s.BingoOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.PrizeMap = sync.Map{}
	s.Config = config
	//s.Orders = make([]*cmodel.TbBingoOrder, 0)
	//s.Transfer = make([]*model.TransferRequestBody, 0)
	s.ExtraPatternsCardCount = sync.Map{}
	s.BingoCardCount = sync.Map{}
	s.CardsName = sync.Map{}
}

func (s *SettleMegaMerchant) Clean() {
	s.ExtraPatternsCards = make([]*CardN, 0)
	s.BingoCards = make([]*CardN, 0)
	s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
	s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
	s.CardsMap = make(map[string]*CardN, 0)
	s.JackpotPrizeCards = make([]*CardN, 0)
	s.BingoPrizeCards = make([]*CardN, 0)
	s.BingoOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.PrizeMap = sync.Map{}
	s.Config = &model.MerchantConfig{}
	//s.Orders = make([]*cmodel.TbBingoOrder, 0)
	//s.Transfer = make([]*model.TransferRequestBody, 0)
	s.ExtraPatternsCardCount = sync.Map{}
	s.BingoCardCount = sync.Map{}
	s.CardsName = sync.Map{}
}
func (s *SettleMegaMerchant) AddCards(c *cmodel.TbCard, v *cmodel.TbBingoOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.PlayCode == "101" {
		card := NewCardFromTbCard(c.Numbers, true)
		card.TbCard = *c
		card.Id = c.Id
		s.BingoCards = append(s.BingoCards, card)
		s.CardsMap[fmt.Sprintf("%s-%s", c.Id, v.ID)] = card
	} else if v.PlayCode == "100" {
		card := NewCardFromTbCard(c.Numbers, true)
		card.Id = c.Id
		card.TbCard = *c
		s.ExtraPatternsCards = append(s.ExtraPatternsCards, card)
		s.CardsMap[fmt.Sprintf("%s-%s", c.Id, v.ID)] = card
	} else if v.PlayCode == "103" {
		card := NewCardFromTbCard(c.Numbers, true)
		card.Id = c.Id
		card.TbCard = *c
		s.BingoCards = append(s.BingoCards, card)
		s.ExtraPatternsCards = append(s.ExtraPatternsCards, card)
		s.CardsMap[fmt.Sprintf("%s-%s", c.Id, v.ID)] = card
	}
}

func (s *SettleMegaMerchant) AddOrderMap(c *cmodel.TbBingoOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.PlayCode == "101" {
		s.BingoOrders = append(s.BingoOrders, c)
		s.BingoOrderMap[c.OrderNumber] = c
		if val, ok := s.BingoCardCount.Load(c.WalletId); ok {
			a := val.(int) + c.Quantity
			s.BingoCardCount.Store(c.WalletId, a)
		} else {
			s.BingoCardCount.Store(c.WalletId, c.Quantity)
		}
	} else if c.PlayCode == "100" {
		s.ExtraPatternsOrders = append(s.ExtraPatternsOrders, c)
		s.ExtraOrderMap[c.OrderNumber] = c
		if val, ok := s.ExtraPatternsCardCount.Load(c.WalletId); ok {
			a := val.(int) + c.Quantity
			s.ExtraPatternsCardCount.Store(c.WalletId, a)
		} else {
			s.ExtraPatternsCardCount.Store(c.WalletId, c.Quantity)
		}
	} else if c.PlayCode == "103" {
		s.ExtraPatternsOrders = append(s.ExtraPatternsOrders, c)
		s.ExtraOrderMap[c.OrderNumber] = c
		s.BingoOrders = append(s.BingoOrders, c)
		s.BingoOrderMap[c.OrderNumber] = c
	}
	s.OrderMap[fmt.Sprintf("%s-%s", c.ID, c.OrderNumber)] = c

	s.CardsName.Store(c.WalletId, c.Nickname)
}

func (s *SettleMegaMerchant) Count(m *SettleMega) {
	if len(s.BingoCards) < 1 && len(s.ExtraPatternsCards) < 1 {
		return
	}
	key := "zset:rank:mega:extra:" + s.Id
	_ = logic.Rdb.Del(context.Background(), key).Err()
	var extraPatternsRank = make([]redis.Z, 0)
	var jackPotRank = make([]redis.Z, 0)
	var bingoRank = make([]redis.Z, 0)
	for _, card := range s.ExtraPatternsCards {
		if len(card.Prize) > 0 {
			var score int
			for _, i3 := range card.Prize {
				score += p2score[i3]
			}
			if order, ok := s.ExtraOrderMap[fmt.Sprintf("%s", card.OrderNumber)]; ok {
				_, amount := s.CalculatePrizeAmount(card, order, m, true)
				count, _ := s.ExtraPatternsCardCount.Load(order.WalletId)
				name, _ := s.CardsName.Load(order.WalletId)
				if amount > 0 && count != nil {
					z := redis.Z{
						Score:  amount,
						Member: fmt.Sprintf("%s-%s-%d-%s", order.WalletId, card.Id, count.(int), name),
					}
					extraPatternsRank = append(extraPatternsRank, z)
				}
			}
		}
	}

	for _, card := range s.BingoCards {
		prizeNum := len(card.Index)
		if prizeNum > 0 {
			unprizeNums := s.FilterUnPrizeNumbers(card)
			if order, ok := s.BingoOrderMap[fmt.Sprintf("%s", card.OrderNumber)]; ok {
				count, _ := s.BingoCardCount.Load(order.WalletId)
				name, _ := s.CardsName.Load(order.WalletId)
				if count != nil {
					z := redis.Z{
						Score:  float64(prizeNum),
						Member: fmt.Sprintf("%s-%s-%d-%s", order.WalletId, unprizeNums, count.(int), name),
					}
					if len(m.PrizeNumber) <= 44 {
						jackPotRank = append(jackPotRank, z)
					}
					bingoRank = append(bingoRank, z)
				}
			}
		}
	}
	//zap.S().Info("未排序前的排行榜中奖数据总数", len(extraPatternsRank))
	sort.SliceStable(extraPatternsRank, func(i, j int) bool {
		return extraPatternsRank[i].Score > extraPatternsRank[j].Score
	})
	if len(extraPatternsRank) > 10 {
		extraPatternsRank = extraPatternsRank[:10]
	}
	if len(extraPatternsRank) > 0 {
		go func(er []redis.Z) {
			err := logic.Rdb.ZAdd(context.Background(), key, er...).Err()
			if err != nil {
				zap.S().Info("err", err.Error())
			} else {
				ms := model.SettleRanKModel{
					Period:     m.PeriodNum,
					RankType:   model.MegaExtraPatternsRank,
					MerchantId: s.Id,
				}
				res, _ := json.Marshal(ms)
				_, err = logic.Js.Publish(consts.GameRank, res)
			}
		}(extraPatternsRank)
	}
	sort.SliceStable(jackPotRank, func(i, j int) bool {
		return jackPotRank[i].Score > jackPotRank[j].Score
	})
	if len(jackPotRank) > 10 {
		jackPotRank = jackPotRank[:10]
	}
	if len(jackPotRank) > 0 {
		go func(jr []redis.Z) {
			err := logic.Rdb.Del(context.Background(), "zset:rank:mega:jackpot:"+s.Id).Err()
			err = logic.Rdb.ZAdd(context.Background(), "zset:rank:mega:jackpot:"+s.Id, jr...).Err()
			if err != nil {
				zap.S().Info("排行榜redis 错误 err", err.Error())
			} else {
				ms := model.SettleRanKModel{
					Period:     m.PeriodNum,
					RankType:   model.MegaJackpotRank,
					MerchantId: s.Id,
				}
				res, _ := json.Marshal(ms)
				_, err = logic.Js.Publish(consts.GameRank, res)
			}
		}(jackPotRank)
	}
	sort.SliceStable(bingoRank, func(i, j int) bool {
		return bingoRank[i].Score > bingoRank[j].Score
	})
	if len(bingoRank) > 10 {
		bingoRank = bingoRank[:10]
	}
	if len(bingoRank) > 0 {
		go func(jr []redis.Z) {
			err := logic.Rdb.Del(context.Background(), "zset:rank:mega:bingo:"+s.Id).Err()
			err = logic.Rdb.ZAdd(context.Background(), "zset:rank:mega:bingo:"+s.Id, jr...).Err()
			if err != nil {
				zap.S().Error("排行榜redis 错误 err", err.Error())
			} else {
				ms := model.SettleRanKModel{
					Period:     m.PeriodNum,
					RankType:   model.MegaBingoRank,
					MerchantId: s.Id,
				}
				res, _ := json.Marshal(ms)
				_, err = logic.Js.Publish(consts.GameRank, res)
			}
		}(bingoRank)
	}
}
func (s *SettleMegaMerchant) FilterUnPrizeNumbers(c *CardN) (nums string) {
	for index, number := range c.Numbers {
		if util.ArraysHasItem(c.Index, index) {
			continue
		}

		if number == 0 {
			continue
		}

		if len(nums) > 0 {
			nums = fmt.Sprintf("%s,%d", nums, number)
			continue
		}

		nums = strconv.Itoa(number)
	}
	return nums
}

func (m *SettleMega) CleanOrderFromRedis() {
	for _, config := range m.MerchantConfig {
		go func(conf *model.MerchantConfig) {
			k1 := fmt.Sprintf("%s%s:BingoMega:%s", RedisPrefix, m.PeriodNum, conf.MerchantID)
			_, err := logic.Rdb.Del(context.Background(), k1).Result()
			if err != nil {
				zap.S().Error("删除redis 订单 失败", err.Error())
			}
		}(config)
	}
}
func (s *SettleMegaMerchant) CleanRedis() {
	_, err := logic.Rdb.Del(context.Background(), "zset:rank:mega:jackpot:"+s.Id).Result()
	if err != nil {
		zap.S().Error("删除mega jackpot 失败", err.Error())
	}
	_, err = logic.Rdb.Del(context.Background(), "zset:rank:mega:bingo:"+s.Id).Result()
	if err != nil {
		zap.S().Error("删除mega jackpot 失败", err.Error())
	}
	_, err = logic.Rdb.Del(context.Background(), "zset:rank:mega:extra:"+s.Id).Result()
	if err != nil {
		zap.S().Error("删除mega bingo 失败", err.Error())
	}
}

func (s *SettleMegaMerchant) CalculatePrizeAmount(card *CardN, order *cmodel.TbBingoOrder, m *SettleMega, rank bool) (name string, amount float64) {
	t1 := mapset.NewSet()
	for i := 0; i < len(card.Prize); i++ {
		t1.Add(fmt.Sprintf("%s", PrizeTypeMap[card.Prize[i]]))
	}
	t := util.EM2(t1)
	for _, ss := range t {
		if strings.Contains(ss, "D") {
			ss = "2L"
		} else if strings.Contains(ss, "G") {
			ss = "3L"
		} else if strings.Contains(ss, "M") {
			ss = "4L"
		}
		if val, ok := m.OddMap[ss]; ok {
			odd, _ := strconv.ParseFloat(val, 0)
			amount = amount + (odd * float64(order.Multiples) / 100)
		} else {
			card.PrizesName = "获取赔率异常"
			zap.S().Error("获取赔率异常", ss)
		}
	}
	//zap.S().Info("extra patterns 中奖类型", t)
	return strings.Join(t, ","), amount
}
