package payment

type Repository interface {
	SaveIfNotExist(*Payment) (bool, error)
	FindByIdempotencyKey(string) (*Payment, error)
	UpdateStatus(string, Status) error
}
