# 多阶段构建 Dockerfile for GreenSoulAI

# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o greensoulai ./cmd/greensoulai

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates 和 tzdata
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S greensoulai && \
    adduser -u 1001 -S greensoulai -G greensoulai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/greensoulai .

# 创建必要的目录
RUN mkdir -p /app/data /app/logs && \
    chown -R greensoulai:greensoulai /app

# 切换到非root用户
USER greensoulai

# 暴露端口（如果需要）
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./greensoulai version || exit 1

# 运行应用程序
ENTRYPOINT ["./greensoulai"]
CMD ["run"]
