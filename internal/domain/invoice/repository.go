package invoice

type Repository interface {
	Save(*Invoice) error
	FindByID(string) (*Invoice, error)
	FindAll() ([]*Invoice, error)
	GetStatus(string) (uint8, error)
	UpdateStatus(string, Status) error
}
