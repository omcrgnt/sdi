package sdi

// type depImpl struct{}

// func (t *depImpl) GetOne() string {
// 	return "some"
// }

// type dep interface {
// 	GetOne() string
// }

// type some struct {
// 	dep dep
// }

// func newSome() *some {
// 	return &some{}
// }

// func (t *some) Deps() []any {
// 	// Передаем указатель на интерфейс.
// 	// Это позволит reflect увидеть, что это именно тип dep.
// 	return []any{(*dep)(nil)}
// }

// func showType(source any, resource []any) {
// 	t := reflect.TypeOf(source)
// 	if t.Kind() == reflect.Ptr {
// 		t = t.Elem()
// 	}

// 	if t.Kind() != reflect.Interface {
// 		panic("not interface")
// 	}

// 	for _, r := range resource {
// 		resType := reflect.TypeOf(r)

// 		fmt.Println(r, t)

// 		if resType.Implements(t) {
// 			fmt.Printf("Ресурс %T реализует интерфейс %v\n", r, t)
// 		} else {
// 			fmt.Printf("Ресурс %T НЕ подходит\n", r)
// 		}
// 	}

// }

// func main() {
// 	var resource = []any{
// 		&depImpl{},
// 	}
// 	var s = newSome()
// 	for _, dep := range s.Deps() {
// 		showType(dep, resource)
// 	}
// }
