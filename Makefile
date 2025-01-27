test:
	go run main.go types.go && diff test.yaml updated_test.yaml
