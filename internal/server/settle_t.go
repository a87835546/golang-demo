package service

import (
	"context"
	"sync"
)

type Settle interface {
	GameEvent(event *model.GameRoundEvent)
	InitData(ctx context.Context, periodNum string) error
	ReadingPeriod(ctx context.Context) error                 // 获取期号
	ReadingOdd(ctx context.Context, gameType []string) error // 使用游戏类型获取赔率
	ReadOrders(ctx context.Context, periodNum string)        // 获取订单数据
	ReadingPrizePool()
	UpdatePrizePool()
}

type SettleMerchant interface {
	Settle(num int) // 结算
	Rank()
	FinalSettle()
}

var (
	settleServices = make(map[string]Settle)
	settleRWMutex  sync.RWMutex
)

func Register(key string, value Settle) {
	settleRWMutex.Lock()
	defer settleRWMutex.Unlock()
	settleServices[key] = value
	return
}

func GetHandlers(key string) (value Settle, ok bool) {
	settleRWMutex.RLock()
	defer settleRWMutex.RUnlock()
	value, ok = settleServices[key]
	return
}
