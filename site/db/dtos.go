package db

type HasDTO interface {
	ToDTO() interface{}
}

func AsDTOs(objs []HasDTO) []interface{} {
	dtos := make([]interface{}, len(objs))
	for i := range objs {
		dtos[i] = objs[i].ToDTO()
	}

	return dtos
}
