package reflect

import "internal/reflectlite"

func VisibleFields(t Type) []StructField {
	return reflectlite.VisibleFields(toRawType(t))
}
