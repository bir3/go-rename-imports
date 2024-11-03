
test:
	go generate
	go test

debug:
	# -v => test files will be written into local 'tmp' folder
	# for inspection
	go generate
	go test -v -failfast
