## grpc demo

学习目的
1. 看源码，看看 grpc 内部是否有使用连接池: 结论是没有！
2. FieldMask 的使用。作用：达到实现 graphql 的作用
    - https://mp.weixin.qq.com/s/L7He7M4JWi84z1emuokjbQ
    - https://mp.weixin.qq.com/s/uRuejsJN37hdnCN4LLeBKQ
3. grpc 错误处理
```go
// google.golang.org/grpc/stream.go: client toRPCErr
err = recv(a.p, cs.codec, a.s, a.dc, m, *cs.callInfo.maxReceiveMessageSize, nil, a.decomp)
if err == nil {
    return toRPCErr(errors.New("grpc: client streaming protocol violation: get <nil>, want <EOF>"))
}

// use
return nil, status.Error(codes.NotFound, "some description")
```


