package model

type Filter struct {

	/* 过滤条件的名称  */
	Name string `json:"name"`

	/* 过滤条件的操作符，默认eq (Optional) */
	Operator *string `json:"operator"`

	/* 过滤条件的值  */
	Values []string `json:"values"`
}
