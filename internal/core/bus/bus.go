package bus

type Bus interface {
	Publish(topic string, event any) error
	Subscribe(topic string, handler func(event any)) error

	RegisterQuery(name string, handler func(req any) (any, error)) error
	Query(name string, req any) (any, error)
}
