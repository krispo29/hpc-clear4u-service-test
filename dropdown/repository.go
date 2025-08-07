package dropdown

type Repository interface {
	// For future database operations if needed
	// Currently using static data, but interface is ready for expansion
}

type repository struct {
	// db connection would go here if needed
}

func NewRepository() Repository {
	return &repository{}
}
