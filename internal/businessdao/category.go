package businessdao

type Category string

var (
	CategoryPets       Category = "pets"
	CategoryAutomotive Category = "automotive"
	CategoryEvents     Category = "events"
	CategoryBeauty     Category = "beauty"
)

func (c Category) IsValid() bool {
	allCategories := []Category{CategoryPets, CategoryAutomotive, CategoryEvents, CategoryBeauty}
	for _, category := range allCategories {
		if c == category {
			return true
		}
	}

	return false
}
