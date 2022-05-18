package businessdao

type Category string

var (
	CategoryPets       Category = "pets"
	CategoryAutomotive Category = "auto"
	CategoryEvents     Category = "events"
	CategoryBeauty     Category = "beauty"
	CategoryHome       Category = "home"
	CategoryHealth     Category = "health"
)

func (c Category) IsValid() bool {
	allCategories := []Category{CategoryPets, CategoryAutomotive, CategoryEvents, CategoryBeauty, CategoryHome, CategoryHealth}
	for _, category := range allCategories {
		if c == category {
			return true
		}
	}

	return false
}
