# 使用带CGO的Go镜像
FROM golang:1.20-bullseye as builder

# 更新包列表并安装libvips
RUN apt-get update && apt-get install -y libvips-dev

# 设置环境变量，启用CGO
ENV CGO_ENABLED=1

# 在容器中设置工作目录
WORKDIR /app

# 复制go模块文件
COPY go.mod go.sum ./

# 下载所有依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN go build -o main .

# 运行阶段，使用轻量级的scratch镜像
FROM debian:bullseye-slim

# 更新包列表并安装libvips
RUN apt-get update && apt-get install -y libvips

# 从builder镜像中复制执行文件
COPY --from=builder /app/main /app/main

# 指定应用程序使用的端口
EXPOSE 8800

# 运行程序
CMD ["/app/main"]
