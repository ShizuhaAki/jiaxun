package repository

import "gorm.io/gorm"

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) Create(obj *T) error {
	return r.db.Create(obj).Error
}

func (r *BaseRepository[T]) GetByID(id uint) (*T, error) {
	var obj T
	err := r.db.First(&obj, id).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *BaseRepository[T]) Update(obj *T) error {
	return r.db.Save(obj).Error
}

func (r *BaseRepository[T]) Delete(id uint) error {
	var obj T
	return r.db.Delete(&obj, id).Error
}

func (r *BaseRepository[T]) GetAll() ([]T, error) {
	var objs []T
	err := r.db.Find(&objs).Error
	if err != nil {
		return nil, err
	}
	return objs, nil
}

func (r *BaseRepository[T]) List(page, pageSize int) ([]T, int64, error) {
	var objs []T
	var total int64

	if err := r.db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := r.db.Offset(offset).Limit(pageSize).Find(&objs).Error; err != nil {
		return nil, 0, err
	}

	return objs, total, nil
}
