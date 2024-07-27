FROM golang:1.20 as build

WORKDIR /app

COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY main.go .
RUN go build -tags lambda.norpc -o main main.go

FROM  public.ecr.aws/lambda/provided:al2023

COPY --from=build /app/main ./main

ENTRYPOINT  ["./main"]
