package invoice

type Repository interface {
	Save(*Invoice) error
	FindByID(string) (*Invoice, error)
	FindAll() ([]*Invoice, error)
	UpdateStatus(string, Status) error
}
