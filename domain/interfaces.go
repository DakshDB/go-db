package domain

// GoDBUsecase is the interface for the GoDB usecase
type GoDBUsecase interface {
	HealthCheck() (string, error)
	Execute(databaseName string, query string) (interface{}, error)
	Save(databaseName string) error
	Load(database string) error
}

// GoDBRepository is the interface for the GoDB repository
type GoDBRepository interface {
}
