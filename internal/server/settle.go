package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"strings"
	"sync"
)

type SettleService struct {
	Periods map[string]*SettlePeriod
}

type RewardPoints []int32 //卡片上的中奖数字
type SettlePeriod struct {
	Cards         map[string]mapset.Set   // 记录卡片 对应订单 orderId-cardId
	Index         map[int32]string        // 记录位置,当前球的位置,key:
	CardsRank     map[string]uint32       // 卡片分数
	CardId2OderId map[string]string       // 卡片2订单 key:card id value:order id
	Count         int                     // 发球次数,开奖类型 44 -- bingo 49
	Lottery       []int32                 // 开奖结果 --> 当前这期所有开奖的数字
	CardI2Lottery map[string]RewardPoints // 卡片对应的开奖结果 key:卡片id  value:卡片中奖结构
}

func (s *SettleService) NatsSettle(m *model.NatsSettle) error {

	key, err := m.GetKey()
	if err != nil {
		return err
	}
	periodKey, err := m.GetPeriodKey()
	if err != nil {
		return err
	}
	period := s.GetPeriod(periodKey, m.Type)
	period.Lottery = append(period.Lottery, m.Number)

	cardKey, err := m.GetCardKey()
	if err != nil {
		return err
	}
	// 位置
	period.IndexHandle(key)
	// 获取卡片
	err = period.GetCards(key, cardKey, m.Number)
	if err != nil {
		return err
	}
	if period.Cards[key] == nil {
		return nil
	}
	// todo 结算逻辑
	// 计算交集
	period.HandleMixe(key)
	// 匹配
	period.Patterns()
	// 排名
	period.RankSort()
	source := make(map[string]consts.MatchesType)
	source["A"] = consts.A
	source["B"] = consts.B
	source["C"] = consts.C

	period.Contain(source, model.Ball{Number: 10, Index: 1})
	// 判断是否完成
	if len(period.CardsRank) == period.Count && len(period.Index) == period.Count {
		period.Clean()
	}
	return nil
}

// GetCards 获取卡片
func (s *SettlePeriod) GetCards(key, cardKey string, number int32) error {

	result, err := logic.Rdb.Keys(context.TODO(), cardKey).Result()
	if err != nil {
		return err
	}
	if len(result) < 1 {
		return nil
	}
	pipeline := logic.Rdb.Pipeline()
	for _, s2 := range result {
		pipeline.SMembers(context.TODO(), s2)
	}
	cmders, err := pipeline.Exec(context.TODO())
	if err != nil {
		return err
	}

	for _, cmder := range cmders {
		res, err := cmder.(*redis.StringSliceCmd).Result()
		if err != nil {
			// error
			return err
		}
		for _, value := range res {
			split := strings.Split(value, "-")
			if len(split) != 2 {
				return errors.New("bad data")
			}
			OderId := split[0]
			CardId := split[1]
			s.CardId2OderId[CardId] = OderId

			if set, ok := s.Cards[key]; !ok {
				s.Cards[key] = mapset.NewSet(value)
			} else {
				set.Add(value)
			}
			// 卡片2结果
			if cl, ok := s.CardI2Lottery[CardId]; !ok {
				s.CardI2Lottery[CardId] = make([]int32, 0)
				s.CardI2Lottery[CardId] = append(s.CardI2Lottery[CardId], number)
			} else {
				cl = append(cl, number)
			}
			// 分数
			if set, ok := s.CardsRank[CardId]; !ok {
				s.CardsRank[CardId] = 1
			} else {
				s.CardsRank[CardId] = set + 1
			}
		}
	}
	return nil
}

// HandleMixe 计算交集
func (s *SettlePeriod) HandleMixe(key string) {
	set := s.Cards[key]
	if set != nil {
		slice := set.ToSlice()
		if len(slice) > 0 {
			//s.Mixe = util.IntersectArray(s.Mixe, slice)
		}
	}
}

// Patterns 匹配图案
func (s *SettlePeriod) Patterns() {
	//patterns := consts.Patterns
	//todo
}

// HandleCardILottery 处理卡片对应结果
func (s *SettlePeriod) HandleCardILottery() {

}

func (s *SettleService) GetPeriod(key string, marketType consts.MarketType) *SettlePeriod {
	if s.Periods == nil {
		s.Periods = make(map[string]*SettlePeriod, 0)
		s.Periods[key] = NewPeriod(marketType)
		return s.Periods[key]
	} else {
		if period, ok := s.Periods[key]; ok {
			return period
		} else {
			s.Periods[key] = NewPeriod(marketType)
			return s.Periods[key]
		}
	}
}

// 幂等性
func (s *SettleService) GetMsg(key string) (bool, error) {
	_, err := logic.Rdb.Get(context.TODO(), key).Result()
	if err != nil {
		if err != redis.Nil {
			zap.S().Error("sceneMsg redis get fail err=", err.Error())
			return false, err
		} else {
			return false, nil
		}
	}
	return true, nil
}

func SetMsg(key string) error {
	_, err := logic.Rdb.Set(context.TODO(), key, 1, consts.MsgIdempotencyTime).Result()
	if err != nil {
		zap.S().Error("sceneMsg redis get fail err=", err.Error())
		return err
	}
	return nil
}

// IndexHandle 位置
func (s *SettlePeriod) IndexHandle(key string) {
	if s.Index == nil {
		s.Index = make(map[int32]string, 0)
		s.Index[0] = key
	} else if len(s.Index) == s.Count {
		if len(s.Cards) == s.Count {
			//
		}
	} else if len(s.Index) == 0 {
		s.Index[0] = key
	} else {
		i := len(s.Index)
		s.Index[int32(i)] = key
	}
}

// RankSort 排名
func (s *SettlePeriod) RankSort() {
	// todo 获取排名需要的数据
	//
	// s.Index
	//logic.Rdb0.ZAdd()
	//写入到redis

}
func (s *SettlePeriod) Contain1(targets map[string]consts.MatchesType, ball model.Ball) map[string]consts.MatchesType {
	// 开奖结果 构造卡片id 数据集
	var t = make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		t = append(t, fmt.Sprintf("%d-%d", ball.Number, i))
	}
	//zap.S().Info("所有的中奖结果的key数组", t)
	group := sync.WaitGroup{}
	var prize = make(map[string]consts.MatchesType)
	//遍历所有的中奖类型 target 单一的中奖结构
	for index, target := range targets {
		group.Add(1)
		go func(tag string, t consts.MatchesType) {
			//匹配每个中奖类型和当前开奖的数据源做匹配
			var t1 []string
			for _, i := range t {
				if t.MatchesTypeContainsValue(i) {
					t1 = append(t1, i)
				}
			}
			//zap.S().Info("t1", t1)
			res := logic.PipelineGetValues1(t1)
			//zap.S().Info("res", res)
			res = util.SliceRemoveDuplicates(res)
			prize[tag] = res
			defer group.Done()
		}(index, target)
	}
	group.Wait()
	//zap.S().Info("prize --->>>", prize)
	return prize
}

// Contain 是否包含中奖内容 target-->> 各种中奖类型的目标key的集合，source-->> 现在已经出奖的数据
func (s *SettlePeriod) Contain(targets map[string]consts.MatchesType, ball model.Ball) map[string]consts.MatchesType {
	// todo 获取排名需要的数据
	//
	// s.Index
	//logic.Rdb0.ZAdd()
	//zap.S().Info("target --->>>", targets, source)

	// 开奖结果 构造卡片id 数据集
	var t = make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		t = append(t, fmt.Sprintf("bingo:card:%d-%d", ball.Number, i))
	}
	zap.S().Info("所有的中奖结果的key数组", t)

	group := sync.WaitGroup{}
	var prize = make(map[string]consts.MatchesType)
	//遍历所有的中奖类型 target 单一的中奖结构
	for i, target := range targets {
		group.Add(1)
		go func(tag string, t consts.MatchesType) {
			//匹配每个中奖类型和当前开奖的数据源做匹配
			var t1 []string
			for _, i := range t {
				if t.MatchesTypeContainsValue(i) {
					t1 = append(t1, i)
				}
			}
			//zap.S().Info("t1", t1)
			res := logic.PipelineGetValues(t1)
			//zap.S().Info("res", res)
			res = util.SliceRemoveDuplicates(res)
			prize[tag] = res
			defer group.Done()
		}(i, target)
	}
	group.Wait()
	zap.S().Info("prize --->>>", prize)
	return prize
}

// MsgHandle 结果处理
func (s *SettleService) MsgHandle(msg *nats.Msg) {
	var m model.NatsSettle
	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		return
	}
	key, err := m.GetMsgKey()
	if err != nil {
		return
	}
	getMsg, err := s.GetMsg(key)
	if err != nil {
		return
	}
	if getMsg {
		return
	}
	// settle
	err = s.NatsSettle(&m)
	if err != nil {
		return
	}
	err = SetMsg(key)
	if err != nil {
		return
	}
}

// Clean 一期完成清除
func (s *SettlePeriod) Clean() {
	s.Index = make(map[int32]string, 0)
	s.CardsRank = make(map[string]uint32, 0)
	s.Cards = make(map[string]mapset.Set, 0)
	s.CardId2OderId = make(map[string]string, 0)
	s.Lottery = make([]int32, 0)
	s.CardI2Lottery = make(map[string]RewardPoints, 0)
}

func NewPeriod(MarketType consts.MarketType) *SettlePeriod {
	var s SettlePeriod
	s.Index = make(map[int32]string, 0)
	s.CardsRank = make(map[string]uint32, 0)
	s.Cards = make(map[string]mapset.Set, 0)
	s.CardId2OderId = make(map[string]string, 0)
	s.Lottery = make([]int32, 0)
	s.CardI2Lottery = make(map[string]RewardPoints, 0)
	//
	if MarketType == consts.MarketTypeMega {
		s.Count = int(consts.MarketCountMega)
	} else if MarketType == consts.MarketTypeRush {
		s.Count = int(consts.MarketCountRush)
	}
	return &s
}
