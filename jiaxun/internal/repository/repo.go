package repository

type Repository[T any] interface {
	Create(obj *T) error
	GetByID(id uint) (*T, error)
	Update(obj *T) error
	Delete(id uint) error
	GetAll() ([]T, error)
	List(page, pageSize int) ([]T, int64, error)
}
