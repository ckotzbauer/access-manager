
.PHONY: unit-test
unit-test:
    cd pkg/reconciler && \
    go test

e2e-test:
    cd e2e && \
    bash test.sh
