package service

import (
	"context"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"math"
	"settlement_go/internal/consts"
	"settlement_go/internal/logic"
	"settlement_go/internal/model"
	"settlement_go/internal/util"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"ph-gitlab.vipsroom.net/bingo_backend/common_core/helper"
	cmodel "ph-gitlab.vipsroom.net/bingo_backend/common_core/tables"
)

var RedisPrefix = "bingo:order:"
var jsonOwn = jsoniter.ConfigCompatibleWithStandardLibrary
var p2score = map[int]int{0: 1, 1: 2, 2: 3, 4: 4, 5: 4, 6: 4, 7: 4, 8: 4, 9: 4, 10: 4, 11: 4, 12: 4, 13: 5, 14: 5, 15: 7}

// UpdateOrder 更新当前商户的订单
func UpdateOrder(orders []*cmodel.TbBingoOrder, merchantId string) {
	temp := make(map[string][]cmodel.TbBingoOrder, 0)
	for i := 0; i < len(orders); i++ {
		order := orders[i]
		order.CreatedAt = time.Now().UnixMilli()
		order.UpdatedAt = time.Now().UnixMilli()
		if val, ok := temp[order.MerchantID]; ok {
			val = append(val, *order)
		} else {
			temp[order.MerchantID] = []cmodel.TbBingoOrder{*order}
		}
	}
	for s2, bingoOrders := range temp {
		go func(key string, val []cmodel.TbBingoOrder) {
			//UpdateOrderMango(merchantId, val...)
		}(s2, bingoOrders)
	}
}
func BatchSaveOrder(list []*cmodel.TbBingoOrder, tableName string) error {
	now := time.Now()
	if len(list) == 0 {
		return nil
	}
	zap.S().Info("批量结算的订单总数--->>>", len(list), "订单的表名--->>", tableName)
	length := 10000
	orderArr, err := helper.SplitArray[*cmodel.TbBingoOrder](list, length)
	if err != nil {
		zap.S().Error("切分订单数组异常", err.Error())
		return errors.WithStack(err)
	}

	var (
		wg = sync.WaitGroup{}
	)
	wg.Add(len(orderArr))
	for _, orders := range orderArr {
		go func(orders []*cmodel.TbBingoOrder, tbn string) {
			defer wg.Done()
			for _, order := range orders {
				md := model.OrderNats{
					TableName: tbn,
					Data:      order,
				}
				marshal, err2 := json.Marshal(md)
				if err2 != nil {
					zap.S().Error("Marshal fail err=", err.Error())
					return
				}
				util.PublishMsg(logic.PublishChan, consts.OrderUpdate, marshal)
			}
		}(orders, tableName)
	}
	wg.Wait()
	if time.Now().UnixMilli()-now.UnixMilli() > 1000 {
		zap.S().Info("写入订单的时长", time.Since(now))
	}
	return nil
}

func SaveOrder(dataList []cmodel.TbBingoOrder, tableName string) error {

	sql, _, err := logic.G.From(tableName).Insert().Rows(dataList).ToSQL()
	if err != nil {
		zap.S().Infof("保存订单-sql-异常: %s", err.Error())
		return err
	}
	// 保存订单
	_, err = logic.Db.Exec(sql)
	if err != nil {
		zap.S().Infof("保存订单-sql-执行失败: %s table name: %s", err.Error(), tableName)
		return err
	}
	return nil
}
func UpdateOrderMango(tableName string, dataList ...*cmodel.TbBingoOrder) error {
	for _, order := range dataList {
		_, err := logic.Mdb.Collection(tableName).ReplaceOne(
			context.TODO(), // 获取上下文参数
			bson.M{"_id": order.ID},
			order,
		)
		if err != nil {
			zap.S().Infof("保存订单-sql-执行失败: %s table name: %s", err.Error(), tableName)
			return err
		}
	}
	return nil
}

func SaveOrder1(dataList []model.SettleOrdersModel) error {
	sql, _, err := logic.G.From(dataList[0].TableName).Insert().Rows(dataList).ToSQL()
	if err != nil {
		zap.S().Infof("保存订单-sql-异常: %s", err.Error())
		return err
	}
	// 保存订单
	_, err = logic.Db.Exec(sql)
	if err != nil {
		zap.S().Infof("保存订单-sql-执行失败: %s", err.Error())
		return err
	}
	return nil
}

/*
	func initRushData(MarketType consts.MarketType, all *RushSettleUseMemoryAllMerchantsService) {
		all.Lottery = make([]string, 0)
		all.Index = 0
		all.MerchantConfig = make(map[string]*model.MerchantConfig, 0)
		all.MerchantServices = make(map[string]*RushSettleUseMemoryService, 0)
		all.OddMap = sync.Map{}
		all.GameType = "BGM:100"
		all.PeriodNum = "2022116112300010"
		if MarketType == consts.MarketTypeMega {
			all.Count = int(consts.MarketCountMega)
		} else if MarketType == consts.MarketTypeRush {
			all.Count = int(consts.MarketCountRush)
		}
		all.pattens = consts.Pattens
		zap.S().Info("init data")
	}

	func initRushServiceItemData(s *RushSettleUseMemoryService) {
		s.mu = sync.Mutex{}
		s.Orders = make([]cmodel.TbBingoOrder, 0)
		s.OrderMap = make(map[string]cmodel.TbBingoOrder, 0)
		s.CardsMap = make(map[string][]*cmodel.TbCard, 0)
		s.RankSet = make(map[string]*zset.SortedSet, 0)
		s.ExtraPatternsRankSet = make(map[string]*zset.SortedSet, 0)
		s.Cache = syncmap.Map{}
		s.PrizeSettlement = sync.Map{}
	}

	func initData(MarketType consts.MarketType) {
		var all = GetSettleUseMemoryAllService()
		all.Lottery = make([]string, 0)
		all.Index = 1
		all.MerchantConfig = make(map[string]*model.MerchantConfig, 0)
		all.MerchantServices = make(map[string]*SettleUseMemoryService, 0)
		all.OddMap = sync.Map{}
		all.GameType = "BGM:100"
		all.PeriodNum = "2022116112300033"

		if MarketType == consts.MarketTypeMega {
			all.Count = int(consts.MarketCountMega)
		} else if MarketType == consts.MarketTypeRush {
			all.Count = int(consts.MarketCountRush)
		}
		all.pattens = consts.Pattens
		zap.S().Info("init data")
	}

	func initRedisData(MarketType consts.MarketType) {
		var all = GetSettleUseRedisService()
		all.Lottery = make([]string, 0)
		all.Index = 1
		all.MerchantConfig = make(map[string]*model.MerchantConfig, 0)
		all.MerchantServices = make(map[string]*SettleUseRedisService, 0)
		all.OddMap = sync.Map{}
		all.GameType = "BGM:100"
		all.PeriodNum = "OKBGM2212260118"
		if MarketType == consts.MarketTypeMega {
			all.Count = int(consts.MarketCountMega)
		} else if MarketType == consts.MarketTypeRush {
			all.Count = int(consts.MarketCountRush)
		}
		all.pattens = consts.ExtraPatterns()
		all.Keys = make([]string, 0)
		zap.S().Info("init data")
	}

	func initRedisData1(MarketType consts.MarketType) {
		var all = GetSettleTwoService()
		all.Lottery = make([]string, 0)
		all.Index = 1
		all.NoData = true
		all.MerchantConfig = make(map[string]*model.MerchantConfig, 0)
		all.MerchantServices = make(map[string]*SettleTwoService, 0)
		all.OddMap = sync.Map{}
		all.GameType = "BGM:100"
		all.PeriodNum = "OKBGM2212260073"
		all.PrizePattens = make(map[string]consts.MatchesType, 0)
		if MarketType == consts.MarketTypeMega {
			all.Count = int(consts.MarketCountMega)
		} else if MarketType == consts.MarketTypeRush {
			all.Count = int(consts.MarketCountRush)
		}
		all.pattens = consts.ExtraPatterns()
		zap.S().Info("init data")
	}

	func initServiceItemData(s *SettleUseMemoryService) {
		s.mu = sync.Mutex{}
		s.Orders = make([]cmodel.TbBingoOrder, 0)
		s.OrderMap = make(map[string]cmodel.TbBingoOrder, 0)
		s.CardsMap = make(map[string][]*cmodel.TbCard, 0)
		s.RankSet = make(map[string]*zset.SortedSet, 0)
		s.ExtraPatternsRankSet = make(map[string]*zset.SortedSet, 0)
		s.Cache1 = syncmap.Map{}
		s.PrizeSettlement1 = &sync.Map{}
	}

	func initRedisServiceItemData(s *SettleUseRedisService) {
		s.mu = sync.Mutex{}
		s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
		s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
		s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
		s.ExtraPatternsCards = make([]*cmodel.TbCard, 0)
		s.BingoCards = make([]*cmodel.TbCard, 0)
		s.RankSet = make(map[string]*zset.SortedSet, 0)
		s.ExtraPatternsRankSet = make(map[string]*zset.SortedSet, 0)
		s.Cache = syncmap.Map{}
		s.PrizeSettlement = sync.Map{}
		s.PrizeSet = zset.New()
		s.CardsMap = sync.Map{}
		s.OneTgPrizeMap = sync.Map{}
		s.TwoTgPrizeMap = sync.Map{}
		s.RedisKeys = make([]string, 0)
		s.BingoPrizeSettlement = make([]string, 0)
		s.JackpotPrizeSettlement = make([]string, 0)
		s.PrizePattens = make(map[string]consts.MatchesType, 0)
	}

	func initRedisServiceItemData2(s *SettleTwoService, key string) {
		s.mu = sync.Mutex{}
		s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
		s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
		s.ExtraPatternsCards = make([]*Card, 0)
		s.BingoCards = make([]*Card, 0)
		s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
		s.RankSet = make(map[string]*zset.SortedSet, 0)
		s.ExtraPatternsRankSet = make(map[string]*zset.SortedSet, 0)
		s.Cache = syncmap.Map{}
		s.PrizeSettlement = sync.Map{}
		s.PrizeSet = zset.New()
		s.OneTgPrizeMap = sync.Map{}
		s.TwoTgPrizeMap = sync.Map{}
		s.RedisKeys = make([]string, 0)
		s.Merchant = key
	}

	func initRedisWithMemoryServiceItemData(s *SettleUseRedisWithMemoryService) {
		s.mu = sync.Mutex{}
		s.Orders = make([]*cmodel.TbBingoOrder, 0)
		s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
		s.Cards = make([]*cmodel.TbCard, 0)
		s.RankSet = make(map[string]*zset.SortedSet, 0)
		s.ExtraPatternsRankSet = make(map[string]*zset.SortedSet, 0)
		s.Cache = syncmap.Map{}
		s.PrizeSettlement = sync.Map{}
		s.PrizeSet = zset.New()
		s.CardsMap = sync.Map{}
		s.OneTgPrizeMap = sync.Map{}
		s.TwoTgPrizeMap = sync.Map{}
		s.RedisKeys = make([]string, 0)
	}
*/
func GetMaxValue(set []interface{}) int32 {
	maxVal := set[0]
	for i := 1; i < len(set); i++ {
		if helper.ConvertAnyToInt32(maxVal) < helper.ConvertAnyToInt32(set[i]) {
			maxVal = set[i]
		}
	}
	return helper.ConvertAnyToInt32(maxVal)
}
func GetMapAllKeys(m map[string]struct{}) []string {
	j := 0
	keys := make([]string, len(m))
	for s2, _ := range m {
		keys[j] = s2
		j++
	}
	return keys
}
func GetSyncMapAllKeys(m *sync.Map) int {
	j := 0
	m.Range(func(key, value any) bool {

		j++
		return true
	})
	return j
}
func ConvertMapToArray(m *sync.Map) []*PrizeModel {
	var temp []*PrizeModel
	m.Range(func(key, value any) bool {
		t := value.(*PrizeModel)
		t.WalletId = key.(string)
		temp = append(temp, t)
		return true
	})
	return temp
}

/*
	func (s *SettleUseRedisService) Clean() {
		s.ExtraPatternsOrders = make([]*cmodel.TbBingoOrder, 0)
		s.BingoOrders = make([]*cmodel.TbBingoOrder, 0)
		s.OrderMap = make(map[string]*cmodel.TbBingoOrder, 0)
		s.ExtraPatternsCards = make([]*cmodel.TbCard, 0)
		s.BingoCards = make([]*cmodel.TbCard, 0)
		s.RankSet = make(map[string]*zset.SortedSet, 0)
		s.ExtraPatternsRankSet = make(map[string]*zset.SortedSet, 0)
		s.Cache = syncmap.Map{}
		s.PrizeSettlement = sync.Map{}
		s.PrizeSet = zset.New()
		s.CardsMap = sync.Map{}
		s.OneTgPrizeMap = sync.Map{}
		s.TwoTgPrizeMap = sync.Map{}
		s.RedisKeys = make([]string, 0)
		s.BingoPrizeSettlement = make([]string, 0)
		s.JackpotPrizeSettlement = make([]string, 0)
		s.PrizePattens = make(map[string]consts.MatchesType, 0)

}

// Clean 一期完成清除

	func (s *SettleUseRedisAllMerchant) Clean() {
		s.Lottery = make([]string, 0)
		s.Index = 1
		s.Keys = make([]string, 0)
		s.OddMap = sync.Map{}
		s.MerchantConfig = make(map[string]*model.MerchantConfig, 0)
		s.MerchantServices = make(map[string]*SettleUseRedisService, 0)
		for _, redisService := range s.MerchantServices {
			redisService.Clean()
		}
		defer debug.FreeOSMemory()
	}
*/
func ConfigBingoPrizePool(t int, amount float64, period string) *model.JackpotRecordModel {
	mm := model.JackpotRecordModel{
		IssueNumber: period,
		CreatedAt:   time.Now().UnixMicro(),
	}
	a1 := decimal.NewFromFloat(amount)
	if a1.IsZero() {
		zap.S().Error("奖池构建的数据", amount)
	}
	mm.Amount = a1
	mm.RecordType = t
	mm.Remark = "奖池池底扣款"
	mm.InOut = -1
	return &mm
}
func ConfigRankHistoryModel(amount string, t int, period string, order *cmodel.TbBingoOrder) *model.RankHistoryModel {
	return &model.RankHistoryModel{
		OrderId:    order.ID,
		WalletId:   order.WalletId,
		Multi:      order.Multiples,
		Amount:     amount,
		Type:       t,
		Period:     period,
		CreatedAt:  time.Now().UnixMilli(),
		MerchantId: order.MerchantID,
		CardNum:    order.Quantity,
		UserId:     order.MerchantUserId,
		Nickname:   order.Nickname,
	}
}
func PublishPrizePool(p model.PrizePot, m map[string]string, t int, period string) {
	a, _ := decimal.NewFromString(m["44B"])
	mm := model.JackpotRecordModel{
		Amount:      a,
		RecordType:  1,
		IssueNumber: period,
		CreatedAt:   time.Now().UnixMilli(),
	}
	switch t {
	case 2:
		a, _ := decimal.NewFromString(p.Bingo)
		mm.Amount = a
		mm.RecordType = 2
		mm.Remark = "BGM Bingo 奖池池底"

		break
	case 1:
		a, _ := decimal.NewFromString(p.JackpotBGM)
		mm.Amount = a
		mm.RecordType = 2
		mm.Remark = "BGM Jackpot 奖池池底"
		break
	case 3:
		a, _ := decimal.NewFromString(p.JackpotBGR)
		mm.Amount = a
		mm.RecordType = 3
		mm.Remark = "BGR Jackpot 奖池池底"
		break
	}
	mm.InOut = 1
	res, err := json.Marshal([]model.JackpotRecordModel{mm})
	if err != nil {
		zap.S().Info("构建数据异常", err.Error())
	}
	err = logic.Nc.Publish(consts.GameJackpotPool, res)
	if err != nil {
		zap.S().Info("发布奖池池底数据异常", err.Error())
	} else {
		zap.S().Info("发布奖池池底数据", string(res))
	}
}

func PublishBingoPrizePool(list []*model.JackpotRecordModel) {
	res, err := json.Marshal(list)
	if err != nil {
		zap.S().Info("构建数据异常", err.Error())
	}
	err = logic.Nc.Publish(consts.GameJackpotPool, res)
	if err != nil {
		zap.S().Info("发布奖池池底数据异常", err.Error())
	} else {
		zap.S().Info("发布奖池池底数据", list)
	}
}
func Payout(transfer []*model.TransferRequestBody) {
	defer func() {
		if a2 := recover(); a2 != nil {
			zap.S().Error("Payment fail a2=", a2)
		}
	}()
	if len(transfer) > 0 {
		set := make(map[string]*model.TransferRequestBody)
		for _, body := range transfer {
			requestBody, ok := set[body.OrderId]
			if ok {
				requestBody.Amount = requestBody.Amount + body.Amount
			} else {
				set[body.OrderId] = body
			}
		}
		slice := make([]*model.TransferRequestBody, 0)
		for i := range set {
			slice = append(slice, set[i])
		}
		zap.S().Info("Payout len=", len(slice))
		now := time.Now()

		length := 100
		orderArr, err := helper.SplitArray[*model.TransferRequestBody](slice, length)
		if err != nil {
			zap.S().Error(err)
		}

		var (
			wg = sync.WaitGroup{}
		)

		wg.Add(len(orderArr))
		for _, orders := range orderArr {
			go func(orders []*model.TransferRequestBody) {
				defer wg.Done()
				//err = Stub.RpcWalletBatchPayout(orders)
				//if err == nil {
				//	return
				//} else {
				//	zap.S().Info("bingo mega 派彩失败 err", err.Error())
				//}
				marshal, err2 := json.Marshal(orders)
				if err2 != nil {
					zap.S().Error("Marshal fail err=", err.Error())
					return
				}
				util.PublishMsg(logic.PublishChan, consts.PayOutPay, marshal)
			}(orders)
		}
		wg.Wait()
		if time.Now().UnixMilli()-now.UnixMilli() > 1000 {
			zap.S().Info("派彩的时长 ", time.Since(now))
		}
	}
}
func FormatFloat(num float64, decimal int) string {
	// 默认乘1
	d := float64(1)
	if decimal > 0 {
		// 10的N次方
		d = math.Pow10(decimal)
	}
	// math.trunc作用就是返回浮点数的整数部分
	// 再除回去，小数点后无效的0也就不存在了
	return strconv.FormatFloat(math.Trunc(num*d)/d, 'f', -1, 64)
}

func PublishFinalResult(tmp []*PrizeModel) {
	if len(tmp) > 100 {
		// 没100份发送一次数据
		orderArr, err := helper.SplitArray[[]*PrizeModel](tmp, 100)
		if err != nil {
			zap.S().Error("SplitArray err=", err)
			return
		}
		for _, orders := range orderArr {
			go func(orders any) {
				res, _ := json.Marshal(orders)
				err := logic.Nc.Publish(consts.GameSettle, res)
				if err != nil {
					zap.S().Error("发生最终中奖结果数据报错", err.Error())
				}
			}(orders)
		}
	} else {
		res, _ := json.Marshal(tmp)
		err := logic.Nc.Publish(consts.GameSettle, res)
		if err != nil {
			zap.S().Error("发生最终中奖结果数据报错", err.Error())
		}
	}
}
