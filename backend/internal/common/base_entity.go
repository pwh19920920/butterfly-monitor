package common

type DeleteStatus int32

const (
	DeletedFalse DeleteStatus = 0
	DeletedTrue  DeleteStatus = 1
)

type BaseEntity struct {
	Id        int64        `json:"id,string" gorm:"primary_key"`
	CreatedAt *LocalTime   `json:"createdAt" gorm:"index;column:created_at"`
	UpdatedAt *LocalTime   `json:"updatedAt" gorm:"column:updated_at"`
	Deleted   DeleteStatus `json:"deleted" gorm:"column:deleted"` // 删除标记
}
