package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"dragonfly-monitor/internal/common"
	"dragonfly-monitor/internal/domain/entity"
	"dragonfly-monitor/internal/infrastructure/persistence"
	"dragonfly-monitor/internal/types"

	"github.com/pwh19920920/snowflake"
)

// MonitorVolatilityDayApplication 波动日管理
type MonitorVolatilityDayApplication struct {
	sequence   *snowflake.Node
	repository *persistence.Repository
}

// NewMonitorVolatilityDayApplication 创建波动日应用服务
func NewMonitorVolatilityDayApplication(sequence *snowflake.Node, repository *persistence.Repository) MonitorVolatilityDayApplication {
	return MonitorVolatilityDayApplication{sequence: sequence, repository: repository}
}

// SelectAll 查询全部
func (app *MonitorVolatilityDayApplication) SelectAll(ctx context.Context) ([]entity.MonitorVolatilityDay, error) {
	return app.repository.MonitorVolatilityDayRepository.SelectAll()
}

// IsSpecialDay 判断当前时间是否落在任一波动日区间内
func (app *MonitorVolatilityDayApplication) IsSpecialDay(ctx context.Context, t time.Time) (bool, error) {
	days, err := app.SelectAll(ctx)
	if err != nil {
		return false, err
	}
	return MatchVolatilityDay(days, t) != nil, nil
}

// Hit 返回 t 命中的第一个波动日，未命中返回 nil
func (app *MonitorVolatilityDayApplication) Hit(ctx context.Context, t time.Time) (*entity.MonitorVolatilityDay, error) {
	days, err := app.SelectAll(ctx)
	if err != nil {
		return nil, err
	}
	return MatchVolatilityDay(days, t), nil
}

// MatchVolatilityDay 在已加载的波动日列表中查找 t 命中的波动日，nil 表示未命中。
// 纯函数，不依赖应用服务状态，故为包级函数而非方法。
func MatchVolatilityDay(days []entity.MonitorVolatilityDay, t time.Time) *entity.MonitorVolatilityDay {
	for i := range days {
		if !t.Before(days[i].StartTime.Time) && !t.After(days[i].EndTime.Time) {
			return &days[i]
		}
	}
	return nil
}

// checkVolatilityDayConflict 校验候选区间与已有列表是否存在「时间重叠但类型不同」的冲突。
// excludeId 用于编辑场景排除自身；同类型重叠不视为冲突。纯函数，便于复用与测试。
//
// 实现：候选 + 已有（排除自身）按开始时间排序后单次扫描，维护「各类型已见过的最大结束时间」，
// 当前区间开始时间早于对侧类型的最大结束时间即冲突 —— 避免两两 O(n²) 比较。
func checkVolatilityDayConflict(candidates, existing []entity.MonitorVolatilityDay, excludeId int64) error {
	items := make([]entity.MonitorVolatilityDay, 0, len(candidates)+len(existing))
	items = append(items, candidates...)
	for _, e := range existing {
		if e.Id != excludeId {
			items = append(items, e)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		si, sj := items[i].StartTime.Time, items[j].StartTime.Time
		if !si.Equal(sj) {
			return si.Before(sj)
		}
		return items[i].EndTime.Time.Before(items[j].EndTime.Time)
	})

	// 各类型已见过区间的最大结束时间（附区间，用于报错）
	type span struct {
		end  time.Time
		item entity.MonitorVolatilityDay
	}
	maxEnd := make(map[entity.VolatilityDayType]span)
	for _, cur := range items {
		opp := oppositeVolatilityType(cur.Type)
		if s, ok := maxEnd[opp]; ok && !cur.StartTime.Time.After(s.end) {
			return fmt.Errorf("区间[%s](%s ~ %s)与区间[%s](%s ~ %s)时间重叠，不能同时配置为高峰和低谷",
				cur.Name, cur.StartTime.Time.Format("2006-01-02 15:04:05"), cur.EndTime.Time.Format("2006-01-02 15:04:05"),
				s.item.Name, s.item.StartTime.Time.Format("2006-01-02 15:04:05"), s.item.EndTime.Time.Format("2006-01-02 15:04:05"))
		}
		if s, ok := maxEnd[cur.Type]; !ok || cur.EndTime.Time.After(s.end) {
			maxEnd[cur.Type] = span{end: cur.EndTime.Time, item: cur}
		}
	}
	return nil
}

// oppositeVolatilityType 对侧波动日类型（高峰↔低谷）
func oppositeVolatilityType(t entity.VolatilityDayType) entity.VolatilityDayType {
	if t == entity.VolatilityDayTypeTrough {
		return entity.VolatilityDayTypePeak
	}
	return entity.VolatilityDayTypeTrough
}

// BatchCreate 批量添加
func (app *MonitorVolatilityDayApplication) BatchCreate(ctx context.Context, req *types.MonitorVolatilityDayBatchCreateRequest) error {
	if err := req.ValidateForCreate(); err != nil {
		return err
	}
	days := make([]entity.MonitorVolatilityDay, 0, len(req.Items))
	for i := range req.Items {
		item := req.Items[i]
		item.BaseEntity = common.BaseEntity{Id: app.sequence.Generate().Int64()}
		item.Name = req.Name
		item.Type = req.Type
		days = append(days, item)
	}
	// 冲突校验：新增区间之间、以及与库中已有区间，不能出现时间重叠但类型不同（一高峰一低谷）
	existing, err := app.repository.MonitorVolatilityDayRepository.SelectAll()
	if err != nil {
		return err
	}
	if err := checkVolatilityDayConflict(days, existing, 0); err != nil {
		return err
	}
	return app.repository.MonitorVolatilityDayRepository.BatchSave(days)
}

// Modify 修改
func (app *MonitorVolatilityDayApplication) Modify(ctx context.Context, day *entity.MonitorVolatilityDay) error {
	if day.Id == 0 {
		return errors.New("id 不能为空")
	}
	// 冲突校验：排除自身后，与库中其它区间不能出现时间重叠但类型不同
	existing, err := app.repository.MonitorVolatilityDayRepository.SelectAll()
	if err != nil {
		return err
	}
	if err := checkVolatilityDayConflict([]entity.MonitorVolatilityDay{*day}, existing, day.Id); err != nil {
		return err
	}
	return app.repository.MonitorVolatilityDayRepository.Modify(day.Id, day)
}

// Delete 删除
func (app *MonitorVolatilityDayApplication) Delete(ctx context.Context, id int64) error {
	return app.repository.MonitorVolatilityDayRepository.Delete(id)
}
