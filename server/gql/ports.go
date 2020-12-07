package gql

//go:generate mockgen -source ports.go -destination portsmock_test.go -package gql_test

type App interface{}
