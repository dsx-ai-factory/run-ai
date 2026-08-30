package database

type ParametrizedQuery struct {
	Query  string
	Params []any
}

func NewParametrizedQuery() *ParametrizedQuery {
	return &ParametrizedQuery{
		Params: make([]any, 0),
	}
}

func (nvq *ParametrizedQuery) Add(query string, params ...any) *ParametrizedQuery {
	nvq.Query += " " + query
	nvq.Params = append(nvq.Params, params...)
	return nvq
}
