# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
ok  	campusqa/cmd/campusqa	0.005s
ok  	campusqa/internal/catalog	0.003s
--- FAIL: Test1054BusinessRegression (0.00s)
    flow_test.go:78: resubmission category changed from examination to enrollment
FAIL
FAIL	campusqa/internal/flow018	0.029s
ok  	campusqa/internal/model	0.001s
ok  	campusqa/internal/report	0.004s
ok  	campusqa/internal/review	0.006s
ok  	campusqa/internal/search	0.005s
ok  	campusqa/internal/store	0.010s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/campusqa): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/campusqa): exit `0`
