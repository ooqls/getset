package keydb

const (
	Query_password = "password"
	Query_KeyType  = "key_type"
)

type Query struct {
	key string
	val string
}

func WithPasswordHash(hash string) Query {
	return Query{
		key: Query_password,
		val: hash,
	}
}

func WithKeyType(key string) Query {
	return Query{
		key: "key_type",
		val: key,
	}
}

