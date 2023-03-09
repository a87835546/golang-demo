package service

import (
	"context"
	"encoding/json"
	"fmt"
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
	"go.uber.org/zap"
	"ph-gitlab.vipsroom.net/bingo_backend/common_core/common"
	cmodel "ph-gitlab.vipsroom.net/bingo_backend/common_core/tables"
)

type SettleRush struct {
	PeriodNum        string //期号
	MerchantServices map[string]*SettleRushMerchant
	MerchantConfig   map[string]*model.MerchantConfig
	PrizeNumber      []int32                      // 开奖号码
	PrizeNumberStr   string                       // 开奖号码
	InitStatus       map[string]consts.DataStatus // key=periodNum  1 initData 2 loadData
	OddMap           map[string]string            //extra patterns 的赔率 key:图案 value:赔率
	JackpotCount     int64
	PrizePool        model.PrizePot
	EsPrizeClient    *helper.EsPrizeClient
}
type SettleRushMerchant struct {
	Id                  string
	mu                  sync.Mutex
	BingoCards          []*CardN                        //所有 bingo and jackpot卡片的集合
	OrderMap            map[string]*cmodel.TbBingoOrder //所有订单的集合 key=wallet-id value=order
	ExtraPatternsOrders []*cmodel.TbBingoOrder          // 这一期中读到内存中一部分数据
	BingoOrders         []*cmodel.TbBingoOrder          // 这一期中读到内存中一部分数据
	CardsMap            map[string]*CardN
	JackpotPrizeCards   []*CardN
	BingoPrizeCards     []*CardN
	BingoOrderMap       map[string]*cmodel.TbBingoOrder //所有订单的集合 key=wallet-id value=order
	ExtraOrderMap       map[string]*cmodel.TbBingoOrder //所有订单的集合 key=wallet-id value=order
	RankSet             []redis.Z
	PrizeMap            sync.Map
	CardCount           sync.Map // 每个用户对应所有的卡片总数 key = wallet id value = count
	CardsName           sync.Map // 每个用户对应所有的卡片总数 key = wallet id value = name
}

func NewSettleRush() *SettleRush {
	InitPatternsRush()
	return &SettleRush{
		InitStatus:       make(map[string]consts.DataStatus, 0),
		MerchantServices: make(map[string]*SettleRushMerchant, 0),
		MerchantConfig:   make(map[string]*model.MerchantConfig, 0),
		OddMap:           make(map[string]string, 0),
		PrizePool:        model.PrizePot{},
		EsPrizeClient:    helper.NewEsPrizeClient(logic.EsClient),
	}
}
func (s *SettleRush) GameEvent(event *model.GameRoundEvent) {
	//zap.S().Info("发球数据--->>", event)
	status := event.GetStatus()
	if status == "" {
		zap.S().Error("bad data")
		return
	}
	s.PeriodNum = event.GameRoundId

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	now := time.Now()
	switch status {
	case consts.Begin:
		s.Clean()
		//zap.S().Info("Begin")
		s.ReadingPeriod(ctx)
		err := s.InitData(ctx, event.GameRoundId)
		s.ReadingOdd(ctx, []string{"BGR:102", "BGR:100"})
		if err != nil {
			zap.S().Error("Begin InitData fail err=", err)
			return
		}
		break
	case consts.PrizeOnGoing:
		s.PrizeNumberStr = event.PrizeNumber
		if len(s.MerchantServices) == 0 {
			s.ReadingPeriod(ctx)
			_ = s.InitData(ctx, event.GameRoundId)
			s.ReadingOdd(ctx, []string{"BGR:102", "BGR:100"})
			s.ReadingPrizePool()
		}
		s.PrizeOnGoing(event)
		s.CalculatePrizeMoney()
		break
	case consts.Finish:
		//zap.S().Info("clean")
		s.CleanOrderFromRedis()
		s.Payment()
		break
	case consts.Cancel:
		//zap.S().Info("cancel")
		//s.ReadingPeriod(ctx)
		//_ = s.InitData(ctx, event.GameRoundId)
		//s.Cancel()
		if s.PeriodNum == event.GameRoundId {
			s.Clean()
		} else {
			s.EsPrizeClient.DeleteByCondition(event.GameRoundId)
		}
		break
	}
	if time.Now().UnixMilli()-now.UnixMilli() > 1000 {
		zap.S().Info("settle time spend ", time.Since(now))
	}
	return
}
func (s *SettleRush) InitData(ctx context.Context, periodNum string) (err error) {
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
			s.MerchantServices[temp.MerchantID] = &SettleRushMerchant{}
			s.MerchantConfig[temp.MerchantID] = &temp
		}
	}
	for key, merchant := range s.MerchantServices {
		merchant.InitData(key)
	}
	s.InitStatus[periodNum] = consts.InitData
	return
}
func (s *SettleRushMerchant) Clean() {
	s.BingoCards = make([]*CardN, 0)
	s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.BingoOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
	s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
	s.CardsMap = make(map[string]*CardN, 0)
	s.JackpotPrizeCards = make([]*CardN, 0)
	s.BingoPrizeCards = make([]*CardN, 0)
	s.RankSet = make([]redis.Z, 0)
	s.mu = sync.Mutex{}
	s.CardsName = sync.Map{}
	s.CardCount = sync.Map{}
}
func (s *SettleRushMerchant) InitData(id string) {
	s.Id = id
	s.mu = sync.Mutex{}
	s.BingoCards = make([]*CardN, 0)
	s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.BingoOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraOrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
	s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
	s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
	s.CardsMap = make(map[string]*CardN, 0)
	s.JackpotPrizeCards = make([]*CardN, 0)
	s.BingoPrizeCards = make([]*CardN, 0)
	s.RankSet = make([]redis.Z, 0)
	s.CardsName = sync.Map{}
	s.CardCount = sync.Map{}
}

func (s *SettleRush) ReadingOdd(ctx context.Context, gameType []string) (err error) {
	if len(gameType) == 0 {
		gameType = []string{"BGR:100"}
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
					s.OddMap[category.BingoPatten] = category.Bonus
					if len(category.BonusPotMin) > 1 {
						s.OddMap[category.BingoPatten] = category.BonusPotMin
					}
				}
			}
		}
	}
	//zap.S().Info("bingo rush odd", s.OddMap)
	return nil
}
func (s *SettleRush) ReadingPeriod(ctx context.Context) error {
	res, err := logic.Rdb.HGet(ctx, consts.MerchantPeriodRedis, "BGR").Result()
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
			s.PeriodNum = mp.GameRoundId
		}
	}
	return nil

} // 获取期号
func (s *SettleRush) ReadOrders(ctx context.Context, periodNum string) {
	for index, config := range s.MerchantConfig {
		k1 := fmt.Sprintf("%s%s:BingoRush:%s", RedisPrefix, periodNum, config.MerchantID)
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
				wg.Add(1)
				ss := s.MerchantServices[index]
				go func(k, val string, memoryService *SettleRushMerchant) {
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
			for id, merchant := range s.MerchantServices {
				if len(merchant.BingoCards) > 0 {
					zap.S().Info("商户", id, "  卡片订单数 ", len(merchant.BingoCards))
				}
			}
		}
	}
	s.InitStatus[periodNum] = consts.LoadData
}
func (s *SettleRushMerchant) AddCards(c *cmodel.TbCard, v *cmodel.TbBingoOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.PlayCode == "102" {
		card := new(CardN)
		card.Str2NumbersRush(c.Numbers)
		card.TbCard = *c
		card.Id = c.Id
		s.BingoCards = append(s.BingoCards, card)
		s.CardsMap[fmt.Sprintf("%s-%s", c.Id, v.ID)] = card
		s.BingoOrderMap[c.OrderNumber] = v
	}
}

func (s *SettleRushMerchant) AddOrderMap(c *cmodel.TbBingoOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.PlayCode == "102" {
		s.BingoOrders = append(s.BingoOrders, c)
	}
	s.OrderMap[fmt.Sprintf("%s-%s", c.ID, c.OrderNumber)] = c
	if val, ok := s.CardCount.Load(c.WalletId); ok {
		a := val.(int) + c.Quantity
		s.CardCount.Store(c.WalletId, a)
	} else {
		s.CardCount.Store(c.WalletId, c.Quantity)
	}
	s.CardsName.Store(c.WalletId, c.Nickname)
}
func (s *SettleRush) UpdatePrizePool() {
	res, _ := json.Marshal(s.PrizePool)
	_, err := logic.Rdb.HSet(context.Background(), consts.MerchantPrizePoolRedis, "BGR", res).Result()
	if err != nil {
		zap.S().Info("更新奖池信息失败", err.Error())
	}
}
func (s *SettleRush) ReadingPrizePool() {
	res, err := logic.Rdb.HGet(context.Background(), consts.MerchantPrizePoolRedis, "BGR").Result()
	if err != nil {
		zap.S().Info("读取redis 商户失败", err.Error())
	} else {
		json.Unmarshal([]byte(res), &s.PrizePool)
		if err != nil {
			zap.S().Info("读取奖池信息出错", err.Error())
		}
	}
	//zap.S().Info("bingo rush 的奖池数据", s.PrizePool.String())
}

func (s *SettleRush) PrizeOnGoing(event *model.GameRoundEvent) {
	numbers, err := event.GetPrizeNumber()
	if err != nil {
		zap.S().Error("GetPrizeNumber fail err=", err)
		return
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	status, ok := s.InitStatus[event.GameRoundId]
	if !ok {
		err := s.InitData(ctx, event.GameRoundId)
		if err != nil {
			zap.S().Error("PrizeOnGoing InitData fail err=", err)
			return
		}
		s.ReadOrders(ctx, event.GameRoundId)
	} else {
		if status == consts.UnKnownStatus {
			zap.S().Error("UnKnownStatus")
			return
		}
		if status == consts.InitData {
			s.ReadOrders(ctx, event.GameRoundId)
			s.ReadingPrizePool()
		}
		verifyNumbers, b := s.VerifyNumbers(numbers)
		if !b {
			for _, number := range verifyNumbers {
				bingoCards := make([]*CardN, 0)
				for _, merchant := range s.MerchantServices {
					bingoCards = append(bingoCards, merchant.BingoCards...)
				}
				if len(bingoCards) > 0 {
					s.BingoJackPotSettle(bingoCards, int(number), nil)
				}
				s.PrizeNumber = append(s.PrizeNumber, number)
			}
		} else {
			number, err := event.GetLastPrizeNumber()
			if err != nil {
				zap.S().Error("GetLastPrizeNumber fail err=", err)
				return
			}
			s.PrizeNumber = append(s.PrizeNumber, number)
			bingoCards := make([]*CardN, 0)
			for _, merchant := range s.MerchantServices {
				bingoCards = append(bingoCards, merchant.BingoCards...)
			}
			if len(s.PrizeNumber) == 26 {
				wg := sync.WaitGroup{}
				if len(bingoCards) > 0 {
					wg.Add(1)
					s.BingoJackPotSettle(bingoCards, int(number), &wg)
				}
				wg.Wait()
				s.M()
			} else {
				if len(bingoCards) > 0 {
					s.BingoJackPotSettle(bingoCards, int(number), nil)
				}
			}

		}
	}
}

//	func (s *SettleRush) Cancel() {
//		tables := make([]string, 0)
//		for _, config := range s.MerchantConfig {
//			tables = append(tables, config.SettledTableName)
//		}
//		err := OrderStub.RpcCancelOrder(s.PeriodNum, consts.BGR)
//		if err != nil {
//			fmt.Println(err)
//			return
//		}
//	}
func (s *SettleRush) M() {
	for _, merchant := range s.MerchantServices {
		for _, card := range merchant.BingoCards {
			if card.Jackpot {
				order := merchant.BingoOrderMap[card.OrderNumber]
				s.JackpotCount = s.JackpotCount + int64(order.Multiples)
				merchant.JackpotPrizeCards = append(merchant.JackpotPrizeCards, card)
			}
		}
	}
	zap.S().Info("bingo rush jackpot 中奖个数", s.JackpotCount)
}
func (s *SettleRush) Payment() {
	wg := sync.WaitGroup{}
	jamount := float64(0)
	if s.JackpotCount > 0 {
		total, _ := strconv.ParseFloat(s.PrizePool.JackpotBGR, 0)
		jamount = total / float64(s.JackpotCount)
	}

	mr := make([]*model.JackpotRecordModel, 0)
	ranks := make([]*model.RankHistoryModel, 0)
	for _, merchant := range s.MerchantServices {
		//TODO
		var transfer = make([]*model.TransferRequestBody, 0)
		var temp = make([]*cmodel.TbBingoOrder, 0)
		wg.Add(1)
		func(ms *SettleRushMerchant) {
			for _, order := range ms.OrderMap {
				cards1 := make([]cmodel.TbCard, 0, len(order.BetContent))
				total := float64(0)
				jackpot := float64(0)
				extra := float64(0)
				for _, card := range order.BetContent {
					ms.mu.Lock()
					val, ok := ms.CardsMap[fmt.Sprintf("%s-%s", card.Id, order.ID)]
					if ok {
						amount := float64(0)
						if order.PlayCode == "102" {
							// jackpot
							if s.JackpotCount > 0 && val.Jackpot {
								amount = amount + (jamount * float64(order.Multiples))
								jackpot = jackpot + amount
								card.PrizesAmount = FormatFloat(jamount*float64(order.Multiples), 2)
								card.PrizesType = 1
								if amount > 0 {
									transfer = append(transfer, s.ConfigOrderRecording(amount, order))
								}
								total = total + amount
								mr = append(mr, ConfigBingoPrizePool(3, amount, s.PeriodNum))
							}

							if !val.Jackpot && len(val.Prize) > 0 {
								// extra patterns
								name, amount1 := ms.CalculatePrizeAmount(val, order, s, false)
								total = total + amount1
								card.PrizesType = 3
								card.PrizesName = name
								amount = amount1
								extra = extra + amount
								card.PrizesAmount = FormatFloat(amount, 2)
								if amount > 0 {
									transfer = append(transfer, s.ConfigOrderRecording(amount, order))
								}
							}
						}
						card.PrizesContent = s.PrizeNumberStr
						cards1 = append(cards1, *card)
					}
					ms.mu.Unlock()
				}
				if total > 0 {
					ranks = append(ranks, ConfigRankHistoryModel(FormatFloat(total, 2), 3, s.PeriodNum, order))
					if val, ok := ms.PrizeMap.Load(order.WalletId); ok {
						if jackpot > 0 {
							pm := val.(*PrizeModel)
							pm.Jackpot = pm.Jackpot + jackpot
							ms.PrizeMap.Store(order.WalletId, pm)
						}
						if extra > 0 {
							pm := val.(*PrizeModel)
							pm.Extra = pm.Extra + extra
							ms.PrizeMap.Store(order.WalletId, pm)
						}
					} else {
						if jackpot > 0 {
							ms.PrizeMap.Store(order.WalletId, &PrizeModel{Jackpot: jackpot, MerchantId: ms.Id, PeriodNum: s.PeriodNum, IsRush: true})
						}
						if extra > 0 {
							ms.PrizeMap.Store(order.WalletId, &PrizeModel{Extra: extra, MerchantId: ms.Id, PeriodNum: s.PeriodNum, IsRush: true})
						}
						if jackpot > 0 && extra > 0 {
							ms.PrizeMap.Store(order.WalletId, &PrizeModel{Jackpot: jackpot, Extra: extra, MerchantId: ms.Id, PeriodNum: s.PeriodNum, IsRush: true})
						}
					}
				}
				order.Status = 2
				order.LotteryContent = s.PrizeNumberStr
				if total > 0 {
					order.WinLose = fmt.Sprintf("+%.2f", total)
				} else if total == 0 {
					order.WinLose = fmt.Sprintf("%.2f", total)
				} else {
					order.WinLose = fmt.Sprintf("-%.2f", total)
				}
				order.AwardAmount = fmt.Sprintf("%.2f", total)
				order.UpdatedAt = time.Now().UnixMilli()
				temp = append(temp, order)
			}
			tmp := ConvertMapToArray(&ms.PrizeMap)
			if len(tmp) > 0 {
				PublishFinalResult(tmp)
			}
			ms.mu.Lock()
			mss, ok := s.MerchantConfig[ms.Id]
			ms.mu.Unlock()
			if ok && len(temp) > 0 {
				zap.S().Info("结算的订单总数--->>", len(temp))
				BatchSaveOrder(temp, mss.SettledTableName)
				if len(transfer) > 0 {
					zap.S().Info("中奖的卡片总数--->>", len(transfer))
					Payout(transfer)
				}
			}
			wg.Done()
		}(merchant)
	}
	wg.Wait()
	if len(ranks) > 0 {
		go s.EsPrizeClient.Save(ranks)
	}
	if s.JackpotCount > 0 {
		s.CheckBingoPrize(mr)
		s.PrizePool.JackpotBGR = s.OddMap["26B"]
		s.UpdatePrizePool()
		PublishPrizePool(s.PrizePool, s.OddMap, 3, s.PeriodNum)
	}
}
func (s *SettleRush) CheckBingoPrize(list []*model.JackpotRecordModel) {
	bb, _ := strconv.ParseInt(s.PrizePool.JackpotBGR, 10, 64)
	aa, _ := strconv.ParseInt(s.OddMap["26B"], 10, 64)
	if bb < aa {
		s.PrizePool.Bingo = fmt.Sprintf("%d", aa)
		s.UpdatePrizePool()
	}
	if len(list) > 0 {
		zap.S().Info("bingo rush bingo:", s.JackpotCount)
		PublishBingoPrizePool(list)
	}
}

// ConfigOrderRecording 派彩
func (s *SettleRush) ConfigOrderRecording(amount float64, order *cmodel.TbBingoOrder) *model.TransferRequestBody {
	t := s.MerchantConfig[order.MerchantID]
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
func (s *SettleRush) BingoJackPotSettle(bingoCards []*CardN, num int, wg *sync.WaitGroup) {
	p := SettleNum
	group := sync.WaitGroup{}
	l := len(bingoCards)
	count := l / p
	for j := 0; j < p; j++ {
		group.Add(1)
		i1 := j * count
		i3 := count * (j + 1)
		ints := bingoCards[i1:i3]
		go func(in []*CardN, n int) {
			for _, i4 := range in {
				if i4.GenIndexRush(n) {
					i4.ExtraPatternsSettleRush()
					if len(s.PrizeNumber) <= 26 {
						i4.BingoJackPotRushSettle()
					}
				}

			}
			group.Done()
		}(ints, num)
	}
	if (count == 0 && l > 0) || l%p != 0 {
		ints := bingoCards[count*p:]
		for _, card := range ints {
			if card.GenIndexRush(num) {
				card.ExtraPatternsSettleRush()
				if len(s.PrizeNumber) <= 26 {
					card.BingoJackPotRushSettle()
				}
			}

		}
	}
	group.Wait()
	if wg != nil {
		wg.Done()
	}
}
func (s *SettleRush) VerifyNumbers(numbers []int32) ([]int32, bool) {
	i := len(numbers) - len(s.PrizeNumber)
	if i == 1 {
		return nil, true
	}
	if len(s.PrizeNumber) == 0 {
		return numbers, false
	}
	array := util.DiffArray(numbers, s.PrizeNumber)
	return array, false
}
func (s *SettleRush) Clean() {
	s.PeriodNum = ""
	for _, merchant := range s.MerchantServices {
		merchant.Clean()
		merchant.CleanRedis()
	}
	s.PrizeNumber = nil
	s.JackpotCount = 0
}

func (s *SettleRush) CalculatePrizeMoney() {
	jamount := float64(0)
	if s.JackpotCount > 0 {
		total, _ := strconv.ParseFloat(s.PrizePool.JackpotBGR, 0)
		jamount = total / float64(s.JackpotCount)
	}
	for _, merchant := range s.MerchantServices {
		go func(ms *SettleRushMerchant) {
			if len(ms.RankSet) > 0 {
				ms.RankSet = make([]redis.Z, 0)
			}
			wg := sync.WaitGroup{}
			for _, order := range ms.OrderMap {
				wg.Add(1)
				go func(bingoOrder *cmodel.TbBingoOrder) {
					for _, card := range bingoOrder.BetContent {
						amount := float64(0)
						if val, ok := ms.CardsMap[fmt.Sprintf("%s-%s", card.Id, bingoOrder.ID)]; ok {
							if s.JackpotCount > 0 {
								if val.Jackpot {
									amount = amount + (jamount * float64(bingoOrder.Multiples))
								}
							}
							if !val.Jackpot && len(val.Prize) > 0 {
								// extra patterns
								_, amount1 := ms.CalculatePrizeAmount(val, bingoOrder, s, true)
								amount = amount + amount1
							}
							if amount > 0 {
								ms.mu.Lock()
								count, _ := ms.CardCount.Load(bingoOrder.WalletId)
								name, _ := ms.CardsName.Load(bingoOrder.WalletId)
								if count != nil {
									ms.RankSet = append(ms.RankSet, redis.Z{
										Score:  amount,
										Member: fmt.Sprintf("%s-%s-%d-%s", bingoOrder.WalletId, card.Id, count.(int), name),
									})
								}
								ms.mu.Unlock()
							}
						}
					}
					wg.Done()
				}(order)
			}
			wg.Wait()
			ms.Rank(s.PeriodNum)
		}(merchant)
	}
}
func (s *SettleRushMerchant) FilterUnPrizeNumbers(c *CardN) (nums string) {
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
func (s *SettleRushMerchant) CleanRedis() {
	_, err := logic.Rdb.Del(context.Background(), "zset:rank:rush:jackpot:"+s.Id).Result()
	if err != nil {
		zap.S().Info("删除mega jackpot 失败", err.Error())
	}
}

func (s *SettleRushMerchant) Rank(m string) {
	if len(s.RankSet) > 0 {
		go func(rs []redis.Z) {
			err := logic.Rdb.Del(context.Background(), "zset:rank:rush:jackpot:"+s.Id).Err()
			sort.SliceStable(s.RankSet, func(i, j int) bool {
				if i > len(s.RankSet) || j > len(s.RankSet) {
					return false
				}
				return s.RankSet[i].Score > s.RankSet[j].Score
			})
			if len(s.RankSet) > 10 {
				s.RankSet = s.RankSet[:10]
			}
			err = logic.Rdb.ZAdd(context.Background(), "zset:rank:rush:jackpot:"+s.Id, rs...).Err()
			if err != nil {
				zap.S().Info("err", err.Error())
			} else {
				ms := model.SettleRanKModel{
					Period:     m,
					RankType:   model.RushJackpotRank,
					MerchantId: s.Id,
				}
				res, _ := json.Marshal(ms)
				err = logic.Nc.Publish(consts.GameRank, res)
				if err != nil {
					zap.S().Info("rush 排行榜插入数据异常", err.Error())
				}
			}
		}(s.RankSet)
	}
}
func (s *SettleRushMerchant) CalculatePrizeAmount(card *CardN, order *cmodel.TbBingoOrder, m *SettleRush, rank bool) (name string, amount float64) {
	t1 := mapset.NewSet()
	for i := 0; i < len(card.Prize); i++ {
		t1.Add(fmt.Sprintf("%s", RushPrizeTypeMap[card.Prize[i]]))
	}
	t := util.EM3(t1)
	for _, ss := range t {
		if strings.Contains(ss, "1L") {
			ss = "1L"
		} else if strings.Contains(ss, "2L") {
			ss = "2L"
		}
		if val, ok := m.OddMap[ss]; ok {
			odd, _ := strconv.ParseFloat(val, 0)
			amount = amount + odd*float64(order.Multiples)
		} else {
			card.PrizesName = "获取赔率异常"
			zap.S().Info("获取赔率异常", ss)
		}
	}
	return strings.Join(t, ","), amount / 100
}
func (s *SettleRush) CleanOrderFromRedis() {
	for _, config := range s.MerchantConfig {
		go func(conf *model.MerchantConfig) {
			k1 := fmt.Sprintf("%s%s:BingoRush:%s", RedisPrefix, s.PeriodNum, conf.MerchantID)
			_, err := logic.Rdb.Del(context.Background(), k1).Result()
			if err != nil {
				zap.S().Error("删除redis 订单 失败", err.Error())
			}
		}(config)
	}
}
