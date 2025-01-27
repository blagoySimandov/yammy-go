test:
	go run main.go && diff test.yaml updated_test.yaml
