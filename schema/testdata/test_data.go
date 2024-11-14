package testdata

func GetTestStructWithPrivateFieldPresent() TestStructWithPrivateField {
	return TestStructWithPrivateField{
		Field1: "test",
		field2: "private",
	}
}

type TestStructWithPrivateField struct {
	Field1 string `json:"field1"`
	field2 string
}
