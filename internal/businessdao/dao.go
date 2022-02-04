package businessdao

import "context"

type Business struct {
	ID string `json:"id"`
}

type Dao struct {
	// TODO: add real database client (Firestore)
}

func NewDao() *Dao {
	return &Dao{}
}

func (dao *Dao) GetBusiness(ctx context.Context, id string) (*Business, error) {
	business := Business{
		ID: id,
	}

	return &business, nil
}
