# Streaming API 예제

이 예제는 PocketBase Go Client의 스트리밍 기능을 보여줍니다.

## 기능

- **메모리 버퍼 스트리밍**: 응답을 메모리 버퍼로 스트리밍
- **파일 스트리밍**: 대용량 응답을 파일로 직접 스트리밍
- **진행률 추적**: 스트리밍 진행률을 실시간으로 추적
- **실시간 스트리밍**: 실시간 이벤트 스트리밍

## 실행 방법

```bash
# PocketBase 서버가 실행 중인지 확인 (http://127.0.0.1:8090)
cd examples/streaming_api
go run main.go
```

## 주요 특징

### WithResponseWriter 옵션

`WithResponseWriter` 옵션을 사용하면 응답 데이터를 메모리에 로드하지 않고 직접 `io.Writer`로 스트리밍할 수 있습니다:

```go
var buf bytes.Buffer
err := client.SendWithOptions(
    ctx,
    "GET",
    "/api/collections",
    nil,
    nil, // responseData는 nil이어야 함
    pb.WithResponseWriter(&buf),
)
```

### 진행률 추적

커스텀 Writer를 구현하여 스트리밍 진행률을 추적할 수 있습니다:

```go
type ProgressWriter struct {
    writer     io.Writer
    totalBytes int64
    onProgress func(bytes int64)
}
```

### 실시간 스트리밍

Realtime 서비스를 통해 실시간 이벤트를 스트리밍할 수 있습니다:

```go
unsubscribe, err := client.Realtime.Subscribe(
    ctx,
    []string{"*"}, // 모든 컬렉션
    func(event *pb.RealtimeEvent, err error) {
        // 실시간 이벤트 처리
    },
)
```

## 생성되는 파일

- `collections_stream.json`: 컬렉션 데이터가 스트리밍된 파일
- `progress_stream.json`: 진행률 추적과 함께 스트리밍된 파일

## 주의사항

1. `WithResponseWriter`와 `responseData` 파라미터는 함께 사용할 수 없습니다
2. 스트리밍 시에는 항상 context를 사용하여 타임아웃을 설정하세요
3. 대용량 파일 스트리밍 시에는 버퍼링된 Writer를 사용하는 것이 좋습니다