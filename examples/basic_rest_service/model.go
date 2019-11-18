package basic_rest_service

type Person struct {
	ID        int64
	FirstName string
	LastName  string
}

var people = []*Person{
	{466354, "Joe", "Dassin"},
	{34733897, "Tigus", "Swinswin"},
	{37543784973982, "Jane", "Doe"},
}

func findPersonByID(id int64) *Person {
	for _, p := range people {
		if p.ID == id {
			return p
		}
	}
	return nil
}
