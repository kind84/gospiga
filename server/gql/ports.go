package gql

// ** TODO: fix missing mockgen in CI //go:generate mockgen -source ports.go -destination portsmock_test.go -package gql_test

type App interface{}
